# Compiling

CC=clang CFLAGS="-g -fsanitize=fuzzer -DFUZZING_BUILD_MODE_UNSAFE_FOR_PRODUCTION -D_FORTIFY_SOURCE=2
-fstack-protector-all" make

Note as we aren't using ASAN, we turn on the less advanced mechanisms instead, akin to what AFL_HARDEN=1 does. Without
these it can be really hard to diagnose what is going on. I don't know what exactly happens when you try and use ASAN,
but it seems unable to progress at all.

# Running

./m1-bad

(yuck: this program dumps onto stdout)

# Harness

This challenge has a requirement that is likely to crop up somewhat regularly: the target works on a file pointer, not
on a buffer. What do? `<stdio.h>` has an answer: `fmemopen`, which presents a file interface to an in-memory buffer.

Definitely worth finding and deleting all of the printf statements in mime1-bad.c.

Copy the original main function to LLVMFuzzerTestOneInput, and hide it behind an #ifdef guard. Firstly delete the
commandline processing part:

```c
     assert(argc == 2);
     temp = fopen(argv[1], "r");
     assert(temp != NULL);
```

And then after the code that sets up the static `header` struct and initializes the envelope, arrange to pass in the
test case as a file, and close it afterwards:

```c
     if (len == 0)
     {
          // otherwise fmemopen will fail
          return 0;
     }
     e->e_dfp = fmemopen(data, len, "r");
     assert(e->e_dfp != NULL);

     mime7to8(header, e);

     fclose(e->e_dfp);
```
