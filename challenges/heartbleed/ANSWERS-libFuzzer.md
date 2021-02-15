This file describes the changes from the approach to AFL. Read the AFL README/HINTS/ANSWERS files first.

# Compiling (README)

To compile for libFuzzer, use:

```shell
    CC="clang -fsanitize=fuzzer-no-link,address" ./config
    make clean && make
```

`-fsanitizer=fuzzer-no-link`: this compiles the library with the fuzzer's instrumentation, but doesn't attempt to link
in libFuzzer's `main`.

Now compile your harness:

```shell
    clang -g -O2 -fsanitize=fuzzer,address libfuzzer-handshake.cc openssl/libssl.a openssl/libcrypto.a -o handshake-libfuzzer -I openssl/include
```

# Running

```shell
    ./handshake-libfuzzer
```

You should get a heap buffer overflow in the tls1_process_heartbeat function in a few seconds.

# Harness

The existing handshake.cc code is almost a libFuzzer harness already (this isn't a happy accident - it comes from a
fuzzing test repo!). Just rename main to `extern "C" int LLVMFuzzerTestOneInput(const uint8_t *data, size_t size)`, we
need the extern C bit as we're compiling with a C++ compiler.
