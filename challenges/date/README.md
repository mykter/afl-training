This bug in gnulib was recently found using AFL against the coreutils date program.

Check out the date manpage - it takes input from the command line, date syscall, environment variables, and optionally
files.

Here we want to fuzz an environment variable - refer to the manpage to identify it.

The basic process for compiling date with afl is familiar. Be sure to checkout the specified version or earlier.

```shell
	$ git submodule init && git submodule update
	$ sudo apt install autopoint bison gperf autoconf # already installed in the container environment
	$ cd coreutils
	$ ./bootstrap # may finish with some errors to do with 'po' files, which can be ignored

	# this old version doesn't work with modern compilers, we need to apply a patch
	$ pushd gnulib && wget https://src.fedoraproject.org/rpms/coreutils/raw/f28/f/coreutils-8.29-gnulib-fflush.patch && patch -p1 < coreutils-8.29-gnulib-fflush.patch && popd

	$ <CC=...> ./configure
	$ <any AFL compile options> make -j src/date # only compile the date binary - saves a lot of time

	$ ./src/date
	Mon Jul  3 08:11:23 PDT 2017
```
