These instructions lead you through setting up a Linux machine
(the instructions assume Ubuntu) to use afl to fuzz the sample
program.

1. Setup
========

Jump to the appropriate part of section 1 based on what you're
configuring, then go to section 2.

Setting up your own machine manually
---------------------------------------

Install dependencies:

    $ sudo apt-get install clang-3.8 build-essential llvm-3.8-dev gnuplot-nox

Work around some Ubuntu annoyances

    $ sudo update-alternatives --install /usr/bin/clang clang `which clang-3.8` 1
    $ sudo update-alternatives --install /usr/bin/clang++ clang++ `which clang++-3.8` 1
    $ sudo update-alternatives --install /usr/bin/llvm-config llvm-config `which llvm-config-3.8` 1
    $ sudo update-alternatives --install /usr/bin/llvm-symbolizer llvm-symbolizer `which llvm-symbolizer-3.8` 1

Make system not interfere with crash detection:

    $ echo core | sudo tee /proc/sys/kernel/core_pattern

Get afl:

    $ cd
    $ wget http://lcamtuf.coredump.cx/afl/releases/afl-latest.tgz
    $ tar xvf afl-latest.tgz

Running the docker image locally
-----------------------------------

From the root of this repository:

    $ docker build -t afltraining .
    $ docker run -it afltraining /bin/bash

Logging in to the provided instance
-------------------------------------

If you're reading these instructions then you've probably already made it!

2. Once you're at a shell in your configured machine/container
==============================================================

Build afl:

    $ cd afl-2.41b   # or whatever it is now
    $ make
    $ cd llvm_mode
    $ make

Build our quickstart program:

    $ cd /path/to/quickstart
    $ CC=~/afl-2.41b/afl-clang-fast AFL_HARDEN=1 make

Test it:

    $ ./vulnerable
    # type some input, e.g. "ececho!"
    $ ./vulnerable < inputs/c
    ...

Fuzz it:

    $ ~/afl-2.41b/afl-fuzz -i inputs -o out ./vulnerable

For comparison you could also test without the provided example inputs, e.g.:

    $ mkdir in
    $ echo "my seed" > in/a
    $ ~/afl-2.41b/afl-fuzz -i in -o out ./vulnerable
