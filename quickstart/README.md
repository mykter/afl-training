These instructions lead you through setup and fuzzing of a sample program.

Setup
========

Jump to the appropriate part of this Setup section based on what you're
configuring, then go to the next section (Building AFL).

Logging in to the provided instance
-------------------------------------

If you're reading these instructions then you've probably already made it! Skip to the Building AFL section.

Running the docker image locally
-----------------------------------

See the "Running locally" section of docker/README.md, then skip to the Building AFL section.

Setting up your own machine manually
---------------------------------------

Install dependencies:

    $ sudo apt-get install clang-4.0 build-essential llvm-4.0-dev gnuplot-nox

Work around some Ubuntu annoyances

    $ sudo update-alternatives --install /usr/bin/clang clang `which clang-4.0` 1
    $ sudo update-alternatives --install /usr/bin/clang++ clang++ `which clang++-4.0` 1
    $ sudo update-alternatives --install /usr/bin/llvm-config llvm-config `which llvm-config-4.0` 1
    $ sudo update-alternatives --install /usr/bin/llvm-symbolizer llvm-symbolizer `which llvm-symbolizer-4.0` 1

Make system not interfere with crash detection:

    $ echo core | sudo tee /proc/sys/kernel/core_pattern

Get afl:

    $ cd
    $ wget http://lcamtuf.coredump.cx/afl/releases/afl-latest.tgz
    $ tar xvf afl-latest.tgz

Building AFL
============

    $ cd afl-2.45b   # replace with whatever the current version is
    $ make
    $ make -C llvm_mode


The `vulnerable` program
========================

Build our quickstart program using the instrumented compiler:

    $ cd /path/to/quickstart # (e.g. ~/afl-training/quickstart)
    $ CC=~/afl-2.45b/afl-clang-fast AFL_HARDEN=1 make

Test it:

    $ ./vulnerable
    # Press enter to get usage instructions.
    # Test it on one of the provided inputs:
    $ ./vulnerable < inputs/c


Fuzzing
=======

Fuzz it:

    $ ~/afl-2.45b/afl-fuzz -i in -o out ./vulnerable

For comparison you could also test without the provided example inputs, e.g.:

    $ mkdir in
    $ echo "my seed" > in/a
    $ ~/afl-2.45b/afl-fuzz -i in -o out ./vulnerable

POC
====

Without instrumentation (`-n`): (`afl-fuzz -i inputs -o out -n ./vulnerable`)
![imaing](https://github.com/enovella/afl-training/blob/master/quickstart/pics/afl-pic-vulnerable-dumbf.png)

With instrumentation (`afl-clang-fast`):
![imaing](https://github.com/enovella/afl-training/blob/master/quickstart/pics/afl-pic-vulnerable.png)
