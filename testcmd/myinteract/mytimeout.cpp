#include <windows.h>
#include <process.h>
#include <stdio.h>

#define BUFFER_MAX 1024

char g_nbstdin_buffer[2][BUFFER_MAX];
HANDLE g_input[2];
HANDLE g_process[2];

DWORD WINAPI console_input(LPVOID lpParameter)
{
	for (;;) {
		int i;
		for (i = 0; i<2; i++) {
			fgets(g_nbstdin_buffer[i], BUFFER_MAX, stdin);
			SetEvent(g_input[i]);
			WaitForSingleObject(g_process[i], INFINITE);
		}
	}
	return 0;
}

void create_nbstdin()
{
	int i;
	DWORD tid;
	CreateThread(NULL, 1024, &console_input, 0, 0, &tid);
	for (i = 0; i<2; i++) {
		g_input[i] = CreateEvent(NULL, FALSE, FALSE, NULL);
		g_process[i] = CreateEvent(NULL, FALSE, FALSE, NULL);
		g_nbstdin_buffer[i][0] = '\0';
	}
}

const char* nbstdin()
{
	DWORD n = WaitForMultipleObjects(2, g_input, FALSE, 0);
	if (n == WAIT_OBJECT_0 || n == WAIT_OBJECT_0 + 1) {
		n = n - WAIT_OBJECT_0;
		SetEvent(g_process[n]);
		return g_nbstdin_buffer[n];
	}
	else {
		return 0;
	}
}

void main()
{
	setvbuf(stdout, NULL, _IONBF, 0);
	printf(">Input q to quit..\n");
	create_nbstdin();
	for (;;) {
		const char *line = nbstdin();
		if (line) {
			printf(">%s", line);
			if (line[0] == 'q' && (line[1] == 0|| line[1] == '\n'))
				return ;
		}
		else {
			printf(".");
			Sleep(100);
		}
	}
}