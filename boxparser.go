package fmp4parser

import "github.com/pkg/errors"

// TODO parse boxes

// boxparser provide an interface for parsing different type of box
type boxparser interface {
	parse(r *BufHandler) error
}

// parsing ftyp box
func (p *FtypBox) parse(r *BufHandler) {
	boxSize := r.GetCurrentBoxSize()
	// full box
	p.majorBrand, _ = r.ReadInt()
	p.minorVersion, _ = r.ReadInt()
	boxSize -= 8
	for i := 0; i < boxSize/4; i++ {
		compatibleBrand, _ := r.ReadInt()
		p.compatibleBrands = append(p.compatibleBrands, compatibleBrand)
	}
}

func (p *MoovBox) parse(r *BufHandler) error {
	// record the checkpoint
	boxSize := r.GetCurrentBoxSize()
	anchor := r.GetAbsPos()
	for {
		err = errors.New("")
		boxSize, _ := r.ReadInt()
		boxType, _ := r.ReadInt()
		switch boxType {
		case mvhdBox:
			{
				p.mvhd.parse(r)
			}
		case trafBox:
			{

			}
		case mvexBox:
			{

			}
		case metaBox:
			{

			}
		case psshBox:
			{

			}
		default:
			{

			}
		}
	}
}

func (p *MvhdBox) parse(r *BufHandler) {
	version, _ := r.ReadByte()
	_, _ = r.Move(3)
	if version == 1 {
		p.creationTime, _ = r.ReadLong()
		p.modificationTime, _ = r.ReadLong()
		p.timescale, _ = r.ReadInt()
		p.duration, _ = r.ReadLong()
	} else {
		tmpCreationTime, _ := r.ReadInt()
		p.creationTime = uint64(tmpCreationTime)
		tmpModificationTime, _ := r.ReadInt()
		p.modificationTime = uint64(tmpModificationTime)
		p.timescale, _ = r.ReadInt()
		tmpDduration, _ := r.ReadInt()
		p.duration = uint64(tmpDduration)
	}
	_, _ = r.Move(70)
	p.nextTrackId, _ = r.ReadInt()
}
