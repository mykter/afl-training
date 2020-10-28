# Building

This sample comes with a handy ready-made wrapper. main.c takes a file specified on the commandline rather than stdin,
but that will suit our purposes.

Compile with something like:

    make clean
    CC=afl-clang-fast make

(you could consider adding deferred forkserver or persistent mode support, or compiling with AFL_HARDEN=1 or ASAN)

# Running

Give afl a seed to run on:

    mkdir in
    echo a > in/1

Then you can tell afl-fuzz how to pass the filename to the target program using an "@@" notation:

    afl-fuzz -i in -o out ./m1-bad @@

# Seeds

Whilst afl does have a very impressive emergent synthesis capability, choosing good input files is an important
prerequisite for many targets.

If you want to speed up your fuzzing performance on this target, you can do better than seeding it with "a". It's a mime
parser, so if you know the format you could hand-write a few examples. Or you could fetch some off the web or another
source (e.g. your own email archives if in a suitable format).

Here's a helper seed to speed things up:

    echo -e "a=\nb=" > in/multiline

Note that files in "in" aren't monitored - you have to put them there before starting the fuzzer.

(This is definitely cheating by the way, a real sample set would include several other inputs, 'diluting' the search
space to some extent. I provide it (a) for if you get bored waiting for your fuzzer and don't want to go parallel and
(b) to demonstrate the performance difference - watch how long it takes to find paths with this input vs the "a" input.)

# Persistent mode and reproducibility

I have experienced non-reproducible crashes when using persistent mode with this target. It looks like it finds the bug,
but the bug doesn't cause a crash straight-away - it corrupts memory somewhere, and as we aren't using ASAN it isn't
immediately detected. A later run then leads to a crash. If you experience this too, simply remove the persistent mode
loop and resume your fuzzing run (`afl-fuzz -i- ...`) - it should quickly find a reproducible version of the crash now
it has a great starting point.

Take-away: persistent mode without ASAN might make crash triage harder.

# Answers

See ANSWERS
