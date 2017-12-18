This is adapted from the libFuzzer example here: https://github.com/google/fuzzer-test-suite/tree/master/openssl-1.0.1f

- Clone the openSSL source from https://github.com/openssl/openssl.git
- Checkout at tag OpenSSL_1_0_1f
- Configure and build with ASAN:

		CC=~/afl-2.xx/afl-clang-fast CXX=~/afl-2.xx/afl-clang-fast++ ./config -d
		AFL_USE_ASAN=1 make
(note you can do "make -j" for faster builds, but there is a race that makes this fail occasionally)

Now fix up the code in handshake.cc to work with afl.  (or cheat and copy it out of ANSWERS)

Build our target:

	AFL_USE_ASAN=1 ~/afl-2.xx/afl-clang-fast++ -g handshake.cc openssl/libssl.a openssl/libcrypto.a -o handshake -I openssl/include -ldl

Pre-emptive hint:
 - Don't worry about seeds. This is easy to find without any.
 - (You've already got the hint to use Address Sanitizer in the build commands above, but have a think about heartbleed and why it's important we use ASAN)
 - Check out afl's docs/notes_for_asan.txt
