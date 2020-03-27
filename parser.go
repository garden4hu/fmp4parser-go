package fmp4parser

import "errors"

// internal function of fmp4parser

type parser struct {
	Buff *bufferHandler
	Ftyp *boxFtyp
	Moov *boxMoov
	Moof []boxMoof
	Mfra []boxMfra
	Styp []boxStyp
	Sidx []boxSidx
	SSix []boxSsix

	state int
}

func NewParser() *parser {
	return &parser{
		Buff: newBufHandler(make([]byte, 0, 0)),
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
	err := errors.New("")
	for {
		switch p.state {
		case StateFtyp: // parse ftyp box
			{
				p.Buff.Reset()
				size, err := p.Buff.FindBox(ftypBox)
				if err != nil {
					goto T
				} else {
					if size > p.Buff.Remainder() {
						return ErrNoEnoughData
					}
				}
				p.Ftyp = new(boxFtyp)
				p.Ftyp.parse(p.Buff)
				p.state = StateMOOV // Move to moov state
			}
		case StateMOOV:
			{
				p.Buff.Reset()
				size, err := p.Buff.FindBox(moovBox)
				if err != nil {
					goto T
				} else {
					if size > p.Buff.Remainder()+8 {
						return ErrNoEnoughData
					}
				}
				p.Moov = new(boxMoov)
				err = p.Moov.parse(p.Buff)
			}
		}
	}

T:
	return err
}
