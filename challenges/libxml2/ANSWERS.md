There are two options to help AFL out with fuzzing XML: using a good seed corpus, and/or using a dictionary. Both options would provide the fuzzer with a comprehensive set of tokens that have special meaning to XML-consuming software.

Happily AFL ships with a ready made XML dictionary, so we can use that:

    afl-fuzz -i in -o out -x ~/afl-2.52b/dictionaries/xml.dict ./fuzzer @@

You should see the numbers of paths found grow much faster using this approach.

Here is a complete example of a harness:
```c
#include "libxml/parser.h"
#include "libxml/tree.h"

int main(int argc, char **argv) {
    if (argc != 2){
        return(1);
    }

    xmlDocPtr doc = xmlReadFile(argv[1], NULL, 0);
    if (doc != NULL) {
        xmlFreeDoc(doc);
        return(0);
    }

    return(2);
}
```

The size of this wrapper (and the fact that it's all you need to find real bugs) reinforces some of the properties described in the README: libxml2 is very amenable to fuzzing.

With this wrapper, AFL should be able to find an out of bounds read in `xmlParseXMLDecl` a couple of minutes. There are some other bugs that it will also find, but they will likely take much longer to find. If your fuzzing session is discovering lots of paths but doesn't find any bugs, double check that you are using an ASAN enabled build.

This challenge was adapted from the [Fuzzer Test Suite](https://github.com/google/fuzzer-test-suite/), which contains a wealth of real world fuzz targets and harnesses.