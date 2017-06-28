#include <string.h>
#include <stdio.h>
#include <unistd.h>
#include <stdlib.h>

int main()
{
	char input[100] = {0};
	char *out;

	// Slurp input
	if (read(STDIN_FILENO, input, 100) < 0) {
		fprintf(stderr, "Couldn't read stdin.\n");
	}
	if(input[0] == 'c') { 
		// count characters
		out = malloc(sizeof(input) - 1 + 3); // enough space for 2 digits + a space + input-1 chars
		sprintf(out, "%lu ", strlen(input) - 1);
		strcat(out, input+1);
		printf("%s", out);
		free(out);
	} else if ((input[0] == 'e') && (input[1] == 'c')) {
		// echo input
		printf(input + 2);
	} else if (strncmp(input, "head", 4) == 0) {
		// head
		if (strlen(input) > 5) {
			input[input[4]] = '\0'; // truncate string at specified position
			printf("%s", input+4);
		} else {
			fprintf(stderr, "head input was too small\n");
		}
	} else if (strcmp(input, "surprise!") == 0) {
		// easter egg!
		*(char *)1=2;
	} else {
		fprintf(stderr, "Tell me what to doooo!\n");
	}
}
