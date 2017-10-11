# cbt (Cranberries Build Tool Version 0.1.0)

Build Tool for C/C++

# Usage

## cbt wandbox

'cbt wandbox' command allows to send your C/C++ codes to wandbox and show program result.

```cpp
// hello.cpp
#include <iostream>

int main(){
  std::cout << "Hello cbt!" << std::endl;
}
```

```
$ cbt wandbox cpp hello.cpp
hello cbt!
```
