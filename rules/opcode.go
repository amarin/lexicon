package rules

// opcode is an instruction kind of a compiled pattern.
type opcode uint8

const (
	opMatch opcode = iota // accept
	opElem                // consume one term or one span
	opSplit               // try x, then y
	opJmp                 // continue at x
	opSave                // record the position in capture slot
)
