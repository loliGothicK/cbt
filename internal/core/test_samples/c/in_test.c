#include <stdio.h>
#include <stdlib.h>

int main(void)
{
    char str[128];
    scanf("%[^\n]" , str);
    printf("%s\n" , str);
    return EXIT_SUCCESS;
}
