There are two options to help AFL out with fuzzing XML: using a good seed corpus and/or using a dictionary. Both options
would provide the fuzzer with a comprehensive set of tokens that have special meaning to XML-consuming software.

Happily AFL ships with a ready made XML dictionary, so we can use that:

    afl-fuzz -i in -o out -x ~/AFLplusplus/dictionaries/xml.dict ./fuzzer @@

You should see the numbers of paths found grow much faster using this approach. Crucially, we'll also uncover a bug that
would never be found without it.

Here is a complete example of a harness using persistent mode. By lifting the parser initializaiton and cleanup outside
the loop you get a moderate speed-up at the cost of a slight stability-loss. We can drop the init/cleanup altogether,
but that has a significant (7%) stability impact.

```c
#include "libxml/parser.h"
#include "libxml/tree.h"

int main(int argc, char **argv) {
    if (argc != 2){
        return(1);
    }

    xmlInitParser();
    while (__AFL_LOOP(1000)) {
        xmlDocPtr doc = xmlReadFile(argv[1], NULL, 0);
        if (doc != NULL) {
            xmlFreeDoc(doc);
        }
    }
    xmlCleanupParser();

    return(0);
}
```

The small size of this wrapper (and the fact that it's all you need to find real bugs) reinforces some of the properties
described in the README: libxml2 is very amenable to fuzzing.

Here's another harness, this time utilizing the in-memory fuzzing capabilities of persistent mode, for another massive
2x speed boost over the previous harness.

```c
#include "libxml/parser.h"
#include "libxml/tree.h"
#include <unistd.h>

__AFL_FUZZ_INIT();

int main(int argc, char **argv) {
    #ifdef __AFL_HAVE_MANUAL_CONTROL
        __AFL_INIT();
    #endif
    unsigned char *buf = __AFL_FUZZ_TESTCASE_BUF;  // must be after __AFL_INIT

    xmlInitParser();
    while (__AFL_LOOP(1000)) {
        int len = __AFL_FUZZ_TESTCASE_LEN;
        xmlDocPtr doc = xmlReadMemory((char *)buf, len, "https://mykter.com", NULL, 0);
        if (doc != NULL) {
            xmlFreeDoc(doc);
        }
    }
    xmlCleanupParser();

    return(0);
}
```

With this wrapper and ASAN instrumentation, AFL should be able to find an out of bounds read in `xmlParseXMLDecl`. How
long it takes to find is dependent on luck, but my single core persistent mode run took about 15 minutes @ ~4k exec/s.
To speed things up beyond the persistent-mode gains, try multi-core fuzzing. There are some other bugs that it will also
find, but they will likely take much longer.

This challenge was adapted from the [Fuzzer Test Suite](https://github.com/google/fuzzer-test-suite/), which contains a
wealth of real world fuzz targets and harnesses.

AFLplusplus have also [written a tutorial](https://aflplus.plus/docs/tutorials/libxml2_tutorial/) around fuzzing this
target, and include an example of how to fuzz it using QEMU mode, as if you didn't have the source code.
