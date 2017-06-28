
/*

MIT Copyright Notice

Copyright 2003 M.I.T.

Permission is hereby granted, without written agreement or royalty fee, to use, 
copy, modify, and distribute this software and its documentation for any 
purpose, provided that the above copyright notice and the following three 
paragraphs appear in all copies of this software.

IN NO EVENT SHALL M.I.T. BE LIABLE TO ANY PARTY FOR DIRECT, INDIRECT, SPECIAL, 
INCIDENTAL, OR CONSEQUENTIAL DAMAGES ARISING OUT OF THE USE OF THIS SOFTWARE 
AND ITS DOCUMENTATION, EVEN IF M.I.T. HAS BEEN ADVISED OF THE POSSIBILITY OF 
SUCH DAMANGE.

M.I.T. SPECIFICALLY DISCLAIMS ANY WARRANTIES INCLUDING, BUT NOT LIMITED TO 
THE IMPLIED WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, 
AND NON-INFRINGEMENT.

THE SOFTWARE IS PROVIDED ON AN "AS-IS" BASIS AND M.I.T. HAS NO OBLIGATION TO 
PROVIDE MAINTENANCE, SUPPORT, UPDATES, ENHANCEMENTS, OR MODIFICATIONS.

$Author: tleek $
$Date: 2004/01/05 17:27:45 $
$Header: /mnt/leo2/cvs/sabo/hist-040105/sendmail/s5/prescan-overflow-bad.c,v 1.1.1.1 2004/01/05 17:27:45 tleek Exp $



*/


/*

Sendmail Copyright Notice


Copyright (c) 1998-2003 Sendmail, Inc. and its suppliers.
     All rights reserved.
Copyright (c) 1983, 1995-1997 Eric P. Allman.  All rights reserved.
Copyright (c) 1988, 1993
     The Regents of the University of California.  All rights reserved.

By using this file, you agree to the terms and conditions set
forth in the LICENSE file which can be found at the top level of
the sendmail distribution.


$Author: tleek $
$Date: 2004/01/05 17:27:45 $
$Header: /mnt/leo2/cvs/sabo/hist-040105/sendmail/s5/prescan-overflow-bad.c,v 1.1.1.1 2004/01/05 17:27:45 tleek Exp $



*/


/*

<source>

*/

#include <stdio.h>
#include <sys/types.h>
#include <stdlib.h>
#include <string.h>
#include <ctype.h>

/* states and character types */
#define OPR		0	/* operator */
#define ATM		1	/* atom */
#define QST		2	/* in quoted string */
#define SPC		3	/* chewing up spaces */
#define ONE		4	/* pick up one character */
#define ILL		5	/* illegal character */

#define NSTATES	6	/* number of states */

#define MAXNAME		40		/* max length of a name */
#define MAXATOM		10		/* max atoms per address */
#define PSBUFSIZE       (MAXNAME + MAXATOM) 

#define MAXMACROID       0377            /* max macro id number */

#define TYPE		017	/* mask to select state type */

/* meta bits for table */
#define M		020	/* meta character; don't pass through */
#define B		040	/* cause a break */
#define MB		M|B	/* meta-break */

#define TRUE 1
#define FALSE 0

#define tTd(flag, level)	(tTdvect[flag] >= (u_char)level)
#define tTdlevel(flag)		(tTdvect[flag])

/* variables */
extern u_char	tTdvect[100];	/* trace vector */

typedef int bool;

#ifndef SIZE_T
# define SIZE_T		size_t
#endif /* ! SIZE_T */

typedef struct envelope	ENVELOPE;

ENVELOPE   *CurEnv;	/* envelope currently being processed */

/* This is a very simplified representation of the envelope structure of Sendmail */
struct envelope
{
	char		*e_to;		/* the target person */	
        ENVELOPE	*e_parent;	/* the message this one encloses */
        char		*e_macro[MAXMACROID + 1]; /* macro definitions */
};

/* Simplified address struct */
struct address
{
	char		*q_paddr;	/* the printname for the address */
	char		*q_user;	/* user name */
	char		*q_ruser;	/* real user name, or NULL if q_user */
	char		*q_host;	/* host name */
	char		*q_home;	/* home dir (local mailer only) */
	char		*q_fullname;	/* full name if known */
	struct address	*q_next;	/* chain */
	struct address	*q_alias;	/* address this results from */
	char		*q_owner;	/* owner of q_alias */
	struct address	*q_tchain;	/* temporary use chain */
	char		*q_orcpt;	/* ORCPT parameter from RCPT TO: line */
	char		*q_status;	/* status code for DSNs */
	char		*q_rstatus;	/* remote status message for DSNs */
	char		*q_statmta;	/* MTA generating q_rstatus */
	short		q_state;	/* address state, see below */
	short		q_specificity;	/* how "specific" this address is */
};

typedef struct address ADDRESS;


static short StateTab[NSTATES][NSTATES] =
{
   /*	oldst	chtype>	OPR	ATM	QST	SPC	ONE	ILL	*/
	/*OPR*/	{	OPR|B,	ATM|B,	QST|B,	SPC|MB,	ONE|B,	ILL|MB	},
	/*ATM*/	{	OPR|B,	ATM,	QST|B,	SPC|MB,	ONE|B,	ILL|MB	},
	/*QST*/	{	QST,	QST,	OPR,	QST,	QST,	QST	},
	/*SPC*/	{	OPR,	ATM,	QST,	SPC|M,	ONE,	ILL|MB	},
	/*ONE*/	{	OPR,	OPR,	OPR,	OPR,	OPR,	ILL|MB	},
	/*ILL*/	{	OPR|B,	ATM|B,	QST|B,	SPC|MB,	ONE|B,	ILL|M	},
};

/* token type table -- it gets modified with $o characters */
static u_char	TokTypeTab[256] =
{
    /*	nul soh stx etx eot enq ack bel  bs  ht  nl  vt  np  cr  so  si   */
	ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM, ATM,SPC,SPC,SPC,SPC,SPC,ATM,ATM,
    /*	dle dc1 dc2 dc3 dc4 nak syn etb  can em  sub esc fs  gs  rs  us   */
	ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM, ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM,
    /*  sp  !   "   #   $   %   &   '    (   )   *   +   ,   -   .   /    */
	SPC,ATM,QST,ATM,ATM,ATM,ATM,ATM, SPC,SPC,ATM,ATM,ATM,ATM,ATM,ATM,
    /*	0   1   2   3   4   5   6   7    8   9   :   ;   <   =   >   ?    */
	ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM, ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM,
    /*	@   A   B   C   D   E   F   G    H   I   J   K   L   M   N   O    */
	ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM, ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM,
    /*  P   Q   R   S   T   U   V   W    X   Y   Z   [   \   ]   ^   _    */
	ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM, ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM,
    /*	`   a   b   c   d   e   f   g    h   i   j   k   l   m   n   o    */
	ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM, ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM,
    /*  p   q   r   s   t   u   v   w    x   y   z   {   |   }   ~   del  */
	ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM, ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM,

    /*	nul soh stx etx eot enq ack bel  bs  ht  nl  vt  np  cr  so  si   */
	OPR,OPR,ONE,OPR,OPR,OPR,OPR,OPR, OPR,OPR,OPR,OPR,OPR,OPR,OPR,OPR,
    /*	dle dc1 dc2 dc3 dc4 nak syn etb  can em  sub esc fs  gs  rs  us   */
	OPR,OPR,OPR,ONE,ONE,ONE,OPR,OPR, OPR,OPR,OPR,OPR,OPR,OPR,OPR,OPR,
    /*  sp  !   "   #   $   %   &   '    (   )   *   +   ,   -   .   /    */
	ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM, ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM,
    /*	0   1   2   3   4   5   6   7    8   9   :   ;   <   =   >   ?    */
	ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM, ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM,
    /*	@   A   B   C   D   E   F   G    H   I   J   K   L   M   N   O    */
	ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM, ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM,
    /*  P   Q   R   S   T   U   V   W    X   Y   Z   [   \   ]   ^   _    */
	ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM, ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM,
    /*	`   a   b   c   d   e   f   g    h   i   j   k   l   m   n   o    */
	ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM, ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM,
    /*  p   q   r   s   t   u   v   w    x   y   z   {   |   }   ~   del  */
	ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM, ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM,
};

/* token type table: don't strip comments */
u_char	TokTypeNoC[256] =
{
    /*	nul soh stx etx eot enq ack bel  bs  ht  nl  vt  np  cr  so  si   */
	ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM, ATM,SPC,SPC,SPC,SPC,SPC,ATM,ATM,
    /*	dle dc1 dc2 dc3 dc4 nak syn etb  can em  sub esc fs  gs  rs  us   */
	ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM, ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM,
    /*  sp  !   "   #   $   %   &   '    (   )   *   +   ,   -   .   /    */
	SPC,ATM,QST,ATM,ATM,ATM,ATM,ATM, OPR,OPR,ATM,ATM,ATM,ATM,ATM,ATM,
    /*	0   1   2   3   4   5   6   7    8   9   :   ;   <   =   >   ?    */
	ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM, ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM,
    /*	@   A   B   C   D   E   F   G    H   I   J   K   L   M   N   O    */
	ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM, ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM,
    /*  P   Q   R   S   T   U   V   W    X   Y   Z   [   \   ]   ^   _    */
	ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM, ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM,
    /*	`   a   b   c   d   e   f   g    h   i   j   k   l   m   n   o    */
	ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM, ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM,
    /*  p   q   r   s   t   u   v   w    x   y   z   {   |   }   ~   del  */
	ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM, ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM,

    /*	nul soh stx etx eot enq ack bel  bs  ht  nl  vt  np  cr  so  si   */
	OPR,OPR,ONE,OPR,OPR,OPR,OPR,OPR, OPR,OPR,OPR,OPR,OPR,OPR,OPR,OPR,
    /*	dle dc1 dc2 dc3 dc4 nak syn etb  can em  sub esc fs  gs  rs  us   */
	OPR,OPR,OPR,ONE,ONE,ONE,OPR,OPR, OPR,OPR,OPR,OPR,OPR,OPR,OPR,OPR,
    /*  sp  !   "   #   $   %   &   '    (   )   *   +   ,   -   .   /    */
	ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM, ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM,
    /*	0   1   2   3   4   5   6   7    8   9   :   ;   <   =   >   ?    */
	ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM, ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM,
    /*	@   A   B   C   D   E   F   G    H   I   J   K   L   M   N   O    */
	ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM, ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM,
    /*  P   Q   R   S   T   U   V   W    X   Y   Z   [   \   ]   ^   _    */
	ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM, ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM,
    /*	`   a   b   c   d   e   f   g    h   i   j   k   l   m   n   o    */
	ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM, ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM,
    /*  p   q   r   s   t   u   v   w    x   y   z   {   |   }   ~   del  */
	ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM, ATM,ATM,ATM,ATM,ATM,ATM,ATM,ATM,
};

#define NOCHAR		-1	/* signal nothing in lookahead token */

#define DELIMCHARS	"()<>,;\r\n"	/* default word delimiters */

char	*OperatorChars;	/* operators (old $o macro) */
int	ConfigLevel;	/* config file level */


/*
**  PRESCAN -- Prescan name and make it canonical
**
**	Scans a name and turns it into a set of tokens.  This process
**	deletes blanks and comments (in parentheses) (if the token type
**	for left paren is SPC).
**
**	This routine knows about quoted strings and angle brackets.
**
**	There are certain subtleties to this routine.  The one that
**	comes to mind now is that backslashes on the ends of names
**	are silently stripped off; this is intentional.  The problem
**	is that some versions of sndmsg (like at LBL) set the kill
**	character to something other than @ when reading addresses;
**	so people type "csvax.eric\@berkeley" -- which screws up the
**	berknet mailer.
**
**	Parameters:
**		addr -- the name to chomp.
**		delim -- the delimiter for the address, normally
**			'\0' or ','; \0 is accepted in any case.
**			If '\t' then we are reading the .cf file.
**		pvpbuf -- place to put the saved text -- note that
**			the pointers are static.
**		pvpbsize -- size of pvpbuf.
**		delimptr -- if non-NULL, set to the location of the
**			terminating delimiter.
**		toktab -- if set, a token table to use for parsing.
**			If NULL, use the default table.
**
**	Returns:
**		A pointer to a vector of tokens.
**		NULL on error.
*/

char **
prescan(addr, delim, pvpbuf, pvpbsize, delimptr, toktab, canary)
	char *addr;
	int delim;
	char pvpbuf[];
	int pvpbsize;
	char **delimptr;
	u_char *toktab;
	char *canary;
{
	register char *p;
	register char *q;
	register int c;

	bool bslashmode;
	bool route_syntax;
	int cmntcnt;
	int anglecnt;
	char *tok;
	int state;
	int newstate;
	char *saveto = CurEnv->e_to;
	/*static char *av[MAXATOM + 1]; */
	/*static char firsttime = FALSE; */
	int errno;

        
	printf("Inside prescan!!\n");
        printf("Max storage of pvpbuf = %d\n", PSBUFSIZE);


	if (toktab == NULL)
		toktab = TokTypeTab;

	/* make sure error messages don't have garbage on them */
	errno = 0;

	q = pvpbuf;
	bslashmode = FALSE;
	route_syntax = FALSE;
	cmntcnt = 0;
	anglecnt = 0;
	
        /* avp = av; */

	state = ATM;
	c = NOCHAR;
	p = addr;
	CurEnv->e_to = p;

	do
	{
		/* read a token */
		tok = q;
		for (;;)
		{
			/* store away any old lookahead character */
			if (c != NOCHAR && !bslashmode)
			{
				/* see if there is room */
			  
				if (q >= &pvpbuf[pvpbsize - 5])
				{
				  printf("553 5.1.1 Address too long\n");
				  
				  if (strlen(addr) > (SIZE_T) MAXNAME)
				    {
				      printf("strlen(addr) > %d\n", MAXNAME);
				      addr[MAXNAME] = '\0';
				    }
				    returnnull:  
					if (delimptr != NULL)
					  *delimptr = p;
					CurEnv->e_to = saveto;
					return NULL;
				}

				/* squirrel it away */
				printf("Writing %c to q!\n", c);
				*q++ = c;
			}

			/* read a new input character */
			c = *p++;
			if (c == '\0')
			{
				/* diagnose and patch up bad syntax */
				if (state == QST)
				{
					printf("653 Unbalanced '\"'");
					c = '"';
				}
				else if (cmntcnt > 0)
				{
					printf("653 Unbalanced '('");
					c = ')';
				}
				else if (anglecnt > 0)
				{
					c = '>';
					printf("653 Unbalanced '<'");
				}
				else
					break;

				p--;
			}
			else if (c == delim && cmntcnt <= 0 && state != QST)
			{
				if (anglecnt <= 0)
					break;

				/* special case for better error management */
				if (delim == ',' && !route_syntax)
				{
					printf("653 Unbalanced '<'");
					c = '>';
					p--;
				}
			}

			/* chew up special characters */
			
                        /*BAD*/  
                        *q = '\0';
			printf ("canaray=[%s]\n", canary);
			printf ("canary should be 5 bytes, max\n");

			if (bslashmode)
			{

			  /*printf("bslashmode = TRUE!!!\n");*/

				bslashmode = FALSE;

				/* kludge \! for naive users */
				if (cmntcnt > 0)
				{
					c = NOCHAR;
					continue;
				}
				else if (c != '!' || state == QST)
				{

				  /*BAD*/	
				  *q++ = '\\';     
				  printf ("canary=[%s]\n", canary);
				  printf ("canary should be 5 bytes, max\n");
				  continue;   /* continue while loop */
				}
			}

			if (c == '\\')
			{
				bslashmode = TRUE;
			}
			else if (state == QST)
			{
				/* EMPTY */
				/* do nothing, just avoid next clauses */
			}
			else if (c == '(' && toktab['('] == SPC)
			{
				cmntcnt++;
				c = NOCHAR;
			}
			else if (c == ')' && toktab['('] == SPC)
			{
				if (cmntcnt <= 0)
				{
					printf("653 Unbalanced ')'");
					c = NOCHAR;
				}
				else
					cmntcnt--;
			}
			else if (cmntcnt > 0)
			{
				c = NOCHAR;
			}
			else if (c == '<')
			{
				char *ptr = p;

				anglecnt++;
				while (isascii((int)*ptr) && isspace((int)*ptr))
					ptr++;
				if (*ptr == '@')
					route_syntax = TRUE;
			}
			else if (c == '>')
			{
				if (anglecnt <= 0)
				{
					printf("653 Unbalanced '>'");
					c = NOCHAR;
				}
				else{
				  anglecnt--;
				}
				route_syntax = FALSE;
			}
			else if (delim == ' ' && isascii(c) && isspace(c))
				c = ' ';

			if (c == NOCHAR){
			  printf("c = NOCHAR.... continuing....!\n");
				continue;
			}

			/* see if this is end of input */
			if (c == delim && anglecnt <= 0 && state != QST)
			  {
			    printf("breaking from for loop!\n");
			  break;
			  }
                        
                        /*printf("oldstate = %d\n", newstate); */
			newstate = StateTab[state][toktab[c & 0xff]];
			/*printf("newstate = %d\n", newstate); */
			
			state = newstate & TYPE;
			if (state == ILL)
			{
				if (isascii(c) && isprint(c))
					printf("653 Illegal character %c", c);
				else
					printf("653 Illegal character 0x%02x", c);
			}
			/* if (bitset(M, newstate)) */
			if (newstate & M)                 /* replacement for bitset */ {
			  c = NOCHAR;
			}
			/* if (bitset(B, newstate)) */
			if (newstate & B)
			  {
			    break;
			  }
		}

		if (tok != q)
		{
		        printf("writing null byte\n");
			/*BAD*/
			*q++ = '\0';
			printf ("canary=[%s]\n", canary);
			printf ("canary should be 5 bytes, max\n");

			if (q - tok > MAXNAME)
			{
				printf("553 5.1.0 prescan: token too long");
				goto returnnull;
			}
		}
	


        
	} while (c != '\0' && (c != delim || anglecnt > 0));

        printf("Exiting while loop!\n");

	p--;
	if (delimptr != NULL)
		*delimptr = p;

	CurEnv->e_to = saveto;

	return NULL;
}

char **
parseaddr(addr, delim, delimptr)
	char *addr;
	int delim;
	char **delimptr;
{

  register char **pvp; 
  char canary[] = "GOOD";
  char pvpbuf[PSBUFSIZE];
  static char *delimptrbuf;

  /*
  **  Initialize and prescan address.
  */


  if (delimptr == NULL)
    delimptr = &delimptrbuf;
  
  pvp = prescan(addr, delim, pvpbuf, sizeof pvpbuf, delimptr, NULL, canary);

  return pvp;


}

int main(){
 
  char *addr;
  int delim;
  
  static char **delimptr;
  char special_char = '\377';  /* same char as 0xff.  this char will get interpreted as NOCHAR */
  int i = 0;
  
  addr = (char *) malloc(sizeof(char) * 500);
  /* This address is valid */
  /* strcpy(addr, "Misha Zitser <misha@mit.edu>"); */


  /* This address causes a buffer overflow and results in a seg fault */  
  /* create malicious address */
  /* pvpbuf inside prescan gets overflowed and 2 LSBs of the saved frame pointer of prescan get */
  /* overwritten with "000C".  if a fake stack frame is set up cleverly, then arbitrary code could
     be executed.*/ 
  
  for(i=0; i<300; i=i+2){   
    addr[i] = '\\';
    addr[i+1] = special_char;       
 }

  delim = '\0';
  delimptr = NULL; 

  OperatorChars = NULL;
 
  ConfigLevel = 5;
  
  CurEnv = (ENVELOPE *) malloc(sizeof(struct envelope));
  CurEnv->e_to = (char *) malloc(strlen(addr) * sizeof(char) + 1); 

  strcpy(CurEnv->e_to, addr);   

  parseaddr(addr, delim, delimptr);
  

  return 0;
}


/*

</source>

*/

