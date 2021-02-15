Because we're using ASAN, we need to run this in a slightly different way, see docs/notes_for_asan.txt:

    sudo ~/AFLplusplus/examples/asan_cgroups/limit_memory.sh -u fuzzer afl-fuzz -i in -o out ./handshake

An alternative is to not use the limit_memory script. afl-fuzz defaults to using `-m none`, so this will work, but it
runs the risk of the target allocating a huge amount of memory, and Linux will then start killing processes underneath
you.
