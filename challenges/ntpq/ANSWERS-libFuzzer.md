# Compiling

The configure script will fail if you try and compile with `-fsanitize=fuzzer`, because it fails its sanity test of "can
you compile a program". Instead, we configure with `-fsanitize-fuzzer-no-link`, but then manually override the CFLAGS
used during actual compilation.

1.  Configure the whole application with the `fuzzer-no-link` sanitizer:
    `CC=clang CFLAGS="-fsanitize=address,fuzzer-no-link -g" ./configure`

2.  Build the ntpq binary using libFuzzer's main: `make clean && make -C ntpq CFLAGS="-fsanitize=address,fuzzer -g"`
    (note if you haven't replaced the existing main function yet, this will fail to link)

# Running

From the ntpq sub-directory: `./ntpq -dict /home/fuzzer/workshop/challenges/ntpq/ntpq.dict corpus`

# Harness

As per the AFL guidance, replace ntpq's existing main but with the usual LLVMFuzzerTestOneInput function.

The tool will spit a lot of errors to stderr; you can replace all of this with /dev/null to make the fuzzer output more
readable (and faster).

For example:

```c
FILE *fdevnull;

int LLVMFuzzerTestOneInput(uint8_t *data, size_t size)
{
    if (fdevnull == NULL)
    {
        fdevnull = fopen("/dev/null", "w");
    }
	if (size > 3)
	{
		int datatype = data[--size];
		int status = data[--size];
		cookedprint(datatype, size, data, status, 1, fdevnull);
	}
	return 0;
}
```

And then search and replace `stderr` for `fdevnull` across the whole file.

When you want to get coverage data, you'll need to provide your own `main` function that can take input from stdin or a
file and call LLVMFuzzerTestOneInput on the contents. For example:

```c
#ifdef FUZZING_BUILD_MODE_UNSAFE_FOR_PRODUCTION
// no main
#elseif FUZZING_COVERAGE_MODE
int main(int argc, char **argv)
{

	char data[4096] = {0};
	int size = read(0, data, 4096);
	return LLVMFuzzerTestOneInput(data, size);
}
#else
<existing main definition>
#endif
```
