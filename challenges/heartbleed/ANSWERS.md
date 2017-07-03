An example way of getting data into sinbio, using stdin and some deferred forkserver goodness.

Just insert this immediately prior to the BIO_write:

    #ifdef __AFL_HAVE_MANUAL_CONTROL
      __AFL_INIT();
    #endif

    uint8_t data[100] = {0};
    size_t size = read(STDIN_FILENO, data, 100);
    if (size == -1) {
      printf("Failed to read from stdin\n");
      return(-1);
    }
