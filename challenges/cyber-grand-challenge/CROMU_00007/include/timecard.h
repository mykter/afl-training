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

#ifndef TIMECARD_H
#define TIMECARD_H

#define WEEKS_IN_A_YEAR 52
#define EMPLOYEE_NAME_LEN 64
#define NUMBER_OF_EMPLOYEES 50
#define PAYROLL_TAX_RATE .0765

typedef struct money_s
{
	int dollars;
	int cents;
} money, *pmoney;

typedef struct time_s
{
	int hours;
	int minutes;
} time, *ptime;

// A function pointer will be used to calculate overtime differently for exempt vs non-exempt employees 
typedef void (*overtime_calc)(pmoney, pmoney, ptime);

// A payroll struct is used to hold time and pay information for one week
typedef struct payroll_s{
	time 	standardtime;
	time 	overtime;
	money 	standardpay;
	money 	overtimepay;
	money 	payroll_tax;
#ifdef PATCHED
	char paycheck[20];
#else
	char paycheck[12];
#endif
	overtime_calc calculate_overtime;
} payroll, *ppayroll;

// The employee structure holds various employee information as well as a payroll
// record for each week of the year. 
typedef struct employee_s{
	char 	name[EMPLOYEE_NAME_LEN];
	int 	id;
	money 	wage;
	int 	exempt;
	payroll paychecks[WEEKS_IN_A_YEAR];
} employee, *pemployee;

void atom(pmoney amount, char *str);
void mtoa(char *str, pmoney amount);
void atoh(ptime t, char *str);
void htoa(char *str, ptime t);
void initialize_employee(pemployee empl);
void add_time(ptime t, int hours, int minutes);
void add_money(pmoney dest, float money);
void add_pay(pmoney pay, pmoney rate, ptime timeworked);
void log_hours(ppayroll paycheck, char *hours);
void log_overtime_hours(ppayroll paycheck, char *hours);
void calculate_standardpay(pmoney pay, pmoney wage, ptime timeworked);
void calculate_totalpay(ppayroll paycheck);
void exempt_overtime(pmoney pay, pmoney wage, ptime timeworked);
void nonexempt_overtime(pmoney pay, pmoney wage, ptime timeworked);
int get_key_value(char *inbuf, size_t length, char **key, char **value);
void process_key_value(pemployee empl, char *key, char *value, int *week);
void merge_employee_records(pemployee empl, pemployee temp);
void process_query(int query, employee employee_list[], pemployee temp, int week);
void output_paycheck(pemployee empl, int week);

int equals(char *a, char *b);

// Read status codes
#define READ_ERROR 			-1
#define NEWLINE_RECEIVED 	 1
#define KEY_VALUE_RECEIVED	 2
#define OTHER_INPUT_RECEIVED 3

// Query codes
#define QUERY_NONE		0
#define QUERY_ONE		1
#define QUERY_ALL		2
#define QUERY_WEEK		3
#define QUERY_WEEK_ALL 	4

#endif //TIMECARD_H
