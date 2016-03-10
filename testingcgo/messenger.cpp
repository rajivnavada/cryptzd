#include <cstdlib>
#include <string>

#include "messenger.h"

using namespace std;

char *prepareMessage(const char *prefix)
{
  size_t prefixlen = strlen(prefix);
  size_t len = (prefixlen + 10) * sizeof(char);
  char *ret = (char *)malloc(len);
  memset(ret, 0, len);

  ret = strncpy(ret, prefix, prefixlen);
  (void)strncpy(ret + prefixlen, " World", 6);

  return ret;
}
