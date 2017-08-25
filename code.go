package main

import (
	"bytes"
	"encoding/binary"
)

var pc = 0
var ep = 0
var datap = 0

var reg int
var regs = []string{"eax", "ebx", "ecx", "edx", "ebp", "esi", "edi"}

var regId = map[string]int{
	"eax": 0, "ecx": 1, "edx": 2, "ebx": 3,
	"esp": 4, "ebp": 5, "esi": 6, "edi": 7}

var code bytes.Buffer

type fix struct {
	pc, jmp int
}

var fixup []*fix
var fixStack []*fix

func out(format string, data ...interface{}) {
	le := binary.LittleEndian
	size := code.Len()
	for i, c := range format {
		switch c {
		case 'b':
			binary.Write(&code, le, uint8(data[i].(int)))
		case 'i':
			binary.Write(&code, le, int32(data[i].(int)))
		case 'u':
			binary.Write(&code, le, uint32(data[i].(int)))
		}
	}
	pc += code.Len() - size
}

func genGlobVar(val int) {
	out("u", val)
	codeBase += 4
}

func genIdent(name string) {
	glob := active == ""
	sym, ok := scopes[active].sym[name]
	if !ok {
		glob = true
		sym = scopes[""].sym[name]
	}
	if sym.typ == "const" {
		genImm(sym.val)
		return
	}
	if glob {
		out("b", 0x8B)
		out("b", 0x05+regId[regs[reg]]*8)
		out("u", dataBase+sym.addr)
	} else {
		out("b", 0x8B)
		out("b", 0x84+regId[regs[reg]]*8)
		out("bu", 0x24, sym.addr)
	}
	reg++
}

func genImm(val int) {
	out("b", 0xB8+regId[regs[reg]])
	out("i", val)
	reg++
}

func genMul() {
	reg--
	r1 := regId[regs[reg-1]]
	r2 := regId[regs[reg]]
	out("bbb", 0x0F, 0xAF, 0xC0+r1*8+r2)
}

func genPush(reg string) {
	out("b", 0x50+regId[reg])
}

func genPop(reg string) {
	out("b", 0x58+regId[reg])
}

func genDiv() {
	if reg > 2 {
		for i := 0; i < reg; i++ {
			genPush(regs[i])
		}
		genPop("ebx")
		genPop("eax")
	}
	/* xor edx, edx */
	out("bb", 0x31, 0xD2)
	/* idiv ebx */
	out("bb", 0xF7, 0xFB)
	if reg > 2 {
		genPush("eax")
		for i := reg - 2; i >= 0; i-- {
			genPop(regs[i])
		}
	}
	reg--
}

func genNeg() {
	out("b", 0xF7)
	out("b", 0xD8+regId[regs[reg-1]])
}

func genAddSub(op string) {
	opc := 0x01
	if op == "-" {
		opc = 0x29
	}
	reg--
	r1 := regId[regs[reg-1]]
	r2 := regId[regs[reg]]
	out("bb", opc, 0xC0+r2*8+r1)
}

func genAssign(name string) {
	glob := active == ""
	sym, ok := scopes[active].sym[name]
	if !ok {
		glob = true
		sym = scopes[""].sym[name]
	}
	reg--
	if glob {
		out("b", 0x89)
		out("b", 0x05+regId[regs[reg]]*8)
		out("u", dataBase+sym.addr)
	} else {
		out("b", 0x89)
		out("b", 0x84+regId[regs[reg]]*8)
		out("b", 0x24)
		out("u", sym.addr)
	}
}

func genCall(fn string) {
	call := scopes[fn].addr - (pc + 5)
	out("bi", 0xE8, call)
}

func genJmp(jpc int) {
	jmp := jpc - (pc + 5)
	out("bi", 0xE9, jmp)
}

func genProc(name, part string) {
	if part == "head" {
		out("bb", 0x81, 0xEC)
		out("u", scopes[name].nlocal*4)
	} else {
		out("bb", 0x81, 0xC4)
		out("u", scopes[name].nlocal*4)
		out("b", 0xC3)
	}
}

func genOdd() {
	reg--
	out("b", 0xF7)
	out("b", 0xC0+regId[regs[reg]])
	out("u", 0x01)
	out("bb", 0x0F, 0x84)
	fixStack = append(fixStack, &fix{pc, 0})
	out("i", 0x00)
}

func genCond(cond string) {
	r1 := regId[regs[reg-2]]
	r2 := regId[regs[reg-1]]
	out("bb", 0x39, 0xC0+r2*8+r1)
	reg -= 2

	var opc int
	switch cond {
	case "=":
		opc = 0x85 // jne
	case "#":
		opc = 0x84 // je
	case "<":
		opc = 0x8D // jnl
	case "<=":
		opc = 0x8F // jg
	case ">":
		opc = 0x8E // jng
	case ">=":
		opc = 0x8C // jl
	default:
		fatal("comparison operator expected")
	}
	out("bb", 0x0F, opc)
	fixStack = append(fixStack, &fix{pc, 0})
	out("i", 0x00)
}

func genLabel() {
	fix := fixStack[len(fixStack)-1]
	fixStack = fixStack[:len(fixStack)-1]
	fix.jmp = pc - (fix.pc + 4)
	fixup = append(fixup, fix)
}

func genFixup(cod []byte) {
	for _, fix := range fixup {
		writeInt32(cod[fix.pc:], int32(fix.jmp))
	}
}

func genRead(name string) {
	genPush("eax")
	genPush("ecx")
	genPush("edx")
	glob := active == ""
	sym, ok := scopes[active].sym[name]
	if !ok {
		glob = true
		sym = scopes[""].sym[name]
	}
	if glob {
		out("bu", 0x68, dataBase+sym.addr)
	} else {
		out("b", 0x8D)
		out("b", 0x84+regId[regs[reg]]*8)
		out("bu", 0x24, sym.addr+12)
		genPush(regs[reg])
	}
	out("bu", 0x68, sFmtAddr)
	out("bbu", 0xFF, 0x15, imp_scanf)
	out("bbb", 0x83, 0xC4, 0x08)
	genPop("edx")
	genPop("ecx")
	genPop("eax")
}

func genWrite() {
	reg--
	genPush("eax")
	genPush("ecx")
	genPush("edx")
	genPush(regs[reg])
	out("bu", 0x68, pFmtAddr)
	out("bbu", 0xFF, 0x15, imp_printf)
	out("bbb", 0x83, 0xC4, 0x08)
	genPop("edx")
	genPop("ecx")
	genPop("eax")
}

func genExit() {
	out("bb", 0x6A, 00)
	out("bbu", 0xFF, 0x15, imp_exit)
	out("bbb", 0x83, 0xC4, 0x04)
}
