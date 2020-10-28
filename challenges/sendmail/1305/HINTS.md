You'll want to adapt this harness to read the input from stdin instead of using a fixed address.

To skip this step, just do:

    cp prescan-overflow-bad-fuzz.c prescan-overflow-bad.c

The makefile uses CC as per normal, so the standard compilation approach will work, e.g.:

    CC=afl-clang-fast AFL_USE_ASAN=1 make

A sensible seed might be your email address, e.g.

    echo -n "michael@mykter.com" > in/seed

You may want to experiment with the deferred fork server and persistent mode, to see how much more performance you can
eke out of it.
