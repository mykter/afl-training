#include <stdlib.h>
#include <stdio.h>
#include <string.h>

void lib_echo(char *data, size_t len){
	if(strlen(data) == 0) {
		return;
	}
	char *buf = calloc(1, len);
	strncpy(buf, data, len);
	printf(buf);
	free(buf);
}

int  lib_mul(int x, int y){
	if(x%2 == 0) {
		return y << x;
	} else if (y%2 == 0) {
		return x << y;
	} else if (x == 0) {
		return 0;
	} else if (y == 0) {
		return 0;
	} else {
		return x * y;
	}
}
