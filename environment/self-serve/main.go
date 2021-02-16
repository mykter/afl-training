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

State is maintained entirely in cookies.
*/

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"html/template"
	"io"
	insecureRand "math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
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
	Requestor      string
	ProvisionStart time.Time
	Provisioned    bool
	VMName         string
	IP             string
	Password       string
	Error          string
}

type Runtime struct {
	cs                     *sessions.CookieStore
	maxVMs                 int
	compute                *compute.Service
	zone                   string
	sourceInstanceTemplate string
	projectID              string
}

const sessionName = "state"
const username = "fuzzer"
const sshPort = 2222
const instanceNamePrefix = "fuzz-training-"
const instanceNameSuffixLen = 6
const createError = "Failed to create instance. Session ID: "
const requestorNameLengthLimit = 30

// cookie values
const sessionID = "id"
const sessionVM = "vm"
const sessionRequestor = "name"
const sessionStart = "start"
const sessionError = "error"

// render the provided main template within the site layout
func writeHtml(w io.Writer, main string, data interface{}) {
	err := template.Must(template.New("html").Parse(`
		{{define "main"}}`+main+`{{end}}
		{{- define "copy" -}}
			{{- if . | len | eq 0 }}
				{{- . }}
			{{- else }}
				{{- . }} <a href='#' onclick="navigator.clipboard.writeText({{.}});"><small>📋</small></a>
			{{- end }}
		{{- end}}
		{{define "machineDetails"}}
			<p><table>
				<tr><td>IP Address</td> <td>{{template "copy" .Status.IP}}</td></tr>
				<tr><td>Username</td> <td>`+username+`</td></tr>
				<tr><td>Password</td> <td>{{template "copy" .Status.Password}}</td></tr>
				<tr><td>Port</td> <td>`+fmt.Sprint(sshPort)+`</td></tr>
				<tr><td>ID</td> <td>{{.Status.VMName}}</td></tr>
			</table></p>
		{{end}}
		{{define "ssh"}}
			<p><code>ssh `+username+`@{{.Status.IP}} -p `+fmt.Sprint(sshPort)+`</code></p>
			<details>
                <summary>SSH config file entry</summary>
                <p>Add this to your ~/.ssh/config file to simplify access, especially helpful if using VSCode Remote:<p>
				<p><pre><code>Host fuzz-training
  HostName {{.Status.IP}}
  User `+username+`
  Port `+fmt.Sprint(sshPort)+`
</code></pre></p>
				<p>Now you can simply do <code>ssh fuzz-training</code></p>
            </details>
		{{end}}

		<head>
			<link rel="stylesheet" href="/mvp.css">
			<title>Fuzz Training</title>
		</head>
		<body>
			<main>
				<h1>Fuzz Training</h1>
				{{ template "main" .}}
			</main>
		</body>
		`)).Execute(w, data)
	if err != nil {
		log.Error().Err(err).Msg("Writing html")
	}
}

// Create a Status from the cookie, querying GCP API if a VMName is specified.
// If session is nil, gets it from the request
// Provided the cookie doesn't already have an error recorded, update the cookie with the updated status
// On error, set an http.Error and return false
func (rt *Runtime) getStatus(session *sessions.Session, w http.ResponseWriter, r *http.Request) (*Status, bool) {
	if session == nil {
		var err error
		session, err = rt.cs.Get(r, sessionName)
		if err != nil {
			log.Error().Err(err).Msg("getStatus reading cookie")
			http.Error(w, "error reading cookie", http.StatusInternalServerError)
			return nil, false
		}
	}

	status := &Status{}

	val, ok := session.Values[sessionID]
	if !ok {
		log.Warn().Msgf("id value not in cookie")
		http.Error(w, "failed to parse cookie; please enable cookies", http.StatusBadRequest)
		return nil, false
	}
	status.ID, ok = val.(string)
	if !ok {
		log.Warn().Msgf("id value in session was not the correct type. got '%v'", val)
		http.Error(w, "failed to parse cookie; please reset your cookies and refresh", http.StatusBadRequest)
		return nil, false
	}

	val, ok = session.Values[sessionVM]
	if ok {
		status.VMName, ok = val.(string)
		if !ok {
			log.Warn().Str("id", status.ID).Msgf("vm value in session was not the correct type. got '%v'", val)
			http.Error(w, "failed to parse cookie; please reset your cookies and refresh", http.StatusBadRequest)
			return nil, false
		}
	}

	val, ok = session.Values[sessionRequestor]
	if ok {
		status.Requestor, ok = val.(string)
		if !ok {
			log.Warn().Str("id", status.ID).Msgf("requestor value in session was not the correct type. got '%v'", val)
			http.Error(w, "failed to parse cookie; please reset your cookies and refresh", http.StatusBadRequest)
			return nil, false
		}
	}

	val, ok = session.Values[sessionStart]
	if ok {
		status.ProvisionStart, ok = val.(time.Time)
		if !ok {
			log.Warn().Str("id", status.ID).Msgf("time value in session was not the correct type. got '%v'", val)
			http.Error(w, "failed to parse cookie; please reset your cookies and refresh", http.StatusBadRequest)
			return nil, false
		}
	}

	val, ok = session.Values[sessionError]
	if ok {
		status.Error, ok = val.(string)
		if !ok {
			log.Warn().Str("id", status.ID).Msgf("error value in session was not the correct type. got '%v'", val)
			http.Error(w, "failed to parse cookie; please reset your cookies and refresh", http.StatusBadRequest)
			return nil, false
		}
	}

	// don't clobber any previous error - the user can manually reset if they want to
	if status.Error == "" {
		rt.updateStatus(status)
		rt.saveStatus(session, status, w, r)
	}
	return status, true
}

// Update status with the details of the provisioned VM
func (rt *Runtime) updateStatus(status *Status) {
	if status.VMName == "" {
		return
	}

	log.Debug().Str("id", status.ID).Msg("updating status")

	// get the instance - it might not exist if this cookie is stale
	inst, err := rt.compute.Instances.Get(rt.projectID, rt.zone, status.VMName).Do()
	if err != nil {
		if herr, ok := err.(*googleapi.Error); ok && herr.Code == http.StatusNotFound {
			log.Warn().Str("id", status.ID).Str("vm-name", status.VMName).Msg("VM not found")
			status.Error = "VM not found. Session ID: " + status.ID
			return
		}
		log.Error().Err(err).Str("id", status.ID).Str("vm-name", status.VMName).Msg("getting instance")
		status.Error = createError + status.ID
		return
	}

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
	// as the instance is already fully up, it should have its IP
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

	log.Info().Str("vm-name", status.VMName).Str("ip", status.IP).Str("requestor", status.Requestor).Msg("got provisioned machine details")
}

func deleteCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: sessionName, MaxAge: -1})
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
		log.Error().Err(err).Msg("handleRoot reading cookie")
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

		session.Values[sessionID] = id
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

	status, ok := rt.getStatus(session, w, r)
	if !ok {
		return
	}

	if status.Error != "" {
		log.Debug().Str("user-error", status.Error).Msg("serving error page")
		writeHtml(w,
			`<p>There was an error whilst provisioning your machine!</p><p><pre>{{.Status.Error}}</pre></p>
			<p>Please contact your facilitator. If they ask you to try again, please <a href="/reset">reset your session</a>.</p>`,
			struct{ Status *Status }{status})
	} else if status.VMName == "" {
		log.Debug().Msg("serving request form")
		writeHtml(w,
			`<form action="/provision" method="post">
				<header>
					<h2>Provision VM</h2>
				</header>
				<label for="input1">Your name (to aid the facilitator):</label>
				<input type="text" id="requestor" name="requestor" size="20">
				<button type="submit" onclick="this.form.submit(); this.innerText='Provisioning…'; this.disabled=true;">Provision</button>
			</form>`,
			nil)
	} else if !status.Provisioned {
		log.Debug().Msg("serving in-progress page")
		writeHtml(w,
			`<p>Your machine is being provisioned! This page will refresh every 5 seconds; your machine should be available within 3 minutes.</p>
			<p>It's been {{.Duration}} so far.</p>
			<p>Current details:</p>
			{{template "machineDetails" .}}
			<script>setTimeout("location.reload(true);",5000);</script>`,
			struct {
				Status   *Status
				Duration string
			}{
				status,
				time.Since(status.ProvisionStart).Round(time.Second).String(),
			})
	} else {
		log.Debug().Msg("serving success page")
		writeHtml(w, `<p>Success! Your machine details: </p>
			{{template "machineDetails" .}}
			{{template "ssh" .}}
			<p>Please <a href="/delete">delete 🗑️</a> your machine when you're finished.<p>`,
			struct{ Status *Status }{status},
		)
	}
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
		log.Error().Err(err).Msg("handleSession reading cookie")
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
		status, ok := rt.getStatus(nil, w, r)
		if !ok {
			writeHtml(w, `<p>Couldn't find session</p>`, nil)
			return
		}

		log.Info().Str("vm-name", status.VMName).Str("id", status.ID).Str("requestor", status.Requestor).Msg("deleting VM")

		req := rt.compute.Instances.Delete(rt.projectID, rt.zone, status.VMName)
		_, err := req.Do()
		if err != nil {
			log.Error().Err(err).Str("id", status.ID).Str("vm-name", status.VMName).Msg("deleting VM")
			status.Error = "Failed to delete VM " + status.VMName
			return
		}

		http.Redirect(w, r, "/reset", http.StatusFound)
		return
	}
	writeHtml(w,
		`<form action="/delete" method="post">
			<header>
				<h2>Delete VM?</h2>
				<p>Are you sure you want to delete your VM? You will lose any work that you haven't copied off it already.</p>
			</header>
			<button type="submit">Delete</button>
			<input type="button" name="cancel" value="cancel" onClick="window.location.href='/';" />
		</form>`,
		nil,
	)
}

// start provisioning if we haven't already, and redirect to /
func (rt *Runtime) handleProvision(w http.ResponseWriter, r *http.Request) {
	status, ok := rt.getStatus(nil, w, r)
	if !ok {
		return
	}

	requestor := strings.TrimSpace(r.FormValue("requestor"))
	if len(requestor) == 0 {
		http.Error(w, "you must enter a name", http.StatusBadRequest)
		return
	}
	if len(requestor) > requestorNameLengthLimit {
		requestor = requestor[:requestorNameLengthLimit]
	}
	status.Requestor = requestor

	if status.VMName != "" {
		log.Warn().Str("id", status.ID).Str("requestor", requestor).Msg("/provision visited by user that is already provisioning")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	rt.provision(status)
	rt.saveStatus(nil, status, w, r)
	http.Redirect(w, r, "/", http.StatusFound)
}

// save the status to the cookie
// if session is nil, get it from the request
func (rt *Runtime) saveStatus(session *sessions.Session, status *Status, w http.ResponseWriter, r *http.Request) {
	if session == nil {
		var err error
		session, err = rt.cs.Get(r, sessionName)
		if err != nil {
			log.Error().Err(err).Msg("saveStatus reading cookie")
			http.Error(w, "error reading cookie", http.StatusInternalServerError)
			return
		}
	}

	session.Values[sessionVM] = status.VMName
	session.Values[sessionStart] = status.ProvisionStart
	session.Values[sessionRequestor] = status.Requestor
	session.Values[sessionError] = status.Error
	err := session.Save(r, w)
	if err != nil {
		log.Error().Err(err).Str("id", status.ID).Msg("whilst saving VMName to session")
	}
}

func randomString(n int) string {
	var set = []rune("abcdefghkmnopqrstuvwxyz0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = set[insecureRand.Intn(len(set))] // nolint: gosec // this doesn't need to be unpredictable
	}
	return string(b)
}

// Start VM creation,
// updates status with relevant details
// updates runtime with VM count if successful
func (rt *Runtime) provision(status *Status) {
	log.Info().Str("id", status.ID).Str("requestor", status.Requestor).Msg("Started provisioning")
	status.ProvisionStart = time.Now()

	if rt.maxVMs > 0 { // a vmlimit is in force
		instances, err := rt.compute.Instances.List(rt.projectID, rt.zone).MaxResults(500).Do()
		if err != nil {
			log.Error().Err(err).Str("id", status.ID).Str("requestor", status.Requestor).Msg("listing instances")
			status.Error = createError + status.ID
			return
		}
		if len(instances.Items) >= rt.maxVMs {
			log.Warn().Int("num-vms", len(instances.Items)).Int("limit", rt.maxVMs).Str("id", status.ID).Msg("Hit VM limit")
			status.Error = "VM limit hit, cannot provision additional VMs."
			return
		}
		log.Debug().Int("num-vms", len(instances.Items)).Str("id", status.ID).Msg("counted instances")
	}

	status.VMName = instanceNamePrefix + randomString(instanceNameSuffixLen)

	ins := rt.compute.Instances.Insert(rt.projectID, rt.zone, &compute.Instance{
		Name:            status.VMName,
		ServiceAccounts: []*compute.ServiceAccount{},
		Zone:            rt.zone,
		// Metadata: can't alter this, as we have to use the template's metadata set via gcloud. GCP Compute doesn't have a REST api method for running containers
	})
	ins.SourceInstanceTemplate(rt.sourceInstanceTemplate)
	_, err := ins.Do()
	if err != nil {
		log.Error().Err(err).Str("id", status.ID).Str("requestor", status.Requestor).Msg("creating instance")
		status.Error = createError + status.ID
		return
	}
	log.Info().Str("id", status.ID).Str("requestor", status.Requestor).Str("vm-name", status.VMName).Msg("created instance")
}

func main() {
	// for compatibility with Google Cloud
	zerolog.LevelFieldName = "severity"
	zerolog.TimestampFieldName = "timestamp"

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	debug := os.Getenv("DEBUG")
	if debug == "1" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// So we can save times in cookies
	gob.Register(time.Time{})

	// for VM names
	insecureRand.Seed(time.Now().UnixNano())

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	rt := Runtime{}

	rt.zone = os.Getenv("ZONE")
	if rt.zone == "" {
		rt.zone = "us-east1-b"
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
	rt.cs.Options.Secure = os.Getenv("INSECURE_COOKIE") != "1"

	maxVMsStr := os.Getenv("VMLIMIT")
	if maxVMsStr != "" {
		maxVMs, err := strconv.Atoi(maxVMsStr)
		if err != nil || maxVMs < 0 || maxVMs > 500 { // 500 is the maximum number of instances the GCP API will return without having to deal with paging
			log.Fatal().Err(err).Msg("VMLIMIT environment variable must be an integer between 0 and 500. 0 = no limit")
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

	log.Info().Str("port", port).Msg("service started")
	log.Fatal().Msg(http.ListenAndServe("0.0.0.0:"+port, nil).Error())
}
