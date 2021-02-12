# Fuzzing `cookedprint`

Here's an example ntpqmain() replacement that uses deferred forkserver and persistent mode:

    #ifdef __AFL_HAVE_MANUAL_CONTROL
            __AFL_INIT();
    #endif
            int datatype=0;
            int status=0;
            char data[1024*16] = {0};
            int length=0;
    #ifdef __AFL_HAVE_MANUAL_CONTROL
            while (__AFL_LOOP(1000)) {
    #endif
                    datatype=0;
                    status=0;
                    memset(data,0,1024*16);
                    read(0, &datatype, 1);
                    read(0, &status, 1);
                    length = read(0, data, 1024 * 16);
                    cookedprint(datatype, length, data, status, stdout);
    #ifdef __AFL_HAVE_MANUAL_CONTROL
            }
    #endif
            return 0;

The 16kb buffer is fairly arbitrary - it could be that a smaller buffer would achieve comparable coverage at higher
speed, or it could be that some bugs can't be hit because the limit is too low.

If you just run with this, you'll notice that the stability percentage is a little lower than it could be. Read up on
this in `docs/status_screen.txt`. It can be significantly improved by adding these lines to the start of `nextvar`, to
ensure that these static variables don't retain data from one run to the next:

    memset(name, 0, sizeof(name));
    memset(value, 0, sizeof(value));

# Coverage & Dictionaries

Without any help, afl won't easily find all of the different formats that can be returned from `varfmt`. This is
apparent when looking at the coverage output as described in HINTS.md - none of the cases except PADDING are hit.

In `varfmt` we notice it is checking strings against entries in an array called `cookedvars`. These can be easily
extracted into a dictionary for afl to use. Whilst we're at it, let's grab `tstflagnames` too as they look like they
might be useful.

In this case afl-clang-lto's auto-dictionary feature fails to detect the list of notable strings, perhaps because they
are indirectly referenced.

A dictionary with these keys in is included in this repo as `ntpq.dict`. Use it with the `-x` option to afl-fuzz, e.g.:

    afl-fuzz -i in -o out -x ntpq.dict ntp-4.2.8p8/ntpq/ntpq

With the help of this dictionary you'll see the number of paths found shoot up. Some issues have been found in 4.2.8p10
using this approach that have been reported but not fixed at the time of writing - if in doubt, contact
security@ntp.org.
