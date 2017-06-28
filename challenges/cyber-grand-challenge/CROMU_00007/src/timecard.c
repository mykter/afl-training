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
#include <string.h>
#include <unistd.h>
#include <ctype.h>
#include "timecard.h"

int equals(char *a, char *b)
{
	return !strcmp(a,b);
}

// Converts a char string into a money struct. 
void atom(pmoney amount, char *str)
{
	size_t i = 0;
	int negative = 0;
	amount->cents = 0;
	amount->dollars = 0;

	if(str[i] == '-')
	{
		negative = 1;
		i++;
	}
	for(; isdigit(str[i]); i++)
	{
		amount->dollars = (amount->dollars * 10) + (str[i] - '0');
	}
	if (str[i] == '.')
	{
		if (isdigit(str[i+1]))
		{
			amount->cents = 10 * (str[i+1] - '0');
			if (isdigit(str[i+2]))
			{
				amount->cents += (str[i+2] - '0');
			}
		}
	}
	if (negative)
	{
		amount->dollars = -amount->dollars;
	}
}

// Converts a money struct into a character string
// str should point to a buffer at least 19 bytes long to handle max value
// $-2,147,483,647.00
void mtoa(char *str, pmoney amount)
{
	char buffer[20];
	size_t pos = 0;
	size_t outpos;
	int value = amount->cents;
	int negative = 0;

	if (value < 0)
	{
		value  = -value;
	}
	buffer[pos++] = '0' + (value % 10);
	value /= 10;
	buffer[pos++] = '0' + (value % 10);
	buffer[pos++] = '.';
	value = amount->dollars;
	if (value < 0)
	{
		value = -value;
		negative = 1;
	}
	do
	{
		if ((pos == 6)||(pos == 10)||(pos==14))
		{
			buffer[pos++] = ',';
		}
		buffer[pos++] = '0' + value % 10;
		value /= 10;
	} while((value > 0) && (pos < 63));
	if (negative)
	{
		buffer[pos++] = '-';
	}
	outpos = 1;
	str[0] = '$';
	while(outpos <= pos)
	{
		str[outpos] = buffer[pos - outpos];
		outpos++;
	}
	str[outpos] = '\0';
}

// Converts a character string into a time struct
void atoh(ptime t, char *str)
{
	int pos = 0;
	int value = 0;
	t->hours = 0;
	t->minutes = 0;
	while(isdigit(str[pos]))
	{
		value = (value * 10) + (str[pos] - '0');
		pos++;
	}
	if(str[pos] == 'h')
	{
		t->hours = value;
		pos++;
	}
	value = 0;
	while(isdigit(str[pos]))
	{
		value = (value * 10) + (str[pos] - '0');
		pos++;
	}
	if (str[pos] == 'm')
	{
		t->minutes = value;
	}
	if (t->hours > 24)
	{
		t->hours = 24;
	}
	if (t->minutes > 59)
	{
		t->minutes = 59;
	}
}


// Converts a time struct into a character string
// str should point to a buffer at least 15 bytes long to handle the max value
// h2147483647m60\0
void htoa(char *str, ptime t)
{
	char buffer[15];
	int pos=0;
	int outpos = 0;
	int value;
	value = t->minutes;
	if (value > 60)
	{
		str[outpos] = '\0';
		return;
	}
	buffer[pos++] = 'm';
	do
	{
		buffer[pos++] = '0' + value % 10;
		value /= 10;
	} while(value > 0);
	buffer[pos++] = 'h';
	value = t->hours;
	do
	{
		buffer[pos++] = '0' + value % 10;
		value /= 10;
	} while(value > 0);
	while(outpos < pos)
	{
		str[outpos] = buffer[pos - outpos - 1];
		outpos++;
	}
	str[outpos] = '\0';
}

// Initialize an employee structure 
void initialize_employee(pemployee empl)
{
	int i,j;
	empl->name[0] = '\0';
	// An id value of -1 is used elsewhere to determine whether an employee has been registered
	empl->id = -1;
	empl->wage.dollars = 0;
	empl->wage.cents = 0;
	empl->exempt = 0;
	for (i=0; i<WEEKS_IN_A_YEAR; i++)
	{
		empl->paychecks[i].standardtime.hours = 0;
		empl->paychecks[i].standardtime.minutes = 0;
		empl->paychecks[i].overtime.hours = 0;
		empl->paychecks[i].overtime.minutes = 0;
		empl->paychecks[i].calculate_overtime = &exempt_overtime;
		empl->paychecks[i].payroll_tax.dollars = 0;
		empl->paychecks[i].payroll_tax.cents = 0;
		memset(empl->paychecks[i].paycheck, 0, sizeof(empl->paychecks[i].paycheck));
	}
}

// Adds hours and minutes to a time struct
void add_time(ptime t, int hours, int minutes)
{
	t->hours += hours + (t->minutes + minutes) / 60;
	t->minutes = (t->minutes + minutes) % 60;
}

// Rounds time to the nearest quarter hour
void round_minutes(ptime t)
{

	if (t->minutes < 8)
	{
		t->minutes = 0;
	}
	else if (t->minutes < 23)
	{
		t->minutes = 15;
	}
	else if (t->minutes < 38)
	{
		t->minutes = 30;
	}
	else if (t->minutes < 53)
	{
		t->minutes = 45;
	}
	else
	{
		t->hours += 1;
		t->minutes = 0;
	}
}

// Adds money to a money struct, rounding cents to nearest 1c.
void add_money(pmoney dest, float money)
{
	float total;
	// Convert dest amount back to float
	total = (float)dest->dollars + ((float)dest->cents / 100.0);
	// Add money and round up/down to nearest cent
	total = total + money;

	if (total > 0)
	{
		total = total + 0.005;
	}
	else
	{
		total = total - 0.005;
	}
	// Convert back to money struct
	dest->dollars = (int)total;
	dest->cents = (int)((total - (float)dest->dollars) * 100);

}

// Calculate pay and add to a money struct
void add_pay(pmoney pay, pmoney rate, ptime timeworked)
{
	float float_rate = (float)rate->dollars + ((float)rate->cents / 100.0);
	float minute_rate = float_rate / 60.0;

	add_money(pay, timeworked->hours * float_rate + timeworked->minutes * minute_rate);
}

// Adds time to an employee's payroll 
void log_hours(ppayroll paycheck, char *hours)
{
	time t;
	atoh(&t, hours);
	// Round time worked to the nearest quarter hour 
	round_minutes(&t);

	// Anything over 8 hours is considered overtime
	if (t.hours >= 8) 
	{
		add_time(&paycheck->standardtime, 8, 0);
		add_time(&paycheck->overtime, t.hours - 8, t.minutes);
	}
	else 
	{
		add_time(&paycheck->standardtime, t.hours, t.minutes);
	}
}

// Adds overtime to an employee's payroll (used for weekends)
void log_overtime_hours(ppayroll paycheck, char *hours)
{
	time t;
	atoh(&t, hours);
	// Round time worked to the nearest quarter hour 
	round_minutes(&t);
	add_time(&paycheck->overtime, t.hours, t.minutes);
}

// Calculates standard pay based on wage and timeworked 
// Used to populate the standardpay field of an employee's weekly payroll
void calculate_standardpay(pmoney pay, pmoney wage, ptime timeworked)
{
	pay->dollars = 0;
	pay->cents = 0;
	
	add_pay(pay, wage, timeworked);
}

// Calculates total pay, subtracts taxes withheld, and outputs total to the 
// paycheck field of an employee's weekly payroll
void calculate_totalpay(ppayroll paycheck)
{
	money total;
	total.dollars = 0;
	total.cents = 0;
	add_money(&total, (float)paycheck->standardpay.dollars + (float)paycheck->standardpay.cents / 100.0);
	add_money(&total, (float)paycheck->overtimepay.dollars + (float)paycheck->overtimepay.cents / 100.0);
	add_money(&total, -1.0 * (((float)total.dollars + (float)total.cents/100.0) * PAYROLL_TAX_RATE) );

	mtoa((char *)&paycheck->paycheck, &total);
}

// Calculates overtime pay for an exempt employee based on wage and overtime worked 
// Used to populate the overtimepay field of an employee's weekly payroll
void exempt_overtime(pmoney pay, pmoney wage, ptime timeworked)
{
	money overtime_wage;
	overtime_wage.dollars = 0;
	overtime_wage.cents = 0;
	pay->dollars = 0;
	pay->cents = 0;

	add_money(&overtime_wage, ((float)wage->dollars + ((float)wage->cents/100.0)) * 1.0);
	add_pay(pay, &overtime_wage, timeworked);
}

// Calculates overtime pay for an nonexempt employee based on wage and overtime worked 
// Used to populate the overtimepay field of an employee's weekly payroll
void nonexempt_overtime(pmoney pay, pmoney rate, ptime timeworked)
{
	pay->dollars = 0;
	pay->cents = 0;

	float float_rate = ((float)rate->dollars + ((float)rate->cents / 100.0)) * 1.5;
	float minute_rate = (float_rate / 60.0);
	add_money(pay, timeworked->hours * float_rate + timeworked->minutes * minute_rate );
}

// Reads input from the network in the format key-value` and parses the 
// key/value pair by searching for delimiters and replacing with NULL
// Returns status code:
// READ_ERROR - Indicates a problem receiving from the network
// NEWLINE_RECEIVED - Indicates a newline character was received
// KEY_VALUE_RECEIVED - Indicates a valid key/value pair received
// 		**key points to start of null terminated key string
//		**value points to start of null terminated value string
// OTHER_INPUT_RECEIVED - Input received that does not match previous options
int get_key_value(char *inbuf, size_t length, char **key, char **value)
{
	int count;
	int pos = 0;
	char c;
	char buffer[80];

	while ( (count = read(STDIN_FILENO, &c, 1)) == 1) {
		inbuf[pos++] = c;
		if(c == '\n')
		{
			return NEWLINE_RECEIVED;
		} else if(c == '`') {
			char *search = inbuf + pos - 1;
			*search = '\0';
			while(search > inbuf)
			{
				search--;
				if (*search == '-')
				{
					*key = inbuf;
					*search = '\0';
					*value = search + 1;
					return KEY_VALUE_RECEIVED;
				}
			}

		}
	}
	if (count != 0) {
		return READ_ERROR;
	}
	return OTHER_INPUT_RECEIVED; // EOF

}

// Checks the key/value pair for known keys. Parses values as necessary
// to populate an employee structure. 
// The week field is used to maintain stateful awareness of the week when logging hours. 
// See merge_employee_records for proper usage of key/value pairs
void process_key_value(pemployee empl, char *key, char *value, int *week)
{
	if (equals(key, "employee_id"))
	{
		empl->id = atoi(value);
	}
	if (equals(key, "employee_name"))
	{
		strncpy(empl->name, value, EMPLOYEE_NAME_LEN);
	}
	if (equals(key, "wage"))
	{
		atom(&empl->wage, value);
		// Cap maximum wage at $75/hr
		if(empl->wage.dollars >= 75)
		{
			empl->wage.dollars = 75;
			empl->wage.cents = 0;
		}
	}
	if (equals(key, "exempt"))
	{
		if (equals(value, "yes"))
		{
			empl->exempt = 1;
		}
	}
	if (equals(key, "week"))
	{
		*week = atoi(value);
		if ((*week < 0) || (*week >= 52))
		{
			*week = 0;
		}
	}
	if (equals(key, "monday") || 
		equals(key, "tuesday") || 
		equals(key, "wednesday") || 
		equals(key, "thursday") || 
		equals(key, "friday"))
	{
		log_hours(&empl->paychecks[*week], value);

	}
	if (equals(key, "saturday") ||
		equals(key, "sunday"))
	{
		log_overtime_hours(&empl->paychecks[*week], value);
	}
}

// Merges the contents of one employee struct with another. 
// Used to add a line of input to the global employee record. 
// 
// When the line contains a name field, the line is treated as a
// registration command which can have the following fields:
//	employee_id-#`
//	employee_name-ASCII`
//	wage-#.##`
//	exempt-(yes|no)`
// Otherwise it is a time log which can have the following fields:
//	employee_id-#`
//	week-#`
//	monday-#h#m`
//	... 
//	sunday-#h#m`
void merge_employee_records(pemployee empl, pemployee temp)
{
	// Only populate employee record the first time a name is given
	if ((empl->id == -1) && (temp->name[0] != '\0'))
	{
		strncpy(empl->name, temp->name, EMPLOYEE_NAME_LEN);
		empl->id = temp->id;
		empl->wage.dollars = temp->wage.dollars;
		empl->wage.cents = temp->wage.cents;
		empl->exempt = temp->exempt;
	}
	// If employee does not yet exist and no name is given, ignore this line
	else if (empl->id == -1)
	{
		return;
	}
	// Otherwise, treat line as a time log
	else
	{
		int week;
		for (week = 0; week < WEEKS_IN_A_YEAR; week++)
		{
			add_time(&empl->paychecks[week].standardtime, temp->paychecks[week].standardtime.hours, 
				temp->paychecks[week].standardtime.minutes);
			add_time(&empl->paychecks[week].overtime, temp->paychecks[week].overtime.hours, 
				temp->paychecks[week].overtime.minutes);

			if (empl->exempt == 0)
			{
				empl->paychecks[week].calculate_overtime = &nonexempt_overtime;
			}

			calculate_standardpay(&empl->paychecks[week].standardpay, &empl->wage, &empl->paychecks[week].standardtime);
			empl->paychecks[week].calculate_overtime(&empl->paychecks[week].overtimepay, &empl->wage, &empl->paychecks[week].overtime);
		}
	}
}

// Process the query based on type
// 	QUERY_ALL - outputs all valid employee records 
// 	QUERY_ONE - outputs one employee record
// 	QUERY_WEEK - outputs one week of one employees record		
void process_query(int query, employee *employee_list, pemployee temp, int week)
{
	switch (query)
	{
		case QUERY_ALL:
		{
			int employee, week;
			for (employee=0; employee < NUMBER_OF_EMPLOYEES; employee++)
			{
				// Ignore unregistered slots
				if (employee_list[employee].id == -1)
					continue;
				
				for (week=0; week < WEEKS_IN_A_YEAR; week++)
				{
					output_paycheck(&employee_list[employee], week);
				}
			
			}
			break;
		}
		case QUERY_ONE:
		{

			int week;
			if ((temp->id >= 0) && (temp->id < NUMBER_OF_EMPLOYEES))
			{
				// Ignore unregistered slots
				if (employee_list[temp->id].id == -1)
					break;

				for (week = 0; week < WEEKS_IN_A_YEAR; week++)
				{
					output_paycheck(&employee_list[temp->id], week);
				}
			}
			break;
		}
		case QUERY_WEEK:
#ifndef PATCHED
			// Ignore unregistered slots
			if (employee_list[temp->id].id == -1)
				break;
#endif 

			if ((temp->id >= 0) && (temp->id < NUMBER_OF_EMPLOYEES)
				&& (week >= 0) && (week < WEEKS_IN_A_YEAR))
			{
#ifdef PATCHED
				// Ignore unregistered slots
				if (employee_list[temp->id].id == -1)
					break;
#endif
				output_paycheck(&employee_list[temp->id], week);
			}
			break;
		case QUERY_WEEK_ALL:
		{
			int employee;
			if ((week >= 0) && (week < WEEKS_IN_A_YEAR))
			{
				for (employee=0; employee<NUMBER_OF_EMPLOYEES; employee++)
				{
					// Ignore unregistered slots
					if (employee_list[employee].id == -1)
						continue;

					output_paycheck(&employee_list[employee], week);
				}
			}
			break;
		}
		case QUERY_NONE:
		default:
			break;
	}			

}

// Output an employee's payroll for one week
void output_paycheck(pemployee empl, int week)
{
	char outbuf[20];
	// Output result
	printf("employee-");
	printf("%s",empl->name);
	printf("`week-%d", week);
	printf("`standardtime-");
	htoa((char *)&outbuf, &empl->paychecks[week].standardtime);
	printf("%s",outbuf);
	printf("`overtime-");
	htoa((char *)&outbuf, &empl->paychecks[week].overtime);
	printf("%s",outbuf);
	printf("`standardpay-");
	mtoa((char *)&outbuf, &empl->paychecks[week].standardpay);
	printf("%s",outbuf);
	printf("`overtimepay-");
	mtoa((char *)&outbuf, &empl->paychecks[week].overtimepay);
	printf("%s",outbuf);
	printf("`netpay-");
	calculate_totalpay(&empl->paychecks[week]);
	printf("%s",empl->paychecks[week].paycheck);
	printf("`\n");
}


// These functions are to be used by the poller to aleviate issues with floating point implementation differences 
// and rounding errors

float c_standardpay(int hours, int minutes, int dollars, int cents)
{
	money total;
	float float_rate = (float)dollars + ((float)cents / 100.0);
	float minute_rate = (float_rate / 60.0);
	total.dollars = 0;
	total.cents = 0;
	add_money(&total, hours * float_rate + minutes * minute_rate);
	return (float)total.dollars + ((float)total.cents / 100.0);
}

float c_overtimepay(int hours, int minutes, int dollars, int cents, int exempt)
{
	money total;
	float float_rate;
	float minute_rate;
	if (exempt == 0)
	{
		float_rate = ((float)dollars + ((float)cents / 100.0)) * 1.5;
	}
	else
	{
		float_rate = (float)dollars + ((float)cents / 100.0); ;
	}
	total.dollars = 0;
	total.cents = 0;
	minute_rate = (float_rate / 60.0);
	add_money(&total, hours * float_rate + minutes * minute_rate);
	return (float)total.dollars + ((float)total.cents / 100.0);
}

float c_netpay(int hours, int minutes, int overhours, int overminutes, int dollars, int cents, int exempt)
{
	money total, rate, standard, overtime;
	time standardtime, overtimetime;
	total.dollars = 0;
	total.cents = 0;
	rate.dollars = dollars;
	rate.cents = cents;
	standardtime.hours = hours;
	standardtime.minutes = minutes;
	calculate_standardpay(&standard, &rate, &standardtime);
	overtimetime.hours = overhours;
	overtimetime.minutes = overminutes;
	if (exempt == 0)
		nonexempt_overtime(&overtime, &rate, &overtimetime);
	else
		exempt_overtime(&overtime, &rate, &overtimetime);
	add_money(&total, (float)standard.dollars + (float)standard.cents/100.0);
	add_money(&total, (float)overtime.dollars + (float)overtime.cents/100.0);
	add_money(&total, -1.0 * (((float)total.dollars + (float)total.cents/100.0)*PAYROLL_TAX_RATE));
	return (float)total.dollars + ((float)total.cents / 100.0);
}

