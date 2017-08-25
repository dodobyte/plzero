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

func out(data interface{}) {
	size := code.Len()
	binary.Write(&code, binary.LittleEndian, data)
	pc += code.Len() - size
}

func genGlobVar(val int) {
	out(int32(val))
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
		out(uint8(0x8B))
		out(uint8(0x05 + regId[regs[reg]]*8))
		out(uint32(dataBase + sym.addr))
	} else {
		out(uint8(0x8B))
		out(uint8(0x84 + regId[regs[reg]]*8))
		out(uint8(0x24))
		out(uint32(sym.addr))
	}
	reg++
}

func genImm(val int) {
	out(uint8(0xB8 + regId[regs[reg]]))
	out(int32(val))
	reg++
}

func genMul() {
	reg--
	r1 := regId[regs[reg-1]]
	r2 := regId[regs[reg]]
	out(uint8(0x0F))
	out(uint8(0xAF))
	out(uint8(0xC0 + r1*8 + r2))
}

func genPush(reg string) {
	out(uint8(0x50 + regId[reg]))
}

func genPop(reg string) {
	out(uint8(0x58 + regId[reg]))
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
	out(uint8(0x31))
	out(uint8(0xD2))
	/* idiv ebx */
	out(uint8(0xF7))
	out(uint8(0xFB))
	if reg > 2 {
		genPush("eax")
		for i := reg - 2; i >= 0; i-- {
			genPop(regs[i])
		}
	}
	reg--
}

func genNeg() {
	out(uint8(0xF7))
	out(uint8(0xD8 + regId[regs[reg-1]]))
}

func genAddSub(op string) {
	opc := uint8(0x01)
	if op == "-" {
		opc = uint8(0x29)
	}
	reg--
	r1 := regId[regs[reg-1]]
	r2 := regId[regs[reg]]
	out(opc)
	out(uint8(0xC0 + r2*8 + r1))
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
		out(uint8(0x89))
		out(uint8(0x05 + regId[regs[reg]]*8))
		out(uint32(dataBase + sym.addr))
	} else {
		out(uint8(0x89))
		out(uint8(0x84 + regId[regs[reg]]*8))
		out(uint8(0x24))
		out(uint32(sym.addr))
	}
}

func genCall(fn string) {
	call := scopes[fn].addr - (pc + 5)
	out(uint8(0xE8))
	out(int32(call))
}

func genJmp(jpc int) {
	jmp := jpc - (pc + 5)
	out(uint8(0xE9))
	out(int32(jmp))
}

func genProc(name, part string) {
	if part == "head" {
		out(uint8(0x81))
		out(uint8(0xEC))
		out(uint32(scopes[name].nlocal * 4))
	} else {
		out(uint8(0x81))
		out(uint8(0xC4))
		out(uint32(scopes[name].nlocal * 4))
		out(uint8(0xC3))
	}
}

func genOdd() {
	reg--
	out(uint8(0xF7))
	out(uint8(0xC0 + regId[regs[reg]]))
	out(uint32(0x01))
	out(uint8(0x0F))
	out(uint8(0x84))
	fixStack = append(fixStack, &fix{pc, 0})
	out(int32(0x00))
}

func genCond(cond string) {
	r1 := regId[regs[reg-2]]
	r2 := regId[regs[reg-1]]
	out(uint8(0x39))
	out(uint8(0xC0 + r2*8 + r1))
	reg -= 2

	var opc uint8
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
	out(uint8(0x0F))
	out(opc)
	fixStack = append(fixStack, &fix{pc, 0})
	out(int32(0x00))
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
		out(uint8(0x68))
		out(uint32(dataBase + sym.addr))
	} else {
		out(uint8(0x8D))
		out(uint8(0x84 + regId[regs[reg]]*8))
		out(uint8(0x24))
		out(uint32(sym.addr + 12))
		genPush(regs[reg])
	}
	out(uint8(0x68))
	out(uint32(sFmtAddr))
	out(uint8(0xFF))
	out(uint8(0x15))
	out(uint32(imp_scanf))
	out(uint8(0x83))
	out(uint8(0xC4))
	out(uint8(0x08))
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
	out(uint8(0x68))
	out(uint32(pFmtAddr))
	out(uint8(0xFF))
	out(uint8(0x15))
	out(uint32(imp_printf))
	out(uint8(0x83))
	out(uint8(0xC4))
	out(uint8(0x08))
	genPop("edx")
	genPop("ecx")
	genPop("eax")
}

func genExit() {
	out(uint8(0x6A))
	out(uint8(0x00))
	out(uint8(0xFF))
	out(uint8(0x15))
	out(uint32(imp_exit))
	out(uint8(0x83))
	out(uint8(0xC4))
	out(uint8(0x04))
}
