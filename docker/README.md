This Dockerfile produces a docker image set up ready for the training.


Building
========

    $ echo <this session's password> > password.txt
    $ docker build . -t afltraining
(note that the password is included in the image - don't publically publish the image with a password you're using for a live training session!)

Running Locally
===============

    $ docker run --privileged -p 22000:22 afltraining
    $ ssh fuzzer@localhost -p 22000

You need to use a privileged container to use the `asan_cgroups/limit_memory.sh` script.

Running on AWS
==============

Set up AWS host
---------------

- For a new Virtual Private Cloud:
    - Create a new VPC using the wizard - add a public subnet in availability zone A
    - If adding a VPC without the wizard:
        - Create a new subnet for it in availability zone A
        - Create a new internet gateway for it
        - Add a route to 0.0.0.0/0 via the new internet gateway
    - Optionally add flow logging for the VPC
- Create a `~/.aws/credentials` file:

        [default]
        aws_access_key_id = A...
        aws_secret_access_key = h...
- Create a hulking EC2 instance (~$2/hour):

        $ docker-machine create --driver amazonec2 --amazonec2-region eu-west-2 --amazonec2-vpc-id <yourVPC> --amazonec2-instance-type c4.8xlarge trainingaws
- Disable crash reporting and frequency scaling on the host:

        $ docker-machine ssh trainingaws
        $ echo core | sudo tee /proc/sys/kernel/core_pattern
        $ echo performance | sudo tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor

Get the image on the host
-------------------------

- Either save the image, transfer it, and load it:
    - Save the image:

            $ docker save afltraining | xz -9 > training.tar.xz
    - Copy it to the remote machine

            $ docker-machine scp training.tar.xz trainingaws:
    - Load it

            $ docker-machine ssh trainingaws
            $ unxz training.tar.xz
            $ sudo docker load -i training.tar
            $ exit
- Or build it remotely:

        $ docker-machine scp -r afl-training trainingaws:
        $ docker-machine ssh trainingaws
        $ cp afl-training/docker/Dockerfile ./
        $ sudo docker build -t afltraining .

Run instances
-------------

- Spin up some instances (assumes you've installed docker-machine bash wrapper). Note you need privileged containers to use the `asan_cgroups/limit_memory.sh` script.

        $ docker-machine use trainingaws
        $ docker-machine ip trainingaws
        $ for PORT in {30500..305NN}; do docker run --privileged -d -p $PORT:22 afltraining ; done
- In the EC2 control panel, check that the docker-machine security group for this VPC has open inbound ports from 30500-305NN

Cleanup
-------

Don't forget to destroy the machine when you're finished:

        $ docker-machine rm trainingaws
