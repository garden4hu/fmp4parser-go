package fmp4parser

import "github.com/pkg/errors"

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

	state int
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

		state: StateFtyp,
	}
}

func (p *parser) Append(b []byte) {
	p.Buff.Append(b)
}

func (p *parser) ParseTracks() error {
	for {
		switch p.state {
		case StateFtyp: // parse ftyp box
			{
				p.Buff.ResetReader()
				size, err := p.Buff.FindBox(ftypBox)
				if err != nil {
					goto T
				} else {
					if size > p.Buff.RemainSize() {
						return ErrNoEnoughData
					}
				}
				p.Ftyp = new(FtypBox)
				p.Ftyp.parse(p.Buff)
				p.state = StateMOOV  // Move to moov state
			}
		case StateMOOV:
			{
				p.Buff.ResetReader()
				size, err := p.Buff.FindBox(moovBox)
				if err != nil {
					goto T
				} else {
					if size > p.Buff.RemainSize() + 8 {
						return ErrNoEnoughData
					}
				}
				p.Moov = new(MoovBox)
				err = p.Moov.parse(p.Buff)
			}
		}
	}

T:
	return err
}
