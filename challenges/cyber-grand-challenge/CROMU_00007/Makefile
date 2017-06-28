POLLS_RELEASE_SEED= -802674260
POLLS_TESTING_SEED= -690922790
AUTHOR_ID  = CROMU
SERVICE_ID = 00007
CFLAGS     = -O0 -g -Werror -Wno-overlength-strings -Wno-packed

POLLS_RELEASE_COUNT = 100
POLLS_TESTING_COUNT = 900

all: cromu_00007

clean:
	rm -f cromu_00007

cromu_00007: src/timecard.c src/service.c
	${CC} ${CFLAGS} -o cromu_00007 src/timecard.c src/service.c -I ./include
  
