This bug in gnulib was recently found using AFL against the coreutils date program.

Check out the date manpage - it takes input from the command line, date syscall, environment variables, and optionally files.

Here we want to fuzz an environment variable - refer to the manpage to identify it.

The basic process for compiling date with afl is familiar. Be sure to checkout the specified version or earlier.

	$ git clone https://git.savannah.gnu.org/git/coreutils.git
	$ cd coreutils
	$ git checkout f4570a9
	$ sudo apt install autopoint bison gperf autoconf
	$ ./bootstrap
	$ <CC=...> ./configure
	$ <any AFL compile options> make -j

	$ ./src/date
	Mon Jul  3 08:11:23 PDT 2017
