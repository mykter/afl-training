# Enable debugging and suppress pesky warnings
CFLAGS ?= -g -w

all:	vulnerable

clean:
	rm -f vulnerable

vulnerable: vulnerable.c
	${CC} ${CFLAGS} vulnerable.c -o vulnerable
