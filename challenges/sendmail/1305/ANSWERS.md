This vulnerability is not a quick catch with the afl approach.

Example crashing input:

    python -c 'print 60*("\\"+"\xff")'  | ./prescan-bad

In a short test run my instance produced an input file with a large number of backlashes. This is on the right path: it
will have been found because the hitcount for a particular pair of basic blocks is higher with each backslash. It also
found the 0xFF character as being interesting - various test cases have runs of it.

However to actually crash the program, you need a long combo of "\<0xff>" pairs, or similar (I haven't studied the
vulnerability). afl will eventually find this through splicing and mutating the test cases mentioned above - whether it
finds it quickly depends on whether the instrumentation guides it towards it.

It will hit it at some point, but it's not clear how long it will take. This is a good example of how it can be hard to
determine when to stop fuzzing. It will go through a lot of cycles without finding any new paths, but each of those
cycles are very quick. A code coverage tool will likely say we've hit every line of code.

Deferred initialization makes a significant different - just under 1.5x performance improvement for me. The obvious
place to put it is immediately before the call to read().

Persistent mode makes a massive difference. This is a very fast target (it's just scanning a 50char string), so the
overhead of forking represents a very large proportion of the total time. I get a ~4x performance improvement on top of
deferred mode. BUT - is it safe? Is it actually representative testing after the first loop? There are a few global
variables - you would have to check the code to make sure these weren't maintaining state that affected the process
between calls to `parseaddr()`. If you were fuzzing this for real and found that this wasn't safe, then given the
performance improvement it might be worth it modifying the code to make it stateless.

This is the output from running afl-tmin on the example crashing input from above:

    000000000000000000000000000000000000000000000\�\�\�\�\�\�\�\�\�\�\�\�\�\�\�\�\�\�\�\�\�

As you can see, this is markedly simpler than the original demo - debugging a crash with this input would likely be a
lot easier.

Vulnerability source: CVE-2003-0161, https://samate.nist.gov/SRD/view_testcase.php?tID=1305
