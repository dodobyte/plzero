package main

import (
	"bufio"
	"fmt"
	"os"
	"unicode"
)

var keyword = map[string]bool{
	"const":     true,
	"var":       true,
	"procedure": true,
	"call":      true,
	"begin":     true,
	"end":       true,
	"if":        true,
	"then":      true,
	"while":     true,
	"do":        true,
	"odd":       true,
	"read":      true,
	"write":     true,
}

var operator = map[string]bool{
	":=": true,
	"=":  true,
	"#":  true,
	"<":  true,
	"<=": true,
	">":  true,
	">=": true,
	"+":  true,
	"-":  true,
	"*":  true,
	"/":  true,
	"(":  true,
	")":  true,
	";":  true,
	",":  true,
	".":  true,
}

type token struct {
	typ  string // integer, identifier, keyword operator
	name string // identifiers name if typ == identifier
	val  int    // integer constant if typ == integer
}

var tok token

type symbol struct {
	typ string // const, var
	val int    // int. constant if typ == const
}
type symtab map[string]symbol        // symbol table for a single scope
var scopes = make(map[string]symtab) // scopes: procedure, global

var line = 1
var reader *bufio.Reader

func fatal(msg string) {
	fmt.Fprintf(os.Stderr, "%d: %s", line, msg)
	os.Exit(1)
}

func readByte() byte {
	c, err := reader.ReadByte()
	if err != nil {
		fatal(err.Error())
	}
	return c
}

func next() {
	tok = token{}
	c := readByte()
	r := rune(c)
	/* skip whitespaces */
	if unicode.IsSpace(r) {
		if c == '\n' {
			line++
		}
		next()
		return
	}
	/* integer constant */
	if unicode.IsDigit(r) {
		reader.UnreadByte()
		tok.typ = "integer"
		fmt.Fscanf(reader, "%d", &tok.val)
		return
	}
	/* keyword or identifier */
	if c == '_' || unicode.IsLetter(r) {
		name := string(c)
		for {
			c := readByte()
			r := rune(c)
			if c == '_' || unicode.IsDigit(r) || unicode.IsLetter(r) {
				name += string(c)
			} else {
				reader.UnreadByte()
				break
			}
		}
		if keyword[name] {
			tok.typ = name
		} else {
			tok.typ = "identifier"
			tok.name = name
		}
		return
	}
	/* operator */
	op := string(c)
	if op == "." {
		tok.typ = op
		return
	}
	c = readByte()
	op2 := op + string(c)
	switch {
	case operator[op2]:
		tok.typ = op2
	case operator[op]:
		tok.typ = op
		reader.UnreadByte()
	default:
		fatal("unknown token")
	}
}

func accept(tokType string) {
	next()
	if tok.typ != tokType {
		fatal(tokType + " expected")
	}
}

func expect(tokType string) {
	if tok.typ != tokType {
		fatal(tokType + " expected")
	}
	next()
}

func declare(scope, typ, name string, val int) {
	_, ok := scopes[scope][name]
	if ok {
		fatal(name + " redeclared")
	}
	scopes[scope][name] = symbol{typ, val}
}

func factor() {
	switch tok.typ {
	case "identifier":
		next()
	case "integer":
		next()
	case "(":
		next()
		expression()
		expect(")")
	default:
		fatal("invalid factor: " + tok.typ)
	}
}

func term() {
	factor()
	for tok.typ == "*" || tok.typ == "/" {
		//op := tok.typ
		next()
		factor()
	}
}

func expression() {
	//neg := false
	if tok.typ == "+" || tok.typ == "-" {
		if tok.typ == "-" {
			//neg = true
		}
		next()
	}
	term()
	for tok.typ == "+" || tok.typ == "-" {
		//op := tok.typ
		next()
		term()
	}
}

func condition() {
	if tok.typ == "odd" {
		next()
		expression()
	} else {
		expression()
		switch tok.typ {
		case "=", "#", "<", "<=", ">", ">=":
			next()
			expression()
		default:
			fatal("comparison operator expected")
		}
	}
}

func statement(scope string) {
	switch tok.typ {
	case "identifier":
		accept(":=")
		next()
		expression()
	case "call", "read":
		accept("identifier")
		next()
	case "write":
		expression()
	case "if":
		next()
		condition()
		expect("then")
		statement(scope)
	case "while":
		next()
		condition()
		expect("do")
		statement(scope)
	case "begin":
		next()
		statement(scope)
		for tok.typ == ";" {
			next()
			statement(scope)
		}
		expect("end")
	default:
		fatal("invalid statement: " + tok.typ)
	}
}

func block(scope string) {
	next()
	if _, ok := scopes[scope]; !ok {
		scopes[scope] = make(symtab)
	}
	if tok.typ == "const" {
		for {
			accept("identifier")
			name := tok.name
			accept("=")
			accept("integer")
			declare(scope, "const", name, tok.val)
			next()
			if tok.typ == ";" {
				break
			}
			if tok.typ != "," {
				fatal(", expected.")
			}
		}
		next()
	}
	if tok.typ == "var" {
		for {
			accept("identifier")
			declare(scope, "var", tok.name, 0)
			next()
			if tok.typ == ";" {
				break
			}
			if tok.typ != "," {
				fatal(", expected.")
			}
		}
		next()
	}
	for tok.typ == "procedure" {
		accept("identifier")
		name := tok.name
		accept(";")
		block(name)
		expect(";")
	}
	statement(scope)
}

func program() {
	block("") // global scope
	if tok.typ != "." {
		fatal(". expected")
	}
}

func main() {
	reader = bufio.NewReader(os.Stdin)
	program()
	fmt.Println(scopes)
}
