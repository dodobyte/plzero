# PL/0 Compiler Written in Go

PL/0 compiler that produces Windows executable.

- Pure Go parsing, no parser generator needed.
- No third party software is needed like assembler or linker.
- Very small and simple, only ~700 lines of code.
  - Recursive descent parser
  - X86 code generator
  - Win32 PE executable creator
  
See the [article](http://dogankurt.com/plzero.html) for more information.

