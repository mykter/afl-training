ntpq is a utility included as part of the NTP Reference Implementation suite of tools. It queries a server (e.g. ntpd)
and provides information to the user.

See if you can find CVE-2009-0159 in ntpq using afl.

If you aren't using the Docker environment, v4.2.2 can be obtained from
https://www.eecis.udel.edu/~ntp/ntp_spool/ntp4/ntp-4.2/ntp-4.2.2.tar.gz

Otherwise, it is already present in the ntp-4.2.2 dir.

You can see a writeup of the bug and the fix here:
https://xorl.wordpress.com/2009/04/13/cve-2009-0159-ntp-remote-stack-overflow/

Use a coverage checker to see what parts of the target function you're exercising, then consider how to expand that
coverage.

Repeat the exercise on version 4.2.8p10, which has a fix.

Compilation tip: compile ntpq by itself with `make -C ntpq`; trying to make everything includes ntpd, which is slower to
compile and depending on how you're trying to compile it, might fail.
