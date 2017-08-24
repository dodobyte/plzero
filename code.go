package main

import "fmt"

var reg int
var regs = []string{"eax", "ebx", "ecx", "edx", "ebp", "esi", "edi"}

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
		fmt.Printf("mov %s, dword[%s]\n", regs[reg] /*sym.addr*/, name)
	} else {
		fmt.Printf("mov %s, dword[esp + %d]\n", regs[reg], sym.addr)
	}
	reg++
}

func genImm(val int) {
	fmt.Printf("mov %s, %d\n", regs[reg], val)
	reg++
}

func genMul() {
	reg--
	fmt.Printf("imul %s, %s\n", regs[reg-1], regs[reg])
}

func genDiv() {
	if reg > 2 {
		for i := 0; i < reg; i++ {
			fmt.Println("push", regs[i])
		}
		fmt.Println("pop ebx")
		fmt.Println("pop eax")
	}
	fmt.Println("xor edx, edx")
	fmt.Println("idiv ebx")
	if reg > 2 {
		fmt.Println("push eax")
		for i := reg - 2; i >= 0; i-- {
			fmt.Println("pop", regs[i])
		}
	}
	reg--
}

func genNeg() {
	fmt.Println("neg", regs[reg-1])
}

func genAddSub(op string) {
	inst := "add"
	if op == "-" {
		inst = "sub"
	}
	reg--
	fmt.Printf("%s %s, %s\n", inst, regs[reg-1], regs[reg])
}

func genOdd() {
	reg--
	fmt.Printf("test %s, 1\n", regs[reg])
	fmt.Printf("jz L%d\n", labelid)
}

func genCond(cond string) {
	inst := ""
	switch cond {
	case "=":
		inst = "jne"
	case "#":
		inst = "je"
	case "<":
		inst = "jge"
	case "<=":
		inst = "jg"
	case ">":
		inst = "jle"
	case ">=":
		inst = "jl"
	default:
		fatal("comparison operator expected")
	}
	fmt.Printf("cmp %s, %s\n", regs[reg-2], regs[reg-1])
	reg -= 2
	fmt.Printf("%s L%d\n", inst, labelid)
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
		fmt.Printf("mov dword[%s], %s\n" /*sym.addr*/, name, regs[reg])
	} else {
		fmt.Printf("mov dword[esp + %d], %s\n", sym.addr, regs[reg])
	}
}

func genCall(fn string) {
	fmt.Println("call", fn)
}

func genLabel() {
	fmt.Printf("L%d:\n", getLabel())
}

func genJmp(pc int) {
	fmt.Printf("jmp WL%d\n", pc)
}

func genProc(name, part string) {
	if part == "head" {
		fmt.Printf("%s:\n", name)
		fmt.Printf("sub esp, %d\n", scopes[name].nlocal*4)
	} else {
		fmt.Printf("add esp, %d\n", scopes[name].nlocal*4)
		fmt.Println("ret")
	}
}
