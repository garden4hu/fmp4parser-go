package fmp4parser

// Fmp4Parser is a object type for fmp4parser
type Fmp4Parser struct {
	obj *parser
}

// NewFmp4Parser return a pointer to a variable of Fmp4Parser
func NewFmp4Parser() *Fmp4Parser {
	return &Fmp4Parser{obj: NewParser()}
}

func (h *Fmp4Parser) Process(rawdata []byte) error {
	if len(rawdata) == 0 {
		return ErrNoEnoughData
	}
	h.obj.Append(rawdata)
	// TODO
	return nil
}
