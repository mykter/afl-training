This is CVE-2017-7476. The fix and vulnerability are described here:
http://git.savannah.gnu.org/gitweb/?p=gnulib.git;a=commit;h=94e01571507835ff59dd8ce2a0b56a4b566965a4

To test that your built version has the vulnerability and is compiled with ASAN, run the POC:

    TZ="aaa00000000000000000000aaaaaab00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" ./src/date --date "2017-03-14 15:00 UTC"

(alter as needed if you've already patched date and the TZ value is coming from stdin - in this case note there is no
trailing newline!)

From the options in HINTS:

1.  This could work, but involves some potentially messy digging around in the code. Are you _sure_ you got every
    instance? Does your code work if it's retrieved multiple times?
2.  In afl's utils/bash_shellshock directory there is a trivial patch for bash that uses this approach. It isn't
    bash-specific - you just make a call to `setenv` - so is easily adapted to this scenario. This is the simplest and
    most robust method, so probably the most suitable for this task. You can insert the same code at the start of
    `src/date.c`'s main function, but change the env var it is setting to TZ.
3.  Creating a getenv replacement that is loaded via LD_PRELOAD might be useful work that you could reuse for other
    targets. But it's more effort than you need right now.

A suitable fuzzing command on a 64 bit machine might be:

    $ echo -n "Europe/London" > in/london

(note the "-n" to suppress a trailing new line)

For vanilla fuzzing:

    $ afl-fuzz -i in -o out -- ./coreutils/src/date --date "2017-03-14 15:00 UTC"

Note the fixed date, but with a timezone specified to ensure we don't skip past any timezone processing code. (this bug
is triggered whether or not you specify a timezone in the date - but consider how this could affect what is tested!)

For ASAN fuzzing:

    $ sudo ~/AFLplusplus/utils/asan_cgroups/limit_memory.sh -u fuzzer ~/AFLplusplus/afl-fuzz -i in -o out -- ./coreutils/src/date --date "2017-03-14T15:00-UTC"

Note the format of the date - when called via the `limit_memory.sh` script the quotes get lost - you can either escape
them or use a date format without spaces.

(If you're having trouble with ASAN, you can try being lazy and just running without the cgroup-based memory limiter.
The kernel might start killing off your processes if you hit an OOM condition, but in this instance it will probably be
ok...)

If your ASAN output in the crash doesn't give line numbers just memory addresses, check you've followed the quickstart
README and setup llvm-symbolizer prior to compiling.
