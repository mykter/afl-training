This Dockerfile produces a docker image set up ready for the training. It is available on the Docker Hub as [mykter/afl-training](http://hub.docker.com/r/mykter/afl-training).


Building
========

    $ docker build . -t afltraining

Running Locally
===============

    $ docker run --privileged -p 22000:22 -e PASSMETHOD=env -e PASS=<password> afltraining
    $ ssh fuzzer@localhost -p 22000

You need to use a privileged container to use the `asan_cgroups/limit_memory.sh` script and to use debuggers like gdb.

Running on AWS
==============

There are lots of ways to run containers on AWS. ECS with Fargate and 'manual' management on EC2 are described here.

SSH Credentials
---------------

Students each get their own container, which they can SSH in to as the `fuzzer` user. The password for this user can either be set through an environment variable, or dynamically retrieved from AWS SSM Parameter Store. In either case, all container instances share the same password - students are assumed to not be malicious towards each other. See the `entrypoint.sh` script for details.

ECS with Fargate
----------------
Fargate has the advantage of being less effort to manage - there is no machine provisioning and each container gets a public IP. It also makes container escapes by marauding students less likely and less impactful, though this isn't too important.

It has the limitation that your containers cannot be privileged, so are missing CAP_PTRACE for debugging and the necessary cgroup privileges for the ASAN script. It's still a viable option - ASAN usually works with `-m none`, and debugging is an optional extra part of the workshop.

These limitates aren't present when using ECS with EC2, but the networking limitations of ECS+EC2 outweigh the benefits in my opinion - it's easier to go down the DIY EC2 route.

EC2
---

First create a VPC with a public subnet, and a security group for this VPC has open inbound ports from 30500-305NN, SSH (22), and Docker (2376). If using the SSM password method, create an instance profile for the host with appropriate permissions.

Create a hulking EC2 instance (~$4/hour), e.g.:

        $ export REGION=<?> VPC=<?> SG=<?> PROFILE=<?>
        $ docker-machine create --driver amazonec2 --amazonec2-region $REGION --amazonec2-vpc-id $VPC --amazonec2-security-group $SG --amazonec2-iam-instance-profile=$PROFILE --amazonec2-instance-type m5a.24xlarge trainingaws
Update the host, then disable crash reporting and frequency scaling:

        $ docker-machine ssh trainingaws
        $ apt update && apt upgrade && shutdown -r now
        $ docker-machine ssh trainingaws
        $ echo core | sudo tee /proc/sys/kernel/core_pattern
        $ echo performance | sudo tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor # if this exists on the host

Repeat as necessary until you have enough hosts for your instances. To allow proper experimentation with multi-core fuzzing, don't schedule more than one container per two vCPUs.

Spin up some instances.
  - The docker-machine bash wrapper is used to simplify things.
  - We limit each instance to specific cores, so they don't clash with each other, and also tell the entrypoint script this (see below).
  - You need privileged containers to use the `asan_cgroups/limit_memory.sh` script.
  - This example uses the SSM Parameter Store method for password provisioning:

        $ docker-machine use trainingaws
        $ export CPUPER=3 INSTANCES=4 BASEPORT=30500
        $ for ((I=0; I<INSTANCES; I++))
          do
              MANUALCPUS="$((CPUPER * I))-$(((CPUPER * (I+1)) - 1))"
              docker run \
                  --cpuset-cpus=$MANUALCPUS -e MANUALCPUS -h "cpus-$MANUALCPUS" \
                  --privileged -d -p $((BASEPORT + I)):22 \
                  -e PASSMETHOD=awsssm -e PASSPARAM=<e.g. afltraining.password> -e PASSREGION=<e.g. eu-west-2> \
                  mykter/afl-training
          done

SSH in:

        $ ssh fuzzer@`docker-machine ip trainingaws` -p $BASE

For max performance, and when doing multi-core fuzzing, we have to manually specify which CPUs to bind to, because AFL can't tell which ones it has access to. In this example, we're on an instance that has access to CPU 4 as indicated in its hostname, set by the above docker run command. The entrypoint script has also already set AFL_NO_AFFINITY in the environment for us. I learned this technique from Brandon Falk's excellent post on [scaling AFL to 256 threads](https://gamozolabs.github.io/fuzzing/2018/09/16/scaling_afl.html).

        $ taskset -c 4 afl-fuzz -i inputs -o out ./vulnerable

Don't forget to destroy the machine when you're finished:

        $ docker-machine rm trainingaws
