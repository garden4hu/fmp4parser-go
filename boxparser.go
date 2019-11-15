package fmp4parser

import (
	"errors"
	"math"
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
		case stsdBox:
			{
				stsdbox := new(boxStsd)
				stsdbox.parser(r)
				p.stsd = stsdbox
			}
		default:
			_, _ = r.Move(int64(boxSize - 8))
		}
	}
}

func (p *boxStsd) parser(r *BufHandler) {
	stsdBoxSize := r.GetCurrentBoxSize()
	anchor := r.GetAbsPos()
	version := r.ReadByte()
	_, _ = r.Move(3)
	entryCount := r.ReadInt()

	if entryCount <= 0 || entryCount >= uint32(stsdBoxSize)/8 {
		_ = r.SetPos(anchor + stsdBoxSize)
		logs.err.Println(" invalid stsd entry data")
		return
	}

	for i := 0; i < int(entryCount); i++ {
		entrySize := r.ReadInt()
		entryType := r.ReadInt()
		trackType := getTrackType(entryType)
		switch trackType {
		case audioTrack:
			p.parseAudioSampleEntry(r)
		case videoTrak:
			p.parseVideoSampleEntry(r, entrySize)
		case subtitleTrack:
			p.parseSubtitleSampleEntry(r, entrySize)
		default:
			{
				logs.info.Println("unsupported sample entry")
			}
		}

	}

}

func (p *boxStsd) parseAudioSampleEntry(r *BufHandler) error {
	r.Move(-8)
	anchor := r.GetAbsPos()
	entrySize := r.ReadInt()
	entryType := r.ReadInt()
	r.Move(8)
	version := r.ReadShort()
	r.Move(6)
	audioEntry := new(audioSampleEntry)
	audioEntry.codecId = entryType
	if version == 0 || version == 1 {
		audioEntry.channelCount = int(r.ReadShort()) // 2 bytes
		audioEntry.sampleSize = int(r.ReadShort())   // 2bytes
		r.Move(4)                                    // 2 bytes + 2 bytes (compressionID + packetsize)
		audioEntry.sampleRate = int(r.ReadShort())
		if audioEntry.sampleRate == 0 {
			audioEntry.sampleRate = int(r.ReadShort())
		} else {
			r.Move(2)
		}
		if version == 1 {
			audioEntry.qttf = true
			audioEntry.qttfVersion = 1
			audioEntry.qttfSamplesPerPacket = int(r.ReadInt())
			audioEntry.qttfBytesPerPacket = int(r.ReadInt())
			audioEntry.qttfBytesPerFrame = int(r.ReadInt())
			audioEntry.qttfBytesPerSample = int(r.ReadInt())
		}
	}
	if version == 2 {
		audioEntry.qttf = true
		audioEntry.qttfVersion = 2
		r.Move(16) // always[3,16,Minus2,0,65536], sizeOfStructOnly
		tmpSampleRate := r.ReadLong()
		audioEntry.sampleRate = int(math.Round(float64(tmpSampleRate)))
		audioEntry.channelCount = int(r.ReadShort()) // 2 bytes
		r.Move(4)                                    // always 0x7F000000
		constBitsPerChannel := int(r.ReadInt())      //	constBitsPerChannel 4 bytes
		flags := int(r.ReadInt())
		r.Move(8) //	constBytesPerAudioPacket(32-bit) + constLPCMFramesPerAudioPacket(32-bit)
		if entryType == lpcmSampleEntry {
			// The way to deal with "lpcm" comes from ffmpeg. Very thanks
			codec := func(bps int, flags int) lpcmCodecId {
				flt := flags & 1
				be := flags & 2
				sflags := 0
				if (flags & 4) != 0 {
					sflags = -1
				}
				if bps <= 0 || bps > 64 {
					return None
				}
				if flt != 0 {
					switch bps {
					case 32:
						if be == 0 {
							return pcmF32LE
						}
						return pcmF32BE
					case 64:
						if be == 0 {
							return pcmF64LE
						}
						return pcmF64BE
					default:
						return None
					}
				} else {
					bps += 7
					bps >>= 3
					if sflags & (1 << (bps - 1)) {
						switch bps {
						case 1:
							return pcmS8
						case 2:
							if be == 0 {
								return pcmS16LE
							}
							return pcmS16BE
						case 3:
							if be == 0 {
								return pcmS24LE
							}
							return pcmS24BE
						case 4:
							if be == 0 {
								return pcmS32LE
							}
							return pcmS32BE
						case 8:
							if be == 0 {
								return pcmS64LE
							}
							return pcmS64BE
						default:
							return None
						}
					} else {
						switch bps {
						case 1:
							return pcmU8
						case 2:
							if be == 0 {
								return pcmU16LE
							}
							return pcmU16BE
						case 3:
							if be == 0 {
								return pcmU24LE
							}
							return pcmU24BE
						case 4:
							if be == 0 {
								return pcmU32LE
							}
							return pcmU32BE
						default:
							return None
						}
					}

				}
			}(constBitsPerChannel, flags)
			switch codec {
			case pcmS8:
				fallthrough
			case pcmU8:
				if constBitsPerChannel == 16 {
					codec = pcmS16BE
				}
				break
			case pcmS16LE:
				fallthrough
			case pcmS16BE:
				if constBitsPerChannel == 8 {
					codec = pcmS8
				} else if constBitsPerChannel == 24 {
					if codec == pcmS16BE {
						codec = pcmS24BE
					} else {
						codec = pcmS24LE
					}
				} else if constBitsPerChannel == 32 {
					if codec == pcmS16BE {
						codec = pcmS32BE
					} else {
						codec = pcmS32LE
					}
				}
			default:
			}
		}
		bitsPerSample := func(codec lpcmCodecId) int {
			switch codec {
			case pcmALaw:
				fallthrough
			case pcmMULaw:
				fallthrough
			case pcmVIDC:
				fallthrough
			case pcmS8:
				fallthrough
			case pcmS8Planar:
				fallthrough
			case pcmU8:
				fallthrough
			case pcmZORK:
				return 8

			case pcmS16BE:
				fallthrough
			case pcmS16BEPlanar:
				fallthrough
			case pcmS16LE:
				fallthrough
			case pcmS16LEPlanar:
				fallthrough
			case pcmU16BE:
				fallthrough
			case pcmU16LE:
				return 16
			case pcmS24DAUD:
				fallthrough
			case pcmS24BE:
				fallthrough
			case pcmS24LE:
				fallthrough
			case pcmS24LEPlanar:
				fallthrough
			case pcmU24BE:
				fallthrough
			case pcmU24LE:
				return 24
			case pcmS32BE:
				fallthrough
			case pcmS32LE:
				fallthrough
			case pcmS32LEPlanar:
				fallthrough
			case pcmU32BE:
				fallthrough
			case pcmU32LE:
				fallthrough
			case pcmF32BE:
				fallthrough
			case pcmF32LE:
				fallthrough
			case pcmF24LE:
				fallthrough
			case pcmF16LE:
				return 32
			case pcmF64BE:
				fallthrough
			case pcmF64LE:
				fallthrough
			case pcmS64BE:
				fallthrough
			case pcmS64LE:
				return 64
			default:
				return 0
			}
		}(codec)
		if bitsPerSample != 0 {
			audioEntry.qttfBytesPerSample = bitsPerSample
		}
	}else {
		_ = r.SetPos(anchor + int64(entrySize))
		logs.err.Println("unsupported version")
		return ErrUnsupportedSampleEntry
	}

	if entryType == encaSampleEntry {
		sinfSize, err := r.FindBoxInterval(sinfBox,uint32(r.GetAbsPos() - anchor))
		if err != nil {
			_ = r.SetPos(anchor + int64(entrySize))
			logs.info.Println("error when dealing with encrypt audio box 'enca' : no sinf box ")
			return ErrIncompleteCryptoBox
		}


	}


	return nil
}

func (p *boxStsd) parseVideoSampleEntry(r *BufHandler, entrySize uint32) {

}

func (p *boxStsd) parseSubtitleSampleEntry(r *BufHandler, entrySize uint32) {

}
