This bug in gnulib was recently found using AFL against the coreutils date program.

Check out the date manpage - it takes input from the command line, date syscall, environment variables, and optionally
files.

Here we want to fuzz an environment variable - refer to the manpage to identify it.

The basic process for compiling date with afl is familiar.

If you aren't in the workshop container environment, grab coreutils and some build dependencies:

```shell
	# not required in the workshop environment
	$ git submodule init && git submodule update
	$ sudo apt install autopoint bison gperf autoconf
```

Building this old version of coreutils is quite fragile - take care to follow these instructions precisely!

```shell
	$ cd coreutils
	$ ./bootstrap # may finish with some errors to do with 'po' files, which can be ignored

	# this old version doesn't work with modern compilers, we need to apply a patch
	# this patch can get overwritten by certain make targets - if you get an error during compilation about fseeko.c, try re-applying the patch
	$ patch --follow-symlinks -p1 < ../coreutils-8.29-gnulib-fflush.patch

	$ <CC=...> ./configure
	$ make # this will build everything. you can try `make src/date`, but the makefile doesn't specify all of the dependencies properly so this will probably fail until you've built everything once

	$ ./src/date
	Mon Jul  3 08:11:23 PDT 2017
```
