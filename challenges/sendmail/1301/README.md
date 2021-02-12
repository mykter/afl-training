An extract from sendmail, the ubiquitous SMTP server.

This function processes email bodies to convert between 7-bit MIME and 8-bit MIME.

The program already comes with a test harness around it (in `main.c`), that runs the conversion routine on a file
specified on the commandline.

To test it out:

    make
    echo "hi!" > input
    ./m1-bad input

You can see from the output (or the source) that the code has been fairly extensively instrumented with debug prints to
watch out for a buffer overflow. Fortunately we don't need to worry about this: afl is going to find the problems for
us.

I recommend you experiment with persistent mode and/or a multicore setup for this challenge.

For multicore read afl's docs/parallel_fuzzing.txt for fairly straightforward instructions on single-system fuzzing. Get
a main and 3 secondaries running to slash the time it takes to solve this challenge.

For persistent mode read `AFLplusplus/instrumentation/README.persistent_mode.md`. There are some gotchas with this
target! See HINTS.md

For hints on fuzzing, see HINTS. If you get your fuzzer running without needing the hints, read the HINTS file whilst
watching the UI for some dicussion on seed file selection. Also check out docs/status_screen.txt.

Warning: this challenge doesn't play nicely with ASAN. Save ASAN for another challenge like heartbleed.
