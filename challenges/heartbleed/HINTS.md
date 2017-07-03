Because we're using ASAN, we need to run this in a slightly different way, see docs/notes_for_asan.txt:

	sudo ~/afl-2.41b/experimental/asan_cgroups/limit_memory.sh -u fuzzer ~/afl-2.41b/afl-fuzz -i in -o out -m none ./handshake
