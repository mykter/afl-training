This Dockerfile produces a docker image set up ready for the training.

# Building

Optional, as the image is available at
[ghcr.io/mykter/fuzz-training](http://ghcr.io/mykter/fuzz-training).

    $ docker build . -t fuzz-training

# Running Locally

    $ docker run --privileged -p 2222:2222 -e PASSMETHOD=env -e PASS=<password> ghcr.io/mykter/fuzz-training
    $ ssh fuzzer@localhost -p 2222

You need to use a privileged container to use the `asan_cgroups/limit_memory.sh` script and to use debuggers like gdb.

# Running in the cloud

There are lots of ways to run containers in the cloud!

Prior to the GrayHat conference, I used a single large VM with multiple instances of the container running on it. This
caused performance and management complications though, so now use a single VM per student. See the history of this file
for guidance on that approach on AWS.

Currently this document describes running on GCP, and assumes you have configured a default project, region, and zone.

## SSH Credentials

Students each get their own container, which they can SSH in to as the `fuzzer` user. The password for this user can
either be set through an environment variable, dynamically retrieved from AWS SSM Parameter Store, or randomly generated
on startup and reported back to a listening host. See the `entrypoint.sh` script for details.

## VMs

First create a VPC with a public subnet, and a firewall rule that allows inbound ssh (22).

Create a template for an n2-standard-4 which gives a student 4 cores & 16gb RAM for ~\$0.2/hour. The template can
specify the VM should use a container OS and run the container, which is a lovely feature.

You need privileged containers for CAP_PTRACE for debugging and the necessary cgroup privileges to use the
`asan_cgroups/limit_memory.sh` script. It's still a viable option to not use privileged containers - ASAN usually works
with `-m none` and `ASAN_OPTIONS=detect_leaks=0`, and debugging is not the focus of the workshop.

We have to configure the container's SSH daemon to listen on a non-default port, as otherwise it conflicts with the
host's.

        $ gcloud compute networks create fuzz-training --subnet-mode=custom
        $ gcloud compute networks subnets create main --network=fuzz-training --enable-flow-logs --range=10.0.0.0/24 --region=us-central1
        $ gcloud compute firewall-rules create allow-workshop-ssh \
                --allow=tcp:2222 --direction=INGRESS --network=fuzz-training --target-tags=student
        $ gcloud compute instance-templates create-with-container fuzz-training \
                --container-image ghcr.io/mykter/fuzz-training \
                --container-privileged \
                --container-env PASSMETHOD=<see below>,SSHPORT=2222,SYSTEMCONFIG=1 \
                --machine-type n2-standard-4 \
                --region us-central1 \
                --network fuzz-training --subnet main \
                --no-service-account --no-scopes \
                --metadata enable-guest-attributes=TRUE \
                --tags student \
                --description "4 core machine running the afl training workshop container image in privileged mode, ssh listening on port 2222"

When creating the instance template you need to specify how to set the password. There are various options, each section
below describes one.

## Individually created VMs

To fix a static password for all instances, use:

            --container-env PASSMETHOD=env,PASS=secret,...

Then try something like this:

    $ gcloud compute instances create fuzz-training-{} --zone=us-central1-a --source-instance-template=fuzz-training

## Facilitator-managed VMs

As a facilitator you can trigger the creation of a number of VMs and have them generate and send their IP address and
SSH password to you, for you to forward on to students.

Create a tiny instance and a firewall rule that allows internal inbound connections on a port such as 1234:

        $ gcloud compute firewall-rules create allow-workshop-pass \
                --allow=tcp:1234 --direction=INGRESS --network=fuzz-training --target-tags=controller --source-tags=student
        $ gcloud compute instances create controller \
                --machine-type=e2-micro --no-service-account --no-scopes \
                --network-interface=private-network-ip=10.13.37.2,network=fuzz-training,subnet=main \
                --tags=controller \
                --description="password manager"

When creating the instance template, use:

            --container-env PASSMETHOD=callback,PASSHOST=<ip of controller>,PASSPORT=1234,...

Spin up some instances (number defined in the brace expansion at the end).

      $ parallel --verbose --keep-order "gcloud compute instances create fuzz-training-{} --zone=us-central1-a --source-instance-template=fuzz-training" ::: {1..2}

## Self-service VMs

With larger groups you can offer a self-service option via a simple website hosted on Cloud Run, that allows students to
provision their own machine. Note it's unauthenticated and hence open to abuse! You should shut it down once the session
is over or everyone has a VM.

Create the template with:

            --container-env PASSMETHOD=gcpmeta,...

Create a service account that has the Compute Admin role (or at least enough permissions to create VMs and read their
metadata), and then deploy a Cloud Run service:

        $ cd self-serve && docker build . -t gcr.io/<myproj>/fuzz-training-provisioner && docker push gcr.io/<myproj>/fuzz-training-provisioner
        $ export KEY=$(head -c 32 /dev/urandom | base64)
        $ gcloud run deploy fuzz-training-provisioner --image=gcr.io/myproj/fuzz-training-provisioner \
                --allow-unauthenticated --max-instances=1 \
                --set-env-vars=PROJECT=my-gcp-proj,ZONE=us-central1-a,TEMPL=fuzz-training,COOKIE_KEY=${KEY},VMLIMIT=100,DEBUG=0 \
                --service-account=<self-serve-account-id>

Point your students at the URL the service is deployed to, and they will be able to create a VM and get the credentials
for it. They can also delete their VM when they've finished with it. Note the VMLIMIT applies to all of the compute
instances in the specified zone - if you have existing instances they are counted against the limit.

Don't forget to delete the service when you're done: it offers free compute to anyone who can find the URL!

        $ gcloud run services delete fuzz-training-provisioner

## Running

SSH in (with the port configured previously):

        $ ssh fuzzer@<IP> -p 2222

For max performance, and when doing multi-core fuzzing, we should manually specify which CPUs to bind to, because AFL
can't tell which ones it has access to. In this workshop, as each student has their own VM and all of the cores on it,
we don't bother with this optimization however. See Brandon Falk's excellent post on
[scaling AFL to 256 threads](https://gamozolabs.github.io/fuzzing/2018/09/16/scaling_afl.html) for details (or the
history of this file for how to use it).
