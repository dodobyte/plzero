package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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

	infile := os.Args[1]
	outfile := strings.Split(infile, ".")[0] + ".exe"

	createImports()
	code.Write(imports)

	compile(os.Args[1])

	codeByte := code.Bytes()
	genFixup(codeByte[len(imports):])

	dumpExe(codeByte, outfile)
}
