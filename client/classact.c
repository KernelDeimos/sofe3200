#include <curl/curl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

#include "classact.h"

FILE* CLASSACT_OUTPUT = NULL;

size_t static write_callback_func(
	void *buffer, size_t size,
	size_t nmemb, void *userp);

char* get_uuid(const char *server);

FILE *get_input_file(const char *server, const char *name)
{
		char* uuid = get_uuid(server);
		// Generate API URL
		char buf[512];
		sprintf(buf, "%s/%s/%s/%s",
			server, "recv", uuid, name);

		CURL* curl = curl_easy_init();
		curl_easy_setopt(curl, CURLOPT_URL, buf);
		FILE* fileIn = fopen("/tmp/classact_input", "w");
		curl_easy_setopt(curl, CURLOPT_WRITEDATA, fileIn);
		curl_easy_perform(curl);
		curl_easy_cleanup(curl);
		fclose(fileIn);

		FILE* fileOut = fopen("/tmp/classact_input", "r");
		return fileOut;
}

FILE *get_output_file(const char *server)
{
		char fileName[512];
		CLASSACT_OUTPUT = fopen("/tmp/classact_output", "w");
		return CLASSACT_OUTPUT;
}

void send_output_file(const char *server, const char *name)
{
		// Generate API URL
		char* uuid = get_uuid(server);
		char buf[512];
		sprintf(buf, "%s/%s/%s/%s",
			server, "send", uuid, name);

		printf("[CLASSACT] Client UUID: %s\n", uuid);
		FILE* file = fopen("/tmp/classact_output", "r");
		
		// Read entire output file into a buffer
		char* buffer = NULL;
		fseek(file, 0, SEEK_END);
		long length = ftell(file);
		rewind(file);
		buffer = calloc(length, sizeof(char));
		fread(buffer, 1, length, file);
		fclose(file);

		// Remove this file (it's not required anymore)
		remove("/tmp/classact_output"); // This feels too easy

		char* qstring = calloc(strlen(buffer)+6, sizeof(char));
		strcat(qstring, "data=");
		strcat(qstring, buffer);

		printf("[CLASSACT] Sending POST to %s?%s\n", buf, qstring);

		CURL* curl = curl_easy_init();
		char* escaped = curl_easy_escape(curl, buffer, 0);
		curl_easy_setopt(curl, CURLOPT_URL, buf);
		curl_easy_setopt(curl, CURLOPT_POSTFIELDS, qstring);
		curl_easy_perform(curl);
}

char* get_uuid(const char *server)
{
		char buf[128];
		strcpy(buf, server);
		strcat(buf, "/uuid");

		char* uuid = NULL;

		// Attempt to get stored UUID
		FILE* uuidFile;
		uuidFile = fopen("/tmp/classact_uuid", "r");
		if (uuidFile) {
				uuid = calloc(36, sizeof(char));
				fread(uuid, 1, 36, uuidFile);
		} else {
			CURL* curl = curl_easy_init();
			curl_easy_setopt(curl, CURLOPT_URL, buf);
			curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, write_callback_func);
			curl_easy_setopt(curl, CURLOPT_WRITEDATA, &uuid);
			curl_easy_perform(curl);
			curl_easy_cleanup(curl);
			uuidFile = fopen("/tmp/classact_uuid", "w");
			fputs(uuid, uuidFile);
		}

		return uuid;
}

/* Source for below function:
http://www.dimuthu.org/blog/2009/01/28/making-web-requests-using-curl-from-c-and-php/
*/
/* the function to invoke as the data recieved */
size_t static write_callback_func(
	void *buffer, size_t size,
	size_t nmemb, void *userp)
{
    char **response_ptr =  (char**)userp;

    /* assuming the response is a string */
    *response_ptr = strndup(buffer, (size_t)(size *nmemb));

}

