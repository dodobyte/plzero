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
	typ  string // const, var
	val  int    // int. constant if typ == const
	addr int    // address of symbol
}
type symtab map[string]symbol // symbol table for a single scope

type procedure struct {
	addr   int
	nlocal int
	sym    symtab
}

var scopes = make(map[string]procedure) // scopes: procedure, global
var active = ""

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

func check(name string, assignOp bool) {
	sym, ok := scopes[active].sym[name]
	g_sym, g_ok := scopes[""].sym[name]
	switch {
	case ok:
		if assignOp && sym.typ == "const" {
			fatal(name + " is constant")
		}
	case g_ok:
		if assignOp && g_sym.typ == "const" {
			fatal(name + " is constant")
		}
	default:
		fatal(name + " undeclared")
	}
}

func declare(scope, typ, name string, val int) {
	_, ok := scopes[scope].sym[name]
	if ok {
		fatal(name + " redeclared")
	}
	if scope == "" {
		scopes[scope].sym[name] = symbol{typ, val, datap}
		datap += 4
		genGlobVar(val)
	} else {
		proc := scopes[scope]
		addr := proc.nlocal * 4
		proc.nlocal++
		proc.sym[name] = symbol{typ, val, addr}
		scopes[scope] = proc
	}
}

func factor() {
	switch tok.typ {
	case "identifier":
		check(tok.name, false)
		genIdent(tok.name)
		next()
	case "integer":
		genImm(tok.val)
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
		op := tok.typ
		next()
		factor()
		if op == "*" {
			genMul()
		} else {
			genDiv()
		}
	}
}

func expression() {
	neg := false
	if tok.typ == "+" || tok.typ == "-" {
		if tok.typ == "-" {
			neg = true
		}
		next()
	}
	term()
	if neg {
		genNeg()
	}
	for tok.typ == "+" || tok.typ == "-" {
		op := tok.typ
		next()
		term()
		genAddSub(op)
	}
}

func condition() {
	if tok.typ == "odd" {
		next()
		expression()
		genOdd()
	} else {
		expression()
		cond := tok.typ
		next()
		expression()
		genCond(cond)
	}
}

func statement() {
	switch tok.typ {
	case "identifier":
		name := tok.name
		check(name, true)
		accept(":=")
		next()
		expression()
		genAssign(name)
	case "call":
		accept("identifier")
		genCall(tok.name)
		next()
	case "read":
		accept("identifier")
		check(tok.name, true)
		genRead(tok.name)
		next()
	case "write":
		next()
		expression()
		genWrite()
	case "if":
		next()
		condition()
		expect("then")
		statement()
		genLabel()
	case "while":
		wpc := pc
		next()
		condition()
		expect("do")
		statement()
		genJmp(wpc)
		genLabel()
	case "begin":
		next()
		statement()
		for tok.typ == ";" {
			next()
			statement()
		}
		expect("end")
	default:
		fatal("invalid statement: " + tok.typ)
	}
}

func block(scope string) {
	next()
	if _, ok := scopes[scope]; !ok {
		scopes[scope] = procedure{addr: pc, sym: make(symtab)}
		active = scope
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
	if scope != "" {
		genProc(scope, "head")
	} else {
		ep = textBase + importSize + pc
	}
	statement()
	if scope != "" {
		genProc(scope, "end")
	}
	active = ""
}

func program() {
	block("") // global scope
	if tok.typ != "." {
		fatal(". expected")
	}
}
