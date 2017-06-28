CFLAGS ?= -g

all:	m1-bad

clean:	
	rm -f *-bad *-ok

m1-bad: mime1-bad.c main.c 
	${CC} ${CFLAGS} -o m1-bad mime1-bad.c main.c -I .

