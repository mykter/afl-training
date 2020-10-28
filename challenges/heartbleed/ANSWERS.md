An example way of getting data into sinbio, using stdin and some deferred forkserver goodness.

Insert this immediately prior to the BIO_write:

    #ifdef __AFL_HAVE_MANUAL_CONTROL
      __AFL_INIT();
    #endif

    uint8_t data[100] = {0};
    size_t size = read(STDIN_FILENO, data, 100);
    if (size == -1) {
      printf("Failed to read from stdin\n");
      return(-1);
    }

Note the rate at which it finds new paths isn't astronomical, but that's OK - give it a few minutes tops and you should
find success.

When you find a crash, stop AFL and run the crash through the fuzzer. You should see in the ASAN stacktrace that the
crash is in a heartbeat function - that's heartbleed!
