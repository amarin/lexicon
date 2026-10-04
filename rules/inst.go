package rules

// inst is one instruction. For opElem, slot is the role index (or -1);
// for opSave it is 2*role (start) or 2*role+1 (end).
type inst struct {
	op   opcode
	sel  *selector
	x, y int
	slot int
}
