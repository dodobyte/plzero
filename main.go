package main

import (
	"bufio"
	"fmt"
	"os"
)

func compile(name string) {
	f, err := os.Open(name)
	if err != nil {
		panic(err)
	}
	reader = bufio.NewReader(f)

	program()

	genExit()
	f.Close()
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: pl0 <source>\n")
		os.Exit(1)
	}
	createImports()
	code.Write(imports)
	compile(os.Args[1])
	cod := code.Bytes()
	genFixup(cod[len(imports):])
	pad := 512 - len(cod)%512
	cod = append(cod, make([]byte, pad)...)
	dumpExe(cod)
}
