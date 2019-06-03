libxml2 is a popular XML library. Libraries like this are perfect for fuzzing, they tick all the boxes:
 [x] Often parse user supplied data
 [x] Written in an unsafe language
 [x] Stateless
 [x] No network or filesystem interaction
 [x] The public documented API contains good targets - no need to identify and isolate an internal component to fuzz
 [x] Fast

This makes it an ideal first target to write a fuzz harness for.

Build and test v2.9.2 with AFL and Address Sanitizer instrumentation by running:
```shell
    git submodule init
    git submodule update
    cd libxml2
    CC=afl-clang-fast ./autogen.sh
    AFL_USE_ASAN=1 make -j 4
    ASAN_OPTIONS=detect_leaks=0 ./testModule    # leak detection doesn't work with the libxml2 test suite, for some reason
```
Now we have a working instrumented build of the library, but no fuzzing harness to use.

Check out the docs - the [examples](http://xmlsoft.org/examples/index.html) are perhaps the easiest to grok - and consider what might be a good approach to creating a fuzzing harness.

If you're comfortable experimenting or confident in your approach, implement a harness and see if you can find any bugs! Or, move right on to [HINTS.md](./HINTS.md) for some specific guidance on making a good libxml2 fuzzing harness.

Once you've implemented a harness, you can compile it using a command like this:

    afl-clang-fast ./harness.c -I libxml2/include libxml2/.libs/libxml2.a -lz -lm -o fuzzer