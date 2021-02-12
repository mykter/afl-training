This is adapted from the libFuzzer example here: https://github.com/google/fuzzer-test-suite/tree/master/openssl-1.0.1f

- Get the openSSL source for version OpenSSL_1_0_1f:

  git submodule init git submodule update

- Configure and build with ASAN:

      	cd openssl
      	CC=afl-clang-fast CXX=afl-clang-fast++ ./config -d
      	AFL_USE_ASAN=1 make

  (note you can do "make -j" for faster builds, but there is a race that makes this fail occasionally)

  (This target doesn't work with afl-clang-lto(++) at the time of writing)

Now fix up the code in handshake.cc to work with afl. (or copy it out of ANSWERS.md!)

Build our target:

    AFL_USE_ASAN=1 afl-clang-fast++ -g handshake.cc openssl/libssl.a openssl/libcrypto.a -o handshake -I openssl/include -ldl

Pre-emptive hint:

- Don't worry about seeds. This is easy to find without any.
- (You've already got the hint to use Address Sanitizer in the build commands above, but have a think about heartbleed
  and why it's important we use ASAN)
- Check out afl's docs/notes_for_asan.txt
