package fmp4parser

// The status parameter represents the top level box currently being processed
const (
	StateFtyp  int = iota // value --> 0
	StateMOOV            // value --> 1
	StateMOOF            // value --> 2
)
