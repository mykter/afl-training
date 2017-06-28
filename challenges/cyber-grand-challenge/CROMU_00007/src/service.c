/*

Author: Debbie Nuttall <debbie@cromulence.com>

Copyright (c) 2014 Cromulence LLC

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.

*/

#include <stdio.h>
#include <stdlib.h>
#include <ctype.h>
#include "timecard.h"

int main(void) 
{
	int i;
	char inbuf[80];
	char *key, *value;
	int status;
	employee temporary_record;
	int week;
	int blank_line = 1;
	int query = 0;
	// Store of all employee data
	employee employee_list[NUMBER_OF_EMPLOYEES];

	for( i=0; i<NUMBER_OF_EMPLOYEES; i++)
	{
		initialize_employee(&employee_list[i]);
	}

	initialize_employee(&temporary_record);
	week = 0;

	// Process input and respond to queries until a blank line is received
	while(1)
	{
		// Read input from network in the form of key-value pairs
		status = get_key_value((char *)&inbuf, sizeof(inbuf), &key, &value);
		switch (status)
		{
			case READ_ERROR:
				printf("ERROR: on receive\n");
				exit(1);
				break;
			case NEWLINE_RECEIVED:
			// Process the input line by merging the temporary record with the global record
				if (blank_line == 1)
				{
					printf("Exiting\n");
					exit(0);
				}
				if (query > 0)
				{
					process_query(query, (employee *)&employee_list, &temporary_record, week);
				}
				else if ((temporary_record.id >= 0) && (temporary_record.id < NUMBER_OF_EMPLOYEES))
				{
					merge_employee_records(&employee_list[temporary_record.id], &temporary_record);
				}
				// Get ready to process a new line
				initialize_employee(&temporary_record);
				week = 0;
				blank_line = 1;
				query = 0;
				break;			
			case KEY_VALUE_RECEIVED:
				// If this is a query line remember it for processing later
				if ((blank_line == 1) && (equals(key, "query")))
				{
					if (equals(value, "all"))
						query = QUERY_ALL;
					else if (equals(value, "one"))
						query = QUERY_ONE;
					else if (equals(value, "week"))
						query = QUERY_WEEK;
					else if (equals(value, "week_all"))
						query = QUERY_WEEK_ALL;
					else
						query = 0;
				}
				blank_line = 0;
				// Add the key_value information to a temporary record
				process_key_value(&temporary_record, key, value, &week);
				break;
			case OTHER_INPUT_RECEIVED:
			default:
				printf("ERROR: invalid input\n");
				exit(1);
				break;
		}
	}
	
}
