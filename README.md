# cbt (Cranberries Build Tool Version 0.3.2) [![CircleCI](https://circleci.com/gh/LoliGothick/cbt/tree/master.svg?style=svg)](https://circleci.com/gh/LoliGothick/cbt/tree/master)

CLI Build Tool

# Usage

## cbt wandbox

`cbt wandbox` command allows to send your codes to wandbox and show program result.

Example

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

## Support Languages

- C
- C++
- Golang
- Ruby

## Never Support Languages

- Java
