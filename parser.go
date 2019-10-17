package fmp4parser

// internal function of fmp4parser

type parser struct {
	Buff *BufHandler
	Ftyp *FtypBox
	Moov *MoovBox
	Moof []MoofBox
	Mfra []MfraBox
	Styp []StypBox
	Sidx []SidxBox
	SSix []SsixBox

	isHead bool
}

func NewParser() *parser {
	return &parser{
		Buff: NewBufHandler(make([]byte, 0, 0)),
		Ftyp: nil,
		Moov: nil,
		Moof: nil,
		Mfra: nil,
		Styp: nil,
		Sidx: nil,
		SSix: nil,
	}
}

func (p *parser) Append(b []byte) {
	p.Buff.Append(b)
}
