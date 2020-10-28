This Dockerfile produces a docker image set up ready for the training. It is available on the Docker Hub as
[mykter/afl-training](http://hub.docker.com/r/mykter/afl-training).

# Building

    $ docker build . -t afltraining

# Running Locally

    $ docker run --privileged -p 22000:22 -e PASSMETHOD=env -e PASS=<password> afltraining
    $ ssh fuzzer@localhost -p 22000

You need to use a privileged container to use the `asan_cgroups/limit_memory.sh` script and to use debuggers like gdb.

# Running in the cloud

There are lots of ways to run containers in the cloud!

Prior to the Grayhat conference, I used a single large VM with multiple instances of the container running on it. This
caused performance and management complications though, so now use a single VM per student. See the history of this file
for guidance on that approach on AWS.

Currently this document describes running on GCP, and assumes you have configured a default project, region, and zone.

## SSH Credentials

Students each get their own container, which they can SSH in to as the `fuzzer` user. The password for this user can
either be set through an environment variable, dynamically retrieved from AWS SSM Parameter Store, or randomly generated
on startup and reported back to a listening host. See the `entrypoint.sh` script for details.

## VMs

First create a VPC with a public subnet, and a firewall rule that allows inbound ssh (22).

If you want to receive passwords on a machine also running in this network, then create a tiny instance and a firewall
rule that allows internal inbound connections on a port such as 1234.

        $ gcloud compute firewall-rules create allow-workshop-pass \
                --allow=tcp:1234 --direction=INGRESS --network=grayhat --target-tags=controller --source-tags=student
        $ gcloud compute instances create controller \
                --machine-type=e2-micro --no-service-account --no-scopes \
                --network-interface=private-network-ip=10.13.37.2,network=grayhat,subnet=main \
                --tags=controller \
                --description="password manager"

Create a template for an n2-standard-4 which gives a student 4 cores & 16gb RAM for ~\$0.2/hour. The template can
specify the VM should use a container OS and run the container, which is a lovely feature.

You need privileged containers for CAP_PTRACE for debugging and the necessary cgroup privileges to use the
`asan_cgroups/limit_memory.sh` script. It's still a viable option to not use privileged containers - ASAN usually works
with `-m none` and `ASAN_OPTIONS=detect_leaks=0`, and debugging is not the focus of the workshop.

We have to configure the container's SSH daemon to listen on a non-default port, as otherwise it conflicts with the
host's.

        $ gcloud compute firewall-rules create allow-workshop-ssh \
                --allow=tcp:2222 --direction=INGRESS --network=grayhat --target-tags=student
        $ gcloud compute instance-templates create-with-container afl-training \
                --container-image mykter/afl-training \
                --container-privileged \
                --container-env PASSMETHOD=callback,PASSHOST=10.13.37.2,PASSPORT=1234,SSHPORT=2222,SYSTEMCONFIG=1 \
                --machine-type n2-standard-4 \
                --region us-central1 \
                --network grayhat --subnet main \
                --no-service-account --no-scopes \
                --tags student \
                --description "4 core machine running the afl training workshop container image in privileged mode, with password callbacks to 10.13.37.2, ssh listening on port 2222"

Spin up some instances (number defined in the brace expansion at the end).

      $ parallel --verbose --keep-order "gcloud compute instances create afl-training-{} --zone=us-central1-a --source-instance-template=afl-training" ::: {1..2}

## Running

SSH in (with the port configured previously):

        $ ssh fuzzer@<IP> -p 2222

For max performance, and when doing multi-core fuzzing, we have to manually specify which CPUs to bind to, because AFL
can't tell which ones it has access to. In this workshop, as each student has their own VM and all of the cores on it,
we don't bother with this optimization however. See Brandon Falk's excellent post on
[scaling AFL to 256 threads](https://gamozolabs.github.io/fuzzing/2018/09/16/scaling_afl.html) for details (or the
history of this file for how to use it).
