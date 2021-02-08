package main

/*
This package implements the following HTTP endpoints and behaviour:

	GET / no cookie -> set cookie, redirect /session
	GET /session no cookie -> error
	GET /session cookie -> redirect /
	GET / with cookie, provisioning not started -> button to /provision
	GET / with cookie, provisioning requested -> show status
	POST /provision with cookie, provisioning not started -> provision, set cookie requested, redirect /
	any other request to /provision -> redirect to /

State is maintained largely in-memory - if the instance dies everything is lost.
If a request is made with a machine ID we don't have a record of, its details are fetched from GCP.
*/

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	insecureRand "math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/gorilla/sessions"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	_ "golang.org/x/oauth2/google" // for application default creds
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

type Status struct {
	ID             string // session ID
	Name           string // who requested this VM
	Requested      bool
	ProvisionStart time.Time
	Provisioned    bool
	Deleted        bool
	VMName         string
	IP             string
	Password       string
	Error          string
	lock           *sync.Mutex
}

type Runtime struct {
	lock                   *sync.Mutex
	vms                    []string
	sessions               map[string]*Status
	cs                     *sessions.CookieStore
	maxVMs                 int
	compute                *compute.Service
	zone                   string
	sourceInstanceTemplate string
	projectID              string
}

const sessionName = "id"
const username = "fuzzer"
const sshPort = 2222
const instanceNamePrefix = "fuzz-training-"
const instanceNameSuffixLen = 6
const createError = "Failed to create instance. Session ID: "

func randomString(n int) string {
	var set = []rune("abcdefghkmnopqrstuvwxyz0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = set[insecureRand.Intn(len(set))] // nolint: gosec // this doesn't need to be unpredictable
	}
	return string(b)
}

func writeHtml(w io.Writer, main string) {
	templ := template.Must(template.New("html").Parse(`
		<head>
			<link rel="stylesheet" href="/mvp.css">
			<title>Fuzz Training</title>
		</head>
		<body>
			<main>
				<h1>Fuzz Training</h1>
				{{.}}
			</main>
		</body>
		`))

	err := templ.Execute(w, main)
	if err != nil {
		log.Error().Err(err).Msg("Writing response")
	}
}

// Get the status associated with this session ID. On error, set an http.Error and return false
// If the cookie includes a VM but the session doesn't exist; try and populate the session with VM details retrieved from the GCP API
func (rt *Runtime) getStatus(session *sessions.Session, w http.ResponseWriter) (*Status, bool) {
	val, ok := session.Values["id"]
	if !ok {
		log.Warn().Msgf("id value not in session")
		http.Error(w, "failed to parse cookie; please enable cookies", http.StatusBadRequest)
		return nil, false
	}
	id, ok := val.(string)
	if !ok {
		log.Warn().Msgf("id value in session was not the correct type. got '%v'", val)
		http.Error(w, "failed to parse cookie; please reset your cookies and refresh", http.StatusBadRequest)
		return nil, false
	}

	vm := ""
	val, ok = session.Values["vm"]
	if ok {
		vm, ok = val.(string)
		if !ok {
			log.Warn().Msgf("vm value in session was not the correct type. got '%v'", val)
			http.Error(w, "failed to parse cookie; please reset your cookies and refresh", http.StatusBadRequest)
			return nil, false
		}
	}

	status, ok := rt.sessions[id]
	if !ok {
		if vm == "" {
			log.Warn().Str("id", id).Msg("ID that isn't in store found; user's session has been reset")
			deleteCookie(w)
			http.Error(w, "cookie error - your session has been reset - please refresh", http.StatusInternalServerError)
			return nil, false
		}

		// there's no guarantee this VM exists, but if it doesn't the user will just get an error, which is what we want
		status = &Status{ID: id, VMName: vm, Requested: true, ProvisionStart: time.Now(), lock: &sync.Mutex{}}
		rt.lock.Lock()
		rt.sessions[id] = status
		rt.lock.Unlock()
	}
	return status, true
}

func deleteCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: "id", MaxAge: -1})
}

func (rt *Runtime) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		log.Debug().Str("method", r.Method).Msg("unexpected method for /")
		http.Error(w, "Unsupported method", 400)
		return
	}
	if (r.URL.Path != "/") && (r.URL.Path != "/index.html") {
		log.Debug().Str("path", r.URL.Path).Msg("unhandled path")
		http.Error(w, "", http.StatusNotFound)
		return
	}

	session, err := rt.cs.Get(r, sessionName)
	if err != nil {
		log.Error().Err(err).Msg("reading cookie")
		http.Error(w, "error reading cookie", http.StatusInternalServerError)
		return
	}

	if session.IsNew {
		b := make([]byte, 15)
		_, err = rand.Read(b)
		if err != nil {
			log.Error().Err(err).Msg("Failed to generate session ID")
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		id := base64.StdEncoding.EncodeToString(b)

		rt.lock.Lock()
		rt.sessions[id] = &Status{ID: id, lock: &sync.Mutex{}}
		rt.lock.Unlock()

		session.Values["id"] = id
		err = session.Save(r, w)
		if err != nil {
			log.Error().Err(err).Msg("whilst saving session")
			http.Error(w, "server error whilst creating session", http.StatusInternalServerError)
			return
		}

		log.Debug().Msg("New session, redirecting to /session to validate")
		http.Redirect(w, r, "/session", http.StatusFound)
		return
	}

	status, ok := rt.getStatus(session, w)
	if !ok {
		return
	}

	if !status.Requested {
		log.Debug().Msg("serving request form")
		writeHtml(w, `
		<form action="/provision" method="post">
			<header>
				<h2>Provision VM</h2>
			</header>
			<label for="input1">Your name (to aid the facilitator):</label>
			<input type="text" id="name" name="name" size="20">
			<button type="submit">Provision</button>
		</form>
		`)
		return
	} else if status.Error != "" {
		writeHtml(w, `<p>There was an error whilst provisioning your machine!</p><p><pre>`+status.Error+`</pre></p>
		 <p>Please contact your facilitator. If they ask you to try again, please <a href="/reset">reset your session</a>.</p>`)
		return
	} else if !status.Provisioned {
		rt.updateStatus(status)
		// this may change status.Provisioned to true
	}
	if !status.Provisioned {
		log.Debug().Msg("serving provisioning page")
		writeHtml(w, `
			<p>Your machine is being provisioned! This page will refresh every 5 seconds; your machine should be available in 2-3 minutes.</p>
			<p>It's been `+time.Since(status.ProvisionStart).Round(time.Second).String()+` so far.</p>
			<p>Current details:</p>`+machineDetails(status)+`
			<script>setTimeout("location.reload(true);",5000);</script>
		`)
		return
	} else {
		log.Debug().Msg("serving success page")
		msg := "<p>Success! Your machine details: </p>" + machineDetails(status)
		msg += "<p><code>ssh " + username + "@" + status.IP + " -p " + fmt.Sprint(sshPort) + "</code></p>"
		msg += `<p>Please <a href="/delete">delete</a> your machine when you're finished.<p>`
		writeHtml(w, msg)
		return
	}
}

// Return an HTML table rendering of the student's VM details
func machineDetails(s *Status) string {
	msg := "<p><table>"
	msg += "<tr><td>IP Address</td> <td>" + s.IP + "</td></tr>"
	msg += "<tr><td>Username</td> <td>" + username + "</td></tr>"
	msg += "<tr><td>Password</td> <td>" + s.Password + "</td></tr>"
	msg += "<tr><td>Port</td> <td>" + fmt.Sprint(sshPort) + "</td></tr>"
	msg += "<tr><td>ID</td> <td>" + s.VMName + "</td></tr>"
	msg += "</table></p>"
	return msg
}

// Verify that the client has a session with us; if so, initialize it and return to /, otherwise give an error
func (rt *Runtime) handleSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		log.Debug().Str("method", r.Method).Msg("unexpected method for /session")
		http.Error(w, "Unsupported method", 400)
		return
	}

	session, err := rt.cs.Get(r, sessionName)
	if err != nil {
		log.Error().Err(err).Msg("reading cookie")
		http.Error(w, "error reading cookie", http.StatusInternalServerError)
		return
	}
	if session.IsNew {
		log.Warn().Msg("Request to /session with no cookies")
		msg := "Sorry it looks like your browser isn't configured to accept cookies. Please enable them and try again."
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	// all is well - we have verified that cookie sessions work
	http.Redirect(w, r, "/", http.StatusFound)
}

func (rt *Runtime) handleReset(w http.ResponseWriter, r *http.Request) {
	// delete the cookie so they can try again
	deleteCookie(w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (rt *Runtime) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		session, err := rt.cs.Get(r, sessionName)
		if err != nil {
			log.Error().Err(err).Msg("reading cookie")
			http.Error(w, "error reading cookie", http.StatusInternalServerError)
			return
		}
		status, ok := rt.getStatus(session, w)
		if !ok {
			writeHtml(w, "<p>Couldn't find session</p>")
			return
		}

		req := rt.compute.Instances.Delete(rt.projectID, rt.zone, status.VMName)
		_, err = req.Do()
		if err != nil {
			log.Error().Err(err).Str("id", status.ID).Str("vm-name", status.VMName).Msg("deleting VM")
			status.Error = "Failed to delete VM " + status.VMName
			return
		}

		status.lock.Lock()
		status.Deleted = true
		status.lock.Unlock()

		http.Redirect(w, r, "/reset", http.StatusFound)
		return
	}
	writeHtml(w, `
		<form action="/delete" method="post">
			<header>
				<h2>Delete VM?</h2>
				<p>Are you sure you want to delete your VM? You will lose any work that you haven't copied off it already.</p>
			</header>
			<button type="submit">Delete</button>
			<input type="button" name="cancel" value="cancel" onClick="window.location.href='/';" />         
		</form>
	
	`)
}

// start provisioning if we haven't already, and redirect to /
func (rt *Runtime) handleProvision(w http.ResponseWriter, r *http.Request) {
	session, err := rt.cs.Get(r, sessionName)
	if err != nil {
		log.Error().Err(err).Msg("reading cookie")
		http.Error(w, "error reading cookie", http.StatusInternalServerError)
		return
	}
	status, ok := rt.getStatus(session, w)
	if !ok {
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if len(name) == 0 {
		http.Error(w, "you must enter a name", http.StatusBadRequest)
		return
	}

	if status.Requested {
		log.Warn().Str("id", status.ID).Str("name", name).Msg("/provision visited by user that is already provisioning")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	if rt.provision(status.ID, name) {
		session.Values["vm"] = status.VMName
		err = session.Save(r, w)
		if err != nil {
			log.Error().Err(err).Str("id", status.ID).Msg("whilst saving VMName to session")
		}
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

// Start VM creation. If a VM was created, returns true
func (rt *Runtime) provision(id string, name string) bool {
	log.Info().Str("id", id).Msg("Started provisioning")

	status := rt.sessions[id]
	status.lock.Lock()
	defer status.lock.Unlock()

	if status.Requested {
		log.Warn().Str("id", id).Msg("Tried to provision an already-provisioning request")
		return false
	}

	shortName := name
	limit := 30
	if len(shortName) > limit {
		shortName = shortName[:limit]
	}
	status.Name = shortName
	status.Requested = true
	status.ProvisionStart = time.Now()

	if rt.maxVMs > 0 && len(rt.vms) >= rt.maxVMs {
		log.Warn().Int("num-vms", len(rt.vms)).Int("limit", rt.maxVMs).Msg("Hit VM limit")
		status.Error = "VM limit hit, cannot provision additional VMs."
		return false
	}

	status.VMName = instanceNamePrefix + randomString(instanceNameSuffixLen)

	ins := rt.compute.Instances.Insert(rt.projectID, rt.zone, &compute.Instance{
		Name:            status.VMName,
		ServiceAccounts: []*compute.ServiceAccount{},
		Zone:            rt.zone,
		// Metadata: can't alter this, as we have to use the template's metadata set via gcloud. GCP don't have a REST api method for running containers
	})
	ins.SourceInstanceTemplate(rt.sourceInstanceTemplate)
	_, err := ins.Do()
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("creating instance")
		status.Error = createError + id
		return false
	}

	rt.lock.Lock()
	rt.vms = append(rt.vms, status.VMName)
	rt.lock.Unlock()

	return true
}

// Update status with the details of the provisioned VM
func (rt *Runtime) updateStatus(status *Status) {
	log.Debug().Str("id", status.ID).Msg("updating status")
	status.lock.Lock()
	defer status.lock.Unlock()

	req := rt.compute.Instances.GetGuestAttributes(rt.projectID, rt.zone, status.VMName)
	req.QueryPath("fuzzing/password")
	attrs, err := req.Do()
	if err != nil {
		if herr, ok := err.(*googleapi.Error); ok && herr.Code == http.StatusNotFound {
			log.Debug().Str("id", status.ID).Str("vm-name", status.VMName).Msg("/fuzzing guest attributes not found yet")
			return
		}
		log.Error().Err(err).Str("id", status.ID).Str("vm-name", status.VMName).Msg("getting instance guest attributes")
		status.Error = createError + status.ID
		return
	}

	if len(attrs.QueryValue.Items) != 1 {
		log.Error().Err(err).Str("id", status.ID).Str("vm-name", status.VMName).Msgf("got %v guest attributes instead of 1", len(attrs.QueryValue.Items))
		status.Error = createError + status.ID
		return
	}

	// got the password
	status.Password = attrs.QueryValue.Items[0].Value
	log.Debug().Str("id", status.ID).Str("vm-name", status.VMName).Msg("got password")

	// get the IP
	// as the instance is running, it should have its IP already
	inst, err := rt.compute.Instances.Get(rt.projectID, rt.zone, status.VMName).Do()
	if err != nil {
		log.Error().Err(err).Str("id", status.ID).Msg("getting instance")
		status.Error = createError + status.ID
	}
	for _, n := range inst.NetworkInterfaces {
		for _, ac := range n.AccessConfigs {
			if ac.NatIP != "" {
				status.IP = ac.NatIP
			}
		}
	}
	if status.IP == "" {
		log.Error().Str("id", status.ID).Str("vm-name", status.VMName).Msg("failed to get instance IP")
		status.Error = "Failed to get instance IP. Request ID: " + status.ID
		return
	}

	status.Provisioned = true

	log.Info().Str("vm-name", status.VMName).Str("ip", status.IP).Str("user", status.Name).Msg("Provisioned machine")
}

func main() {
	// for compatibility with Google Cloud
	zerolog.LevelFieldName = "severity"
	zerolog.TimestampFieldName = "timestamp"

	log.Logger = log.With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	debug := os.Getenv("DEBUG")
	if debug == "1" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// for VM names
	insecureRand.Seed(time.Now().UnixNano())

	rt := Runtime{
		lock:     &sync.Mutex{},
		sessions: make(map[string]*Status),
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	rt.zone = os.Getenv("ZONE")
	if rt.zone == "" {
		rt.zone = "us-central1-a"
	}

	rt.projectID = os.Getenv("PROJECT")
	if rt.projectID == "" {
		log.Fatal().Msg("You must specify a GCP Project ID in the PROJECT environment variable")
	}

	templ := os.Getenv("TEMPL")
	if templ == "" {
		log.Fatal().Msg("You must specify a source instance template name in the TEMPL environment variable")
	}
	rt.sourceInstanceTemplate = "projects/" + rt.projectID + "/global/instanceTemplates/" + templ

	var err error
	rt.compute, err = compute.NewService(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't create compute API client")
	}

	authKeyStr := os.Getenv("COOKIE_KEY")
	authKey, err := base64.StdEncoding.DecodeString(authKeyStr)
	if err != nil || len(authKey) != 32 {
		log.Fatal().Err(err).Msg("invalid authentication key: COOKIE_KEY environment variable must be a base64-encoded 32 byte random value")
	}
	rt.cs = sessions.NewCookieStore(authKey)
	rt.cs.Options.HttpOnly = true
	rt.cs.Options.SameSite = http.SameSiteLaxMode
	rt.cs.Options.Secure = false // so we can test it locally

	maxVMsStr := os.Getenv("VMLIMIT")
	if maxVMsStr != "" {
		maxVMs, err := strconv.Atoi(maxVMsStr)
		if err != nil || maxVMs < 0 {
			log.Fatal().Err(err).Msg("VMLIMIT environment variable must be a non-negative integer. 0 = no limit")
		}
		rt.maxVMs = maxVMs
	}

	http.Handle("/", http.HandlerFunc(rt.handleRoot))
	http.Handle("/session", http.HandlerFunc(rt.handleSession))
	http.Handle("/provision", http.HandlerFunc(rt.handleProvision))
	http.Handle("/reset", http.HandlerFunc(rt.handleReset))
	http.Handle("/delete", http.HandlerFunc(rt.handleDelete))
	http.Handle("/mvp.css", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "mvp.css")
	}))
	http.Handle("/favicon.ico", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "favicon.ico")
	}))

	log.Info().Str("host", port).Msg("service started")
	log.Fatal().Msg(http.ListenAndServe("localhost:"+port, nil).Error())
}
