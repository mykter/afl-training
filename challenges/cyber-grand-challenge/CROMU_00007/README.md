This is an extract from the cyber grand challenge. README.cgc.md has a descriptions of the program and vulnerabilities.
This version has been modified so that it runs on Linux not DECREE, the custom OS the challenge was designed for.

There are two* vulnerabilities in this program. The first is easily accessible.

The second vulnerability looks like it is very hard to hit with fuzzing - crash.input has a sample crashing input for it.

Have a go and let me know if my intuition is correct!

You might want to fix the first vulnerability once you've found it so it's easier to notice if you catch the second. The source contains a patch - just tweak some ifdefs (but note if you define PATCHED, you'll also patch out the other vulnerability!).

Definitely one for multicore; persistent mode; and use of a dictionary (or a starting set of input files that contain all of the keywords). Quite possibly one for manual source code review instead of fuzzing.

An interesting one to put into afl-analyse to see what it notices about the file format - it correctly annotates much of it, but misses on a few.


\* unless I introduced more in porting to Linux
