package compiler

import (
	"fmt"
	"strconv"
	"strings"
)

type ByteCode byte
type OpCode uint8

const (
	OpNoop OpCode = iota

	OpJmp
	OpJmpIfTrue
	OpJmpIfFalse
	OpCall
	OpCallNative
	OpReturn

	OpPushConst
	OpPushTrue
	OpPushFalse
	OpPushNone
	OpPushArr
	OpPushMap
	OpPop
	OpDup

	OpLoadLocal
	OpStoreLocal
	OpLoadExtern

	OpAdd
	OpSub
	OpMul
	OpDiv
	OpMod
	OpNeg

	OpEq
	OpNeq
	OpLt
	OpGt
	OpLte
	OpGte

	OpAnd
	OpOr
	OpNot
	OpInc
	OpDec

	OpMakeIter
	OpNextIter
)

var codeToStr = map[OpCode]string{
	OpNoop:       "NOOP",
	OpJmp:        "JMP",
	OpJmpIfTrue:  "JMP_IF_TRUE",
	OpJmpIfFalse: "JMP_IF_FALSE",
	OpCall:       "CALL",
	OpCallNative: "CALL_NATIVE",
	OpReturn:     "RETURN",
	OpPushConst:  "PUSH_CONST",
	OpPushTrue:   "PUSH_TRUE",
	OpPushFalse:  "PUSH_FALSE",
	OpPushNone:   "PUSH_NONE",
	OpPushArr:    "PUSH_ARRAY",
	OpPushMap:    "PUSH_MAP",
	OpPop:        "POP",
	OpDup:        "DUP",
	OpLoadLocal:  "LOAD_LOCAL",
	OpLoadExtern: "LOAD_EXTERN",
	OpStoreLocal: "STORE_LOCAL",
	OpAdd:        "ADD",
	OpSub:        "SUB",
	OpMul:        "MUL",
	OpDiv:        "DIV",
	OpMod:        "MOD",
	OpNeg:        "NEG",
	OpEq:         "EQ",
	OpNeq:        "NEQ",
	OpLt:         "LT",
	OpGt:         "GT",
	OpLte:        "LTE",
	OpGte:        "GTE",
	OpAnd:        "AND",
	OpOr:         "OR",
	OpNot:        "NOT",
	OpInc:        "INCREMENT",
	OpDec:        "DECREMENT",
	OpMakeIter:   "MAKE_ITER",
	OpNextIter:   "NEXT_ITER",
}

var operandCountTable = map[OpCode]uint16{
	OpJmp:        1,
	OpJmpIfTrue:  1,
	OpJmpIfFalse: 1,
	OpCall:       2,
	OpCallNative: 2,
	OpPushConst:  1,
	OpLoadLocal:  1,
	OpLoadExtern: 1,
	OpStoreLocal: 1,
	OpPushArr:    1,
	OpPushMap:    1,
	OpReturn:     1,
	OpInc:        1,
	OpDec:        1,
	OpNextIter:   1,
}

// Encodes an OpCode and its operand count into a single ByteCode.
// The upper 15 bits store the opcode, the lowest bit stores the operand count (0 or 1).
func (op OpCode) Pack(operandCount uint16) ByteCode {
	return ByteCode(op)
}

// Return the opcode as a string.
func (op OpCode) String() string {
	if name, ok := codeToStr[op]; ok {
		return name
	}
	return fmt.Sprintf("UNKNOWN(%d)", op)
}

// Unpacks a ByteCode into its OpCode and operand count.
func (b ByteCode) Decode() (OpCode, uint16) {
	op := OpCode(b)
	count, ok := operandCountTable[op]
	if !ok {
		count = 0
	}
	return op, count
}

type Chunk struct {
	code []ByteCode
}

func NewChunk() *Chunk {
	return &Chunk{}
}

// Encodes an instruction (opcode + operands) and appends it to the chunk.
func (c *Chunk) Emit(op OpCode, operands ...ByteCode) int {
	index := len(c.code)
	c.code = append(c.code, op.Pack(uint16(len(operands))))
	c.code = append(c.code, operands...)
	return index
}

// Appends another chunk's bytecode onto this one.
func (c *Chunk) Join(other Chunk) {
	c.code = append(c.code, other.code...)
}

// Read operands from start to (start + count).
func (c Chunk) ReadOperands(start int, count uint16) []ByteCode {
	if count == 0 {
		return []ByteCode{}
	}
	return c.code[start : start+int(count)]
}

// Returns true if the chunk is empty, false otherwise.
func (c Chunk) Empty() bool {
	return len(c.code) == 0
}

// Returns the last valid index, or -1 if the chunk is empty.
func (c Chunk) Last() int {
	if len(c.code) == 0 {
		return -1
	}
	return len(c.code) - 1
}

// Sets the operand of a jump instruction at index to target.
func (c *Chunk) Patch(index int, target int) error {
	if index < 0 || index >= len(c.code) {
		return fmt.Errorf("index %d out of bounds", index)
	}
	if target < 0 || target >= len(c.code) {
		return fmt.Errorf("jump target %d out of bounds", target)
	}

	op, operandCount := c.code[index].Decode()
	if op != OpJmp && op != OpJmpIfFalse && op != OpJmpIfTrue && op != OpNextIter {
		return fmt.Errorf("must patch a jump, got %s", op)
	}
	if operandCount != 1 {
		return fmt.Errorf("jumps must have only one operand")
	}

	c.code[index+1] = ByteCode(target)
	return nil
}

// Sets the operand of a jump instruction to the last index.
func (c *Chunk) PatchToEnd(index int) error {
	if c.Empty() {
		return fmt.Errorf("chunk is empty")
	}
	return c.Patch(index, c.Last())
}

func joinOperands(operands []ByteCode) string {
	var str strings.Builder
	if len(operands) > 0 {
		for i, v := range operands {
			if i != 0 {
				str.WriteString(", ")
			}
			str.WriteString("\x1b[33m")
			str.WriteString(strconv.FormatUint(uint64(v), 10))
			str.WriteString("\x1b[0m")
		}
	}
	return str.String()
}

// Returns the chunk as a string.
func (c Chunk) String() string {
	var str strings.Builder
	// lastI := 0
	str.WriteString("\x1b[90mIndex\tOperation\tOperands\n\x1b[0m")
	for i := 0; i < len(c.code); {
		op, count := c.code[i].Decode()
		operands := c.ReadOperands(i+1, count)
		// if i-lastI > 1 {
		// 	fmt.Fprintf(&str, "(step %d)\n", i-lastI)
		// }
		opStr := op.String()
		fmt.Fprintf(
			&str,
			"\x1b[35m%05d\x1b[0m%s\x1b[32m%s\x1b[0m%s%s\n",
			i,
			"\t",
			opStr,
			strings.Repeat(" ", strDiff(opStr, "Operation"))+"\t",
			joinOperands(operands),
		)
		i += 1 + int(count)
		// lastI = i
	}
	return str.String()
}

func strDiff(a, b string) int {
	al, bl := len(a), len(b)
	if al > bl {
		return al - bl
	} else {
		return bl - al
	}
}

// Returns the chunk as an array of byte.
func (c Chunk) Bytes() []byte {
	buf := make([]byte, len(c.code))
	for i, b := range c.code {
		buf[i] = byte(b)
	}
	return buf
}
