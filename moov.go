package fmp4parser

import "errors"

// parsing ftyp box
func (p *boxFtyp) parse(r *bufferHandler) {
	boxSize, _ := r.GetCurrentBoxHeaderInfo()
	// full box
	p.majorBrand = r.Read4()
	p.minorVersion = r.Read4()
	boxSize -= 8
	for i := 0; i < boxSize/4; i++ {
		p.compatibleBrands = append(p.compatibleBrands, r.Read4())
	}
}

// parse moov box
func (p *boxMoov) parse(r *bufferHandler) error {
	// record the checkpoint
	moovBoxSize, _ := r.GetCurrentBoxHeaderInfo()
	anchor := r.Position()
	endPosition := anchor + moovBoxSize
	var latestBoxSize = 0
	err := errors.New("")
	for {
		if r.Position() >= endPosition {
			logs.info.Println(" end of moov box")
			break
		}
		_ = r.MoveTo(latestBoxSize + anchor)
		anchor = r.Position()

		boxSize := r.Read4()
		boxType := r.Read4()
		latestBoxSize = int(boxSize)
		switch boxType {
		case mvhdBox:
			{
				mvhdbox := new(boxMvhd)
				mvhdbox.parse(r)
				p.mvhd = mvhdbox
			}
		case trakBox:
			{
				trackbox := new(boxTrak)
				trackbox.parse(r)
				p.tracks = append(p.tracks, *trackbox)

			}
		case mvexBox:
			{
				mvexbox := new(boxMvex)
				mvexbox.parse(r)
				p.Mvex = mvexbox
			}
		default:
			{
				r.Move(int64(boxSize - 8))
			}
		}
	}
	return err
}
