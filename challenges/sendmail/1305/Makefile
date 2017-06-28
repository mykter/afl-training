CFLAGS ?= -g

all:	prescan-bad

clean:	
	rm -f *-bad *-ok

prescan-bad: prescan-overflow-bad.c
	$(CC) $(CFLAGS) -g -o prescan-bad prescan-overflow-bad.c


