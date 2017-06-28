
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
$Date: 2004/01/05 17:27:43 $
$Header: /mnt/leo2/cvs/sabo/hist-040105/sendmail/s3/main.c,v 1.1.1.1 2004/01/05 17:27:43 tleek Exp $



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
$Date: 2004/01/05 17:27:43 $
$Header: /mnt/leo2/cvs/sabo/hist-040105/sendmail/s3/main.c,v 1.1.1.1 2004/01/05 17:27:43 tleek Exp $



*/


/*

<source>

*/

#include "my-sendmail.h"
#include <assert.h>

int main(int argc, char **argv){

  HDR *header;
  register ENVELOPE *e;
  FILE *temp;

  assert (argc == 2);
  temp = fopen (argv[1], "r");
  assert (temp != NULL);
 
  header = (HDR *) malloc(sizeof(struct header));
  
  header->h_field = (char *) malloc(sizeof(char) * 100);
  header->h_field = "Content-Transfer-Encoding";
  header->h_value = (char *) malloc(sizeof(char) * 100);
  header->h_value = "quoted-printable";

  e = (ENVELOPE *) malloc(sizeof(struct envelope));
 
  e->e_id = (char *) malloc(sizeof(char) * 50);
  e->e_id = "First Entry";

 
  e->e_dfp = temp;
  mime7to8(header, e);

  fclose(temp);

  return 0;

}

/*

</source>

*/

