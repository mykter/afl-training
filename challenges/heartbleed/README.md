This is adapted from the libFuzzer example here: https://github.com/google/fuzzer-test-suite/tree/master/openssl-1.0.1f

Get the openSSL source for version OpenSSL 1.0.1f:

    git submodule init
    git submodule update

Configure and build with ASAN:

    cd openssl
    CC=afl-clang-fast CXX=afl-clang-fast++ ./config -d
    AFL_USE_ASAN=1 make

(You can do "make -j" for faster builds, but there is a race that makes this fail occasionally)

(This target doesn't work with afl-clang-lto(++) at the time of writing)

The target-specific hard work has been done for you in this challenge: `handshake.cc` sets up the process as a server
awaiting a connection, and has prepared the OpenSSL-specific types necessary to send data to this server without
involving the network.

What remains is to fix up the code in handshake.cc to work with afl. (or copy it out of ANSWERS.md!). Once you're done
that, build our target:

    AFL_USE_ASAN=1 afl-clang-fast++ -g handshake.cc openssl/libssl.a openssl/libcrypto.a -o handshake -I openssl/include -ldl

Pre-emptive hints:

- Don't worry about seeds. This is easy to find without any.
- (You've already got the hint to use Address Sanitizer in the build commands above, but if you're familiar with the bug
  already, have a think about heartbleed and why it's important we use ASAN)
- Check out afl's docs/notes_for_asan.txt
