If you've got AFL working then even without a good seed or parallelisation it might find a crashing input in the order
of 15 minutes (subject to how fast your computer is and how lucky you get). It might take much longer.

With the cheater-seed from HINTS you should find a crash in just a few minutes.

With parallelisation and deferred forkserver / persistent mode you will get the corresponding performance increases
which should find a crash in a few minutes even without the seed.

Here's a sample crashing input after being put through the afl-tmin grinder:

    0000000000000000000=
    000000000000000000000000000000000000000000=

(now you can see why the sample seed was somewhat unfair)

Try taking a look at the source code or running the program under gdb to see why this leads to a buffer overflow.

If you put some of your crashing inputs through afl-tmin, do you get the same input as above each time? If you don't, it
means that afl-tmin couldn't shrink the inputs any further without leading to a crash, yet they are still different - to
understand why you'll have to turn to the source.

Vulnerability source: CVE-1999-0206, https://samate.nist.gov/SRD/view_testcase.php?tID=1301
