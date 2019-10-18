package fmp4parser

import (
	"errors"
)

// boxparser provide an interface for parsing different type of box
// type boxparser interface {
// 	parse(r *BufHandler) error
// }

// parsing ftyp box
func (p *boxFtyp) parse(r *BufHandler) {
	boxSize := r.GetCurrentBoxSize()
	// full box
	p.majorBrand = r.ReadInt()
	p.minorVersion = r.ReadInt()
	boxSize -= 8
	for i := 0; i < boxSize/4; i++ {
		p.compatibleBrands = append(p.compatibleBrands, r.ReadInt())
	}
}

// parse moov box
func (p *boxMoov) parse(r *BufHandler) error {
	// record the checkpoint
	moovBoxSize := r.GetCurrentBoxSize()
	anchor := r.GetAbsPos()
	err := errors.New("")
	for {
		if r.GetAbsPos() >= anchor+int64(moovBoxSize) {
			logs.info.Println(" end of moov box")
			break
		}
		boxSize := r.ReadInt()
		boxType := r.ReadInt()
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
				_, _ = r.Move(int64(boxSize - 8))
			}
		}
	}
	return err
}

// parse mvhd box
func (p *boxMvhd) parse(r *BufHandler) {
	version := r.ReadByte()
	_, _ = r.Move(3)
	if version == 1 {
		p.creationTime = r.ReadLong()
		p.modificationTime = r.ReadLong()
		p.timescale = r.ReadInt()
		p.duration = r.ReadLong()
	} else {
		p.creationTime = uint64(r.ReadInt())
		p.modificationTime = uint64(r.ReadInt())
		p.timescale = r.ReadInt()
		p.duration = uint64(r.ReadInt())
	}
	_, _ = r.Move(70)
	p.nextTrackId = r.ReadInt()
}

// parse mvex box
func (p *boxMvex) parse(r *BufHandler) {
	// record the checkpoint
	mvexBoxSize := r.GetCurrentBoxSize()
	anchor := r.GetAbsPos()
	for {
		if r.GetAbsPos() >= anchor+int64(mvexBoxSize) {
			logs.info.Println(" end of mvex box")
			break
		}
		boxSize := r.ReadInt()
		boxType := r.ReadInt()
		switch boxType {
		case mehdBox:
			{
				// read mehd box
				version := r.ReadByte()
				_, _ = r.Move(3)
				if version == 1 {
					p.mehdbox.fragmentDuration = r.ReadLong()
				} else {
					p.mehdbox.fragmentDuration = uint64(r.ReadInt())
				}
			}
		case trexBox:
			{
				trexbox := new(boxTrex)
				_, _ = r.Move(4) // 1 byte version, 3 bites 0
				trexbox.trackId = r.ReadInt()
				trexbox.defaultSampleDescriptionIndex = r.ReadInt()
				trexbox.defaultSampleDuration = r.ReadInt()
				trexbox.defaultSampleSize = r.ReadInt()
				trexbox.defaultSampleFlags = r.ReadInt()
				p.trexbox = append(p.trexbox, *trexbox)
			}
		case psshBox:
			{
				psshbox := new(boxPssh)
				version := r.ReadByte()
				_, _ = r.Move(3)
				psshbox.systemId, _, _ = r.ReadBytes(16)
				if version > 0 {
					psshbox.kIdCount = r.ReadInt()
					for i := 0; i < int(psshbox.kIdCount); i++ {
						tmpKId, _, _ := r.ReadBytes(16)
						psshbox.kId = append(psshbox.kId, tmpKId)
					}
				}
				psshbox.dataSize = r.ReadInt()
				psshbox.data, _, _ = r.ReadBytes(int(psshbox.dataSize))
				p.pssh = append(p.pssh, *psshbox)
				logs.info.Println(" find pssh box in the mvex box (Container : moov)")
			}
		default:
			_, _ = r.Move(int64(boxSize - 8))
		}
	}
}

// parse trak box
func (p *boxTrak) parse(r *BufHandler) {
	// record the checkpoint
	trakBoxSize := r.GetCurrentBoxSize()
	anchor := r.GetAbsPos()
	for {
		if r.GetAbsPos() >= anchor+int64(trakBoxSize) {
			logs.info.Println(" end of trak box")
			break
		}
		boxSize := r.ReadInt()
		boxType := r.ReadInt()
		switch boxType {
		case tkhdBox:
			{
				tkhdbox := new(boxTkhd)
				tkhdbox.parse(r)
				p.tkhd = tkhdbox
			}
		case mdiaBox:
			{
				mdiabox := new(boxMdia)
				mdiabox.parse(r)
				p.mdia = mdiabox
			}
		default:
			_, _ = r.Move(int64(boxSize - 8))
		}
	}
}

// parse tkhd box
func (p *boxTkhd) parse(r *BufHandler) {
	version := r.ReadByte()
	_, _ = r.Move(3)
	if version == 1 {
		p.creationTime = r.ReadLong()
		p.modificationTime = r.ReadLong()
		p.trackId = r.ReadInt()
		_, _ = r.Move(4)
		p.duration = r.ReadLong()
	} else {
		p.creationTime = uint64(r.ReadInt())
		p.modificationTime = uint64(r.ReadInt())
		p.trackId = r.ReadInt()
		_, _ = r.Move(4)
		p.duration = uint64(r.ReadInt())
	}
	_, _ = r.Move(12) // unsigned int(32)[2] reserved = 0; int(16) layer = 0; int(16) alternate_group = 0;
	p.volume = r.ReadShort()
	_, _ = r.Move(2)  // reserved = 0
	_, _ = r.Move(36) // matrix= { 0x00010000,0,0,0,0x00010000,0,0,0,0x40000000 };
	p.width = r.ReadInt()
	p.hight = r.ReadInt()
}

// parse mdia box
func (p *boxMdia) parse(r *BufHandler) {
	mdiaBoxSize := r.GetCurrentBoxSize()
	anchor := r.GetAbsPos()
	for {
		if r.GetAbsPos() >= anchor+int64(mdiaBoxSize) {
			logs.info.Println(" end of trak box")
			break
		}
		boxSize := r.ReadInt()
		boxType := r.ReadInt()
		switch boxType {
		case mdhdBox:
			{
				mdhdbox := new(boxMdhd)
				mdhdbox.parse(r)
				p.mdhd = mdhdbox
			}
		case hdlrBox:
			{
				hdlrbox := new(boxHdlr)
				hdlrbox.parse(r)
				p.hldr = hdlrbox
			}
		case minfBox:
			{
				minfbox := new(boxMinf)
				minfbox.parse(r)
				p.minf = minfbox

			}
		default:
			_, _ = r.Move(int64(boxSize - 8))
		}
	}
}

// parse mdhd box
func (p *boxMdhd) parse(r *BufHandler) {
	version := r.ReadByte()
	_, _ = r.Move(3)
	if version == 1 {
		p.creationTime = r.ReadLong()
		p.modificationTime = r.ReadLong()
		p.timeScale = r.ReadInt()
		p.duration = r.ReadLong()
	} else {
		p.creationTime = uint64(r.ReadInt())
		p.modificationTime = uint64(r.ReadInt())
		p.timeScale = r.ReadInt()
		p.duration = uint64(r.ReadInt())
	}
	p.language, _, _ = r.ReadBytes(2)
	_, _ = r.Move(2)
}

// parse hdlr box
func (p *boxHdlr) parse(r *BufHandler) {
	hdlrBoxSize := r.GetCurrentBoxSize()
	_, _ = r.Move(4) // version flags
	_, _ = r.Move(4) // pre_defined
	p.handlerType = r.ReadInt()
	_, _ = r.Move(12)
	name, _, _ := r.ReadBytes(hdlrBoxSize - 24)
	p.name = string(name)
}

// parse minf box
func (p *boxMinf) parse(r *BufHandler) {
	mdiaBoxSize := r.GetCurrentBoxSize()
	anchor := r.GetAbsPos()
	for {
		if r.GetAbsPos() >= anchor+int64(mdiaBoxSize) {
			logs.info.Println(" end of trak box")
			break
		}
		boxSize := r.ReadInt()
		boxType := r.ReadInt()
		switch boxType {
		case stblBox:
			{

			}
		default:
			_, _ = r.Move(int64(boxSize - 8))
		}
	}
}

func (p *boxStbl) parse(r *BufHandler) {
	mdiaBoxSize := r.GetCurrentBoxSize()
	anchor := r.GetAbsPos()
	for {
		if r.GetAbsPos() >= anchor+int64(mdiaBoxSize) {
			logs.info.Println(" end of trak box")
			break
		}
		boxSize := r.ReadInt()
		boxType := r.ReadInt()
		switch boxType {
		case stsdBox:{
			stsdbox := new(boxStsd)
			stsdbox.parser(r)
			p.stsd = stsdbox
		}
		default:
			_, _ = r.Move(int64(boxSize - 8))
		}
	}
}

func (p *boxStsd) parser(r *BufHandler){

}