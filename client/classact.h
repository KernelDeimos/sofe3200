#ifndef CLASSACT_H_
#define CLASSACT_H_

#include <stdio.h>

extern FILE *get_input_file(const char *server, const char *name);
extern FILE *get_output_file();

extern void send_output_file(const char *server, const char *name);

#endif /* CLASSACT_H_ */

