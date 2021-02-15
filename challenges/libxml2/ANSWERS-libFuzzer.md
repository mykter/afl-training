This file describes the changes from the approach to AFL. Read the AFL README/HINTS/ANSWERS files first.

# Compiling (README)

To compile libxml2 for libFuzzer, use:

```shell
    CC=clang CFLAGS=-fsanitize=fuzzer-no-link,address ./autogen.sh
    make -j4
```

Note the fuzzer-no-link: this compiles the library with the right instrumentation, but doesn't attempt to link in
libFuzzer's `main`.

Now compile your harness:

```shell
    clang -g -O2 -fsanitize=fuzzer,address ./libfuzzer-harness.c -I libxml2/include/ libxml2/.libs/libxml2.a -lz -lm -o libxml2-libfuzzer
```

# Running (HINTS)

```shell
    ./libxml2-libfuzzer -max_len=64 -dict=/home/fuzzer/AFLplusplus/dictionaries/xml.dict
```

Note we're specifying a pretty small limit on the input size - 64 bytes is enough to exercise a lot of XML
functionality, and the smaller the input the faster it will run. We're also cheating: we happen to know this bug doesn't
need a larger input to trigger it.

You should get a result in a few minutes.

You can go multi-core by specifying `-jobs=2` in the fuzzer invocation, for example. To track progress, try
`tail -f fuzz-0.log`.

# Harness (ANSWERS)

This is probably the smallest harness you'll ever write. (You can make it smaller still by removing the error handling
part, but that doesn't play nicely with libFuzzer, as it will clobber the output and slow down execution.)

```c
#include <stdint.h>
#include <libxml/parser.h>
#include <libxml/tree.h>

void quietError(void *ctx, const char *msg, ...) {}

int LLVMFuzzerTestOneInput(const uint8_t *Data, size_t Size)
{
    xmlSetGenericErrorFunc(NULL, &quietError); // suppress all error output
    xmlDocPtr doc = xmlReadMemory((const char *)Data, Size, "noname.xml", NULL, 0);
    if (doc != NULL)
    {
        xmlFreeDoc(doc);
    }
}
```
