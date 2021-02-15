// I've just removed some debug prints from the original file. -- Michael Macnair

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
$Date: 2004/02/05 15:19:59 $
$Header: /mnt/leo2/cvs/sabo/hist-040105/sendmail/s3/mime1-bad.c,v 1.2 2004/02/05 15:19:59 tleek Exp $



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
$Date: 2004/02/05 15:19:59 $
$Header: /mnt/leo2/cvs/sabo/hist-040105/sendmail/s3/mime1-bad.c,v 1.2 2004/02/05 15:19:59 tleek Exp $



*/


/*

<source>

*/

# include <string.h>
# include "my-sendmail.h"


char *
xalloc(sz)
        register int sz;
{
        register char *p;
 
        /* some systems can't handle size zero mallocs */
        if (sz <= 0)
                sz = 1;
 
        p = malloc((unsigned) sz);
        if (p == NULL)
        {
                perror("Out of memory!!");
        }
        return (p);
}

char * hvalue(char *, HDR *);

/*
**  MIME7TO8 -- output 7 bit encoded MIME body in 8 bit format
**
**  This is a hack. Supports translating the two 7-bit body-encodings
**  (quoted-printable and base64) to 8-bit coded bodies.
**
**  There is not much point in supporting multipart here, as the UA
**  will be able to deal with encoded MIME bodies if it can parse MIME
**  multipart messages.
**
**  Note also that we wont be called unless it is a text/plain MIME
**  message, encoded base64 or QP and mailer flag '9' has been defined
**  on mailer.
**
**	Parameters:
**		header -- the header for this body part.
**		e -- envelope.
**
**	Returns:
**		none.
*/


void
mime7to8(header, e)
	HDR *header;
	register ENVELOPE *e;
{
	register char *p;
	u_char *obp;
	char buf[MAXLINE];
	char canary[10];
	u_char obuf[MAXLINE];

	strcpy(canary, "GOOD"); /* use canary to see if obuf gets overflowed */ 

	p = (char *) hvalue("Content-Transfer-Encoding", header);
	if (p == NULL)
	  {
	    printf("Content-Transfer-Encoding not found in header\n");
	    return;
	  }

	/*
	**  Translate body encoding to 8-bit.  Supports two types of
	**  encodings; "base64" and "quoted-printable". Assume qp if
	**  it is not base64.
	*/
	
	/* Misha: This is greatly modified */
	
	if (strcasecmp(p, "base64") == 0)
	{
	  printf("We do not handle base64 encoding...\n");
	  return;
	}
	else
	  {
	    /* quoted-printable */
	    obp = obuf;
	    while (fgets(buf, sizeof buf, e->e_dfp) != NULL)
	      {
		printf ("buf-obuf=%u\n", buf-(char *)obuf);
		printf ("obp-obuf=%u\n", obp-obuf);
		printf ("canary-obuf=%u\n", canary-(char *)obuf);
		
		if (mime_fromqp((u_char *) buf, &obp, 0, MAXLINE) == 0) {
		  printf("canary = %s\n", canary);		  
		  continue;
		}
		/*
		putline((char *) obuf, mci);
		*/
		obp = obuf;
		printf("canary = %s\n", canary);		
	      }
	    
	  }

	printf("obuf = %s\n",obuf);
	printf("canary should be GOOD\n");
	printf("canary = %s\n", canary);
}


/*
**  The following is based on Borenstein's "codes.c" module, with
**  simplifying changes as we do not deal with multipart, and to do
**  the translation in-core, with an attempt to prevent overrun of
**  output buffers.
**
**  What is needed here are changes to defined this code better against
**  bad encodings. Questionable to always return 0xFF for bad mappings.
*/

static char index_hex[128] =
{
	-1,-1,-1,-1, -1,-1,-1,-1, -1,-1,-1,-1, -1,-1,-1,-1,
	-1,-1,-1,-1, -1,-1,-1,-1, -1,-1,-1,-1, -1,-1,-1,-1,
	-1,-1,-1,-1, -1,-1,-1,-1, -1,-1,-1,-1, -1,-1,-1,-1,
	0, 1, 2, 3,  4, 5, 6, 7,  8, 9,-1,-1, -1,-1,-1,-1,
	-1,10,11,12, 13,14,15,-1, -1,-1,-1,-1, -1,-1,-1,-1,
	-1,-1,-1,-1, -1,-1,-1,-1, -1,-1,-1,-1, -1,-1,-1,-1,
	-1,10,11,12, 13,14,15,-1, -1,-1,-1,-1, -1,-1,-1,-1,
	-1,-1,-1,-1, -1,-1,-1,-1, -1,-1,-1,-1, -1,-1,-1,-1
};

#define HEXCHAR(c)  (((c) < 0 || (c) > 127) ? -1 : index_hex[(c)])

int
mime_fromqp(infile, outfile, state, maxlen)
	u_char *infile;
	u_char **outfile;
	int state;		/* Decoding body (0) or header (1) */
	int maxlen;		/* Max # of chars allowed in outfile */
{
  u_char *ooutfile;
	int c1, c2;
	int nchar = 0;

	ooutfile = *outfile; 

	while ((c1 = *infile++) != '\0')
	{
		if (c1 == '=')
		{
			if ((c1 = *infile++) == 0)
				break;

			if (c1 == '\n') /* ignore it */
			{
				if (state == 0)
					return 0;
			}
			else
			{
				if ((c2 = *infile++) == '\0')
					break;

				c1 = HEXCHAR(c1);
				c2 = HEXCHAR(c2);

				if (++nchar > maxlen)
					break;

				/* BAD */
				*(*outfile)++ = c1 << 4 | c2;
			}
		}
		else
		{
			if (state == 1 && c1 == '_')
				c1 = ' ';

			if (++nchar > maxlen)
				break;

			/*BAD*/
			*(*outfile)++ = c1;

			if (c1 == '\n')
			break;
		}
	}
	

	/*BAD*/
	*(*outfile)++ = '\0';
	return 1;
}


/*
**  HVALUE -- return value of a header.
**
**	Only "real" fields (i.e., ones that have not been supplied
**	as a default) are used.
**
**	Parameters:
**		field -- the field name.
**		header -- the header list.
**
**	Returns:
**		pointer to the value part.
**		NULL if not found.
**
**	Side Effects:
**		none.
*/

char * hvalue(field, header)
	char *field;
	HDR *header;
{
	register HDR *h;

	for (h = header; h != NULL; h = h->h_link)
	{
		if (!bitset(H_DEFAULT, h->h_flags) &&
		    strcasecmp(h->h_field, field) == 0)
			return (h->h_value);
	}
	return (NULL);
}



/*

</source>

*/

