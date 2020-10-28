The environment variable in question is TZ:

    $ ./src/date
    Mon Jul  3 08:11:23 PDT 2017
    $ TZ='Asia/Tokyo' ./src/date
    Tue Jul  4 00:12:34 JST 2017

So, how to fuzz an environment variable? AFL doesn't have support for it built in. A few options:

1.  Find all instances in the source code where the TZ environment variable is read, and replace it with reading from
    stdin.
2.  Write a harness that sets the environment variable, then continues as normal (e.g. modify date.c's main function).
3.  Use LD_PRELOAD to replace calls to getenv with a custom wrapper that picks up the value from stdin.

Consider the pros and cons of each, and pick whichever suits!

(In this particular case, as the manpage highlights, you can also pass TZ in the command line. Follow that route if
you'd rather.)

Notice that the default behaviour of date is to print out the current time - this might interfere with afl's
determination of whether a particular input led to a change in execution path. Perhaps you could force it to output a
fixed time.

This bug can manifest itself with or without ASAN. ASAN will make triaging it easier and increases the likelyhood of
detection, but you could run your crashes or queue on an ASAN compiled version of it separately, if you want.

For fuzzing with ASAN: `$ AFL_USE_ASAN=1 make -j` and then refer to docs/notes_for_asan.txt

This target works with afl-clang-lto, with the customary long link time.
