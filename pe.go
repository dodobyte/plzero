package main

import (
	"bytes"
	"encoding/binary"
	"os"
)

const (
	imageBase  = 0x400000
	textBase   = 0x1000
	importSize = 100
)

var dataBase = imageBase + textBase + importSize
var codeBase = dataBase

var imports = make([]byte, 512)

var imp_exit int
var imp_scanf int
var imp_printf int

var sFmtAddr int
var pFmtAddr int

func writeInt32(dest []byte, data int32) {
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, data)
	copy(dest, buf.Bytes())
}

func int2byte(data int) []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, int32(data))
	return buf.Bytes()
}

/*
 * Since the necessary functions are predefined, we can construct
 * import table before compiling program.
 * Import table resides at the beginning of first section,
 * so addresses are static and don't depend on code sections size.
 */
func createImports() {
	/*
	 * import desc. = 20 bytes. end marker = 20 bytes.
	 * thunk array = 16 bytes (3 imports + end marker)
	 * no org. first thunk. we can't be bound, no big deal.
	 */
	copy(imports[56:], "msvcrt.dll\x00")
	copy(imports[67:], "\x00\x00exit\x00")
	copy(imports[74:], "\x00\x00scanf\x00")
	copy(imports[82:], "\x00\x00printf\x00")
	copy(imports[91:], "%d\x00")
	copy(imports[94:], "%d\n\x00")
	imports = imports[:importSize]

	/* name rva */
	writeInt32(imports[12:], 0x1000+56)
	/* first thunk */
	writeInt32(imports[16:], 0x1000+40)
	/* thunk array */
	writeInt32(imports[40:], 0x1000+67)
	writeInt32(imports[44:], 0x1000+74)
	writeInt32(imports[48:], 0x1000+82)

	/* this addresses will be used during compilation */
	imp_exit = 0x400000 + 0x1000 + 40
	imp_scanf = 0x400000 + 0x1000 + 44
	imp_printf = 0x400000 + 0x1000 + 48
	sFmtAddr = 0x400000 + 0x1000 + 91
	pFmtAddr = 0x400000 + 0x1000 + 94
}

/*
 * Construct necessary data structures for a minimum valid PE file.
 * Create final exe file with the compiled code and dump it to file.
 */
func dumpExe(code []byte) {
	f, err := os.Create("out.exe")
	if err != nil {
		panic(err)
	}

	/* DOS header. */
	f.Write([]byte("MZ"))
	f.Write(make([]byte, 0x3a))
	f.Write([]byte("\x40\x00\x00\x00"))

	/* NT File header. PE, x86, 1 section */
	f.Write([]byte("PE\x00\x00\x4C\x01\x01\x00"))
	f.Write(make([]byte, 0x0c))
	/* size of opt. header, 32 bit, exe */
	f.Write([]byte("\xE0\x00\x0F\x03"))

	/* Optional header. */
	f.Write([]byte("\x0B\x01\x00\x00"))

	/* sizeof code & init. data */
	f.Write(int2byte(len(code)))
	f.Write(make([]byte, 0x08))

	/* entry point = 0x1000 */
	f.Write(int2byte(ep))
	/* base of code & data, img base */
	f.Write(int2byte(textBase))
	f.Write(int2byte(textBase))
	f.Write(int2byte(imageBase))

	/* section & file alignment */
	f.Write([]byte("\x00\x10\x00\x00\x00\x02\x00\x00"))
	f.Write(make([]byte, 0x08))
	f.Write([]byte("\x04\x00\x00\x00\x00\x00\x00\x00"))

	/* sizeof image & headers */
	pad := 0x1000 - len(code)%0x1000
	f.Write(int2byte(len(code) + pad + 0x1000))
	f.Write([]byte("\x00\x02\x00\x00\x00\x00\x00\x00"))

	/* subsystem = console */
	f.Write([]byte("\x03\x00\x00\x00"))
	/* stack & heap size */
	f.Write([]byte("\x00\x00\x20\x00\x00\x10\x00\x00"))
	f.Write([]byte("\x00\x00\x10\x00\x00\x10\x00\x00"))
	/* flags & #data dirs */
	f.Write([]byte("\x00\x00\x00\x00\x10\x00\x00\x00"))

	/* data directories */
	f.Write(make([]byte, 0x08))
	f.Write([]byte("\x00\x10\x00\x00"))
	f.Write(int2byte(len(imports)))
	f.Write(make([]byte, 14*8))

	/* Section headers. */
	f.Write([]byte(".dodo\x00\x00\x00"))
	f.Write(int2byte(len(code)))
	f.Write([]byte("\x00\x10\x00\x00"))
	f.Write(int2byte(len(code)))
	f.Write([]byte("\x00\x02\x00\x00"))
	f.Write(make([]byte, 12))
	f.Write([]byte("\x00\x00\x50\xE0"))

	/* padding */
	ndump, err := f.Seek(0, os.SEEK_CUR)
	if err != nil {
		panic(err)
	}
	f.Write(make([]byte, 512-ndump%512))

	/* write section */
	f.Write(code)
}
