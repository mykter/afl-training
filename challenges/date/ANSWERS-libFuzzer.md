# Compiling

You can't compile the whole program with `-fsanitize=fuzzer`, because the configure script will fail to pass its sanity
check. Instead we configure with `-fsanitize=fuzzer-no-link` but then manually override that when it comes to actually
compiling `date`.

1.  Follow the instructions in README.md, but using these compiler options:
    `CC=clang CFLAGS="-fsanitize=address,fuzzer-no-link -g"`

2.  Rebuild the date binary, using libFuzzer's main:
    `rm src/date src/date.o && make src/date CFLAGS="-g -fsanitize=address,fuzzer -DFUZZING_BUILD_MODE_UNSAFE_FOR_PRODUCTION"`
    (note if you haven't replaced the existing main function yet, this will fail to link as there will be two mains)

# Running

This version of date leaks memory (which isn't hugely important given that date terminates immediately after doing its
job), but ASAN will detect that and complain. Tell it not to: `src/date -detect_leaks=0 in/`

# Harness

Here is one way to modify `date.c` to have libFuzzer invoke the original main function with a fixed "commandline" and TZ
set based on the fuzzer's input.

As this target uses getopt, and we care about the commandline used, there is global state we have to manage to ensure
getopt parses the commandline each time.

We're using ifdefs so we can compile the target both in its original form and for fuzzing.

You could also replace all references to stdout with a reference to /dev/null - see the ntpq's ANSWERS-libFuzzer.md for
an example of this.

```c
int
#ifdef FUZZING_BUILD_MODE_UNSAFE_FOR_PRODUCTION
origmain(int argc, char **argv)
#else
main(int argc, char **argv)
#endif

...

int LLVMFuzzerTestOneInput(uint8_t *data, size_t len)
{
  char *tz = calloc(1, len + 1);
  memcpy(tz, data, len);
  setenv("TZ", tz, 1);

  char *argv[] = {"date", "--date=2017-03-14 15:00 UTC", NULL};
  optind = 1; // getopt is used, which has global state that needs to be reset
#ifdef FUZZING_BUILD_MODE_UNSAFE_FOR_PRODUCTION
  origmain(2, argv);
#else
  // this will never get called, but it demonstrates a good practice of keeping your fuzz harnesses building alongside the normal application
  main(2,argv);
#endif
  free(tz);
  return 0;
}
```
