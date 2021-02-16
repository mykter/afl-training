libxml2 has quite a broad API, but the most tempting target is the core XML parsing logic. This is going to be easy to
fuzz, is definitely exposed to user-provided input in many situations, and is at high risk of containing bugs (parsing a
complex data format in an unsafe language).

This functionality is exposed in the [parser](http://xmlsoft.org/html/libxml-parser.html) API, and whilst you could dig
through this documentation, the easiest approach is to look at an example.
[`parse1.c`](http://xmlsoft.org/examples/parse1.c) (also in the repo under doc/examples/parse1.c) shows two core
functions: `xmlReadFile` followed by `xmlFreeDoc`.

Recall the type of program AFL can fuzz (from AFL's QuickStartGuide.txt):

    Find or write a reasonably fast and simple program that takes data from
    a file or stdin, processes it in a test-worthy way, then exits cleanly.

We can wrap these two library functions in a little boiler-plate code to call xmlReadFile with a file specified as a
commandline argument, and then free the resulting parsed document. As libxml can be compiled with clang and the API
should be stateless, we can also include the AFL persistent-mode loop.

Equivalently (looking at `parse3.c`), we could wrap `xmlReadMemory` with code that reads from stdin instead of a
specified filename.

Both of these approaches are good, but from here on we'll just look at the `xmlReadFile` option for simplicity.

Once you've implemented the harness, compile it (refer back to README.md for the include & linker flags you need with
libxml2), and then test your `fuzzer` executable by specifying an XML file on the commandline, e.g.
`./fuzzer ./libxml2/regressions.xml`. There shouldn't be any visible result (unless you added some kind of output to
your harness). We're now ready to fuzz in the usual manner for an ASAN-instrumented binary; here's a reminder of how to
do it for the file-argument approach:

```shell
    mkdir in
    echo "<hi></hi>" > in/hi.xml
    afl-fuzz -i in -o out ./harness @@
```

This will work, but as XML isn't a compact binary format, a lot of its syntax will remain undiscovered by the fuzzer
without further help.

Note we're using the default memory limit, which (as of AFL++ 3.0c) is `-m none` - this is _probably_ safe in this
instance, but there's always a chance the library will try and allocate a ridiculous amount of memory, which could cause
system instability. We also are less likely to detect bugs where an attacker can cause the process to allocate
disproportionately large amounts of memory. See AFL's `notes_for_asan.txt` doc for more robust approaches, one of which
is also covered in the heartbleed challenge.
