package fmp4parser

import (
	"errors"
	"math"
)

// boxparser provide an interface for parsing different type of box
// type boxparser interface {
// 	parse(r *bufferHandler) error
// }

// parse mvhd box
func (p *boxMvhd) parse(r *bufferHandler) {
	version := r.ReadByte()
	r.Move(3)
	if version == 1 {
		p.creationTime = r.Read8()
		p.modificationTime = r.Read8()
		p.timescale = r.Read4()
		p.duration = r.Read8()
	} else {
		p.creationTime = uint64(r.Read4())
		p.modificationTime = uint64(r.Read4())
		p.timescale = r.Read4()
		p.duration = uint64(r.Read4())
	}
	r.Move(70)
	p.nextTrackId = r.Read4()
}

// parse mvex box
func (p *boxMvex) parse(r *bufferHandler) {
	// record the checkpoint
	mvexBoxSize, _ := r.GetCurrentBoxHeaderInfo()
	anchor := r.Position()
	endPosition := anchor + mvexBoxSize
	var latestBoxSize = uint32(0)
	for {
		if r.Position() >= endPosition {
			logs.info.Println(" end of mvex box")
			break
		}
		_ = r.MoveTo(int64(latestBoxSize) + anchor)
		anchor = r.Position()

		boxSize := r.Read4()
		boxType := r.Read4()
		latestBoxSize = boxSize
		switch boxType {
		case mehdBox:
			{
				// read mehd box
				version := r.ReadByte()
				r.Move(3)
				if version == 1 {
					p.mehdbox.fragmentDuration = r.Read8()
				} else {
					p.mehdbox.fragmentDuration = uint64(r.Read4())
				}
			}
		case trexBox:
			{
				trexbox := new(boxTrex)
				r.Move(4) // 1 byte version, 3 bites 0
				trexbox.trackId = r.Read4()
				trexbox.defaultSampleDescriptionIndex = r.Read4()
				trexbox.defaultSampleDuration = r.Read4()
				trexbox.defaultSampleSize = r.Read4()
				trexbox.defaultSampleFlags = r.Read4()
				p.trexbox = append(p.trexbox, *trexbox)
			}
		case psshBox:
			{
				psshbox := new(boxPssh)
				version := r.ReadByte()
				r.Move(3)
				psshbox.systemId, _, _ = r.ReadBytes(16)
				if version > 0 {
					psshbox.kIdCount = r.Read4()
					for i := 0; i < int(psshbox.kIdCount); i++ {
						tmpKId, _, _ := r.ReadBytes(16)
						psshbox.kId = append(psshbox.kId, tmpKId)
					}
				}
				psshbox.dataSize = r.Read4()
				psshbox.data, _, _ = r.ReadBytes(int(psshbox.dataSize))
				p.pssh = append(p.pssh, *psshbox)
				logs.info.Println(" find pssh box in the mvex box (Container : moov)")
			}
		default:
			r.Move(int64(boxSize - 8))
		}
	}
}

// parse trak box
func (p *boxTrak) parse(r *bufferHandler) {
	// record the checkpoint
	trakBoxSize, _ := r.GetCurrentBoxHeaderInfo()
	anchor := r.Position()
	endPosition := anchor + trakBoxSize
	var latestBoxSize = uint32(0)
	for {
		if r.Position() >= endPosition {
			logs.info.Println(" end of trak box")
			break
		}
		_ = r.MoveTo(int64(latestBoxSize) + anchor)
		anchor = r.Position()
		boxSize := r.Read4()
		boxType := r.Read4()
		latestBoxSize = boxSize
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
			r.Move(int64(boxSize - 8))
		}
	}
}

// parse tkhd box
func (p *boxTkhd) parse(r *bufferHandler) {
	version := r.ReadByte()
	r.Move(3)
	if version == 1 {
		p.creationTime = r.Read8()
		p.modificationTime = r.Read8()
		p.trackId = r.Read4()
		r.Move(4)
		p.duration = r.Read8()
	} else {
		p.creationTime = uint64(r.Read4())
		p.modificationTime = uint64(r.Read4())
		p.trackId = r.Read4()
		r.Move(4)
		p.duration = uint64(r.Read4())
	}
	r.Move(12) // unsigned int(32)[2] reserved = 0; int(16) layer = 0; int(16) alternate_group = 0;
	p.volume = r.Read2()
	r.Move(2)  // reserved = 0
	r.Move(36) // matrix= { 0x00010000,0,0,0,0x00010000,0,0,0,0x40000000 };
	p.width = r.Read4()
	p.hight = r.Read4()
}

// parse mdia box
func (p *boxMdia) parse(r *bufferHandler) {
	mdiaBoxSize, _ := r.GetCurrentBoxHeaderInfo()
	anchor := r.Position()
	endPosition := anchor + mdiaBoxSize
	var latestBoxSize = uint32(0)
	for {
		if r.Position() >= endPosition {
			logs.info.Println(" end of trak box")
			break
		}
		_ = r.MoveTo(int64(latestBoxSize) + anchor)
		anchor = r.Position()

		boxSize := r.Read4()
		boxType := r.Read4()
		latestBoxSize = boxSize
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
			r.Move(int64(boxSize - 8))
		}
	}
}

// parse mdhd box
func (p *boxMdhd) parse(r *bufferHandler) {
	version := r.ReadByte()
	r.Move(3)
	if version == 1 {
		p.creationTime = r.Read8()
		p.modificationTime = r.Read8()
		p.timeScale = r.Read4()
		p.duration = r.Read8()
	} else {
		p.creationTime = uint64(r.Read4())
		p.modificationTime = uint64(r.Read4())
		p.timeScale = r.Read4()
		p.duration = uint64(r.Read4())
	}
	p.language, _, _ = r.ReadBytes(2)
	r.Move(2)
}

// parse hdlr box
func (p *boxHdlr) parse(r *bufferHandler) {
	hdlrBoxSize, _ := r.GetCurrentBoxHeaderInfo()
	r.Move(4) // version flags
	r.Move(4) // pre_defined
	p.handlerType = r.Read4()
	r.Move(12)
	name, _, _ := r.ReadBytes(hdlrBoxSize - 24)
	p.name = string(name)
}

// parse minf box
func (p *boxMinf) parse(r *bufferHandler) {
	minfBoxSize, _ := r.GetCurrentBoxHeaderInfo()
	anchor := r.Position()
	endPosition := anchor + minfBoxSize
	var latestBoxSize = uint32(0)
	for {
		if r.Position() >= endPosition {
			logs.info.Println(" end of trak box")
			break
		}
		_ = r.MoveTo(int64(latestBoxSize) + anchor)
		anchor = r.Position()

		boxSize := r.Read4()
		boxType := r.Read4()
		latestBoxSize = boxSize
		switch boxType {
		case stblBox:
			{

			}
		default:
			r.Move(int64(boxSize - 8))
		}
	}
}

func (p *boxStbl) parse(r *bufferHandler) {
	stblBoxSize, _ := r.GetCurrentBoxHeaderInfo()
	anchor := r.Position()
	endPosition := anchor + stblBoxSize
	var latestBoxSize = uint32(0)
	for {
		if r.Position() >= endPosition {
			logs.info.Println(" end of trak box")
			break
		}
		_ = r.MoveTo(int64(latestBoxSize) + anchor)
		anchor = r.Position()
		boxSize := r.Read4()
		boxType := r.Read4()
		latestBoxSize = boxSize
		switch boxType {
		case stsdBox:
			{
				stsdbox := new(boxStsd)
				stsdbox.parse(r)
				p.stsd = stsdbox
			}
		default:
			r.Move(int64(boxSize - 8))
		}
	}
}

func (p *boxStsd) parse(r *bufferHandler) {
	stsdBoxSize, _ := r.GetCurrentBoxHeaderInfo()
	anchor := r.Position()
	// endPosition := anchor + stsdBoxSize
	version := r.ReadByte()
	r.Move(3)
	entryCount := r.Read4()

	if entryCount <= 0 || entryCount >= uint32(stsdBoxSize)/8 {
		logs.err.Println(" invalid stsd entry data")
		return
	}
	latestBoxSize := uint32(0)
	anchor = r.Position()
	for i := 0; i < int(entryCount); i++ {
		_ = r.MoveTo(int64(latestBoxSize) + anchor)
		anchor = r.Position()
		entrySize := r.Read4()
		entryType := r.Read4()
		latestBoxSize = entrySize
		trackType := getTrackType(entryType)
		switch trackType {
		case audioTrack:
			_ = p.parseAudioSampleEntry(r)
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

func (p *boxStsd) parseAudioSampleEntry(r *bufferHandler) error {
	r.Move(-8)
	anchor := r.Position()
	entrySize := r.Read4()
	entryType := r.Read4()
	r.Move(8)
	version := r.Read2()
	r.Move(6)
	audioEntry := new(audioSampleEntry)
	audioEntry.codecId = int(entryType)
	if version == 0 || version == 1 {
		audioEntry.channelCount = int(r.Read2()) // 2 bytes
		audioEntry.sampleSize = int(r.Read2())   // 2bytes
		r.Move(4)                                // 2 bytes + 2 bytes (compressionID + packetsize)
		audioEntry.sampleRate = int(r.Read2())
		if audioEntry.sampleRate == 0 {
			audioEntry.sampleRate = int(r.Read2())
		} else {
			r.Move(2)
		}
		if version == 1 {
			audioEntry.qttf = true
			audioEntry.qttfVersion = 1
			audioEntry.qttfSamplesPerPacket = int(r.Read4())
			audioEntry.qttfBytesPerPacket = int(r.Read4())
			audioEntry.qttfBytesPerFrame = int(r.Read4())
			audioEntry.qttfBytesPerSample = int(r.Read4())
		}
	}
	if version == 2 {
		audioEntry.qttf = true
		audioEntry.qttfVersion = 2
		r.Move(16) // always[3,16,Minus2,0,65536], sizeOfStructOnly
		tmpSampleRate := r.Read8()
		audioEntry.sampleRate = int(math.Round(float64(tmpSampleRate)))
		audioEntry.channelCount = int(r.Read2()) // 2 bytes
		r.Move(4)                                // always 0x7F000000
		constBitsPerChannel := int(r.Read4())    //	constBitsPerChannel 4 bytes
		flags := int(r.Read4())
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
	} else {
		_ = r.MoveTo(anchor + int64(entrySize))
		logs.err.Println("unsupported version")
		return ErrUnsupportedSampleEntry
	}
	absPosition := r.Position()
	if entryType == encaSampleEntry {
		sinfSize, err := r.FindBoxInterval(sinfBox, uint32(r.Position()-anchor))
		if err != nil || sinfSize < 8 {
			logs.info.Println("error when dealing with encrypt audio box 'enca' : no sinf box ")
			return ErrIncompleteCryptoBox
		}
		_ = r.MoveTo(absPosition)
		curBoxStartPos := r.Position()
		latestBoxSize := uint32(0)
		for {
			if r.Position() >= anchor+int64(entrySize) {
				break
			}
			_ = r.MoveTo(curBoxStartPos + int64(latestBoxSize))
			curBoxStartPos = r.Position()
			boxSize := r.Read4()
			boxType := r.Read4()
			latestBoxSize = boxSize
			switch boxType {
			case sinfBox:
				sinf := new(boxSinf)
				_ = sinf.parse(r)
				audioEntry.enca.sinf = append(audioEntry.enca.sinf, *sinf)
			default:
				if boxSize < 8 {
					return ErrIncompleteBox
				}
			}
		}
	}
	_ = r.MoveTo(absPosition)
	curBoxStartPos := r.Position()
	latestBoxSize := uint32(0)
	for {
		if r.Position() >= anchor+int64(entrySize) {
			break
		}
		_ = r.MoveTo(curBoxStartPos + int64(latestBoxSize))
		curBoxStartPos = r.Position()
		boxSize := r.Read4()
		boxType := r.Read4()
		latestBoxSize = boxSize
		switch boxType {
		case esdsBox:
			esds := new(esdsDescriptors)
			_ = esds.parseDescriptors(r)
			audioEntry.codecId = esds.audioCodec
			audioEntry.descriptor = esds
		case waveBox:
			{
				if audioEntry.qttf != true {
					break
				}
				_, err := r.FindBoxInterval(esdsBox, boxSize)
				if err != nil {
					break
				}
				esds := new(esdsDescriptors)
				_ = esds.parseDescriptors(r)
				audioEntry.codecId = esds.audioCodec
				audioEntry.descriptor = esds
			}
		case aacBox:
		}

	}

	return nil
}

func (p *boxStsd) parseVideoSampleEntry(r *bufferHandler, entrySize uint32) {

}

func (p *boxStsd) parseSubtitleSampleEntry(r *bufferHandler, entrySize uint32) {

}

func (p *boxSinf) parse(r *bufferHandler) error {
	// record the checkpoint
	sinfBoxSize, _ := r.GetCurrentBoxHeaderInfo()
	anchor := r.Position()
	endPosition := anchor + sinfBoxSize
	var latestBoxSize = uint32(0)
	err := errors.New("")
	var schiSize = uint32(0)
	var schiPos = int64(0)
	for {
		if r.Position() >= anchor+int64(sinfBoxSize) {
			logs.info.Println(" end of moov box")
			break
		}
		_ = r.MoveTo(int64(latestBoxSize) + anchor)
		anchor = r.Position()
		boxSize := r.Read4()
		boxType := r.Read4()
		latestBoxSize = boxSize
		switch boxType {
		case frmaBox:
			p.codingName = r.Read4()
		case schmBox:
			r.Move(4)
			p.schemeType = r.Read4()
			p.schemeVersion = r.Read4()
			r.Move(int64(boxSize) - (r.Position() - curPosition))
		case schiBox:
			schiPos = r.Position()
			schiSize = boxSize - 8
			break
		default:
		}
	}
	/*
			    protection schemes: (ref: ISO/IEC 23001-7)
		    	'cenc' 0x63656e63 (le)
			    'cbc1' 0x63626331 (le)
			    'cens' 0x63656e73 (le)
			    'cbcs' 0x63626373 (le)
	*/
	if p.schemeType == 0x63656e63 || p.schemeType == 0x63626331 ||
		p.schemeType == 0x63656e73 || p.schemeType == 0x63626373 {
		_ = r.MoveTo(schiPos)
		anchor = r.Position()
		latestBoxSize = uint32(0)
		for {
			if r.Position() > endPosition {
				break
			}
			_ = r.MoveTo(anchor + int64(latestBoxSize))
			anchor = r.Position()
			boxSize := r.Read4()
			boxType := r.Read4()
			latestBoxSize = boxSize
			switch boxType {
			case tencBox:
				p.tenc.version = int(r.Read4())
				p.tenc.version = 0x000000FF & (version >> 24)
				r.Move(1)
				if p.tenc.version == 0 {
					r.Move(1)
				} else {
					block := r.ReadByte()
					p.tenc.defaultCryptByteBlock = uint8(block&0xF0) >> 4
					p.tenc.defaultSkipByteBlock = uint8(block & 0x0F)
				}
				p.tenc.defaultIsProtected = r.ReadByte()
				p.tenc.defaultPerSampleIVSize = r.ReadByte()
				p.tenc.defaultKID, _, _ = r.ReadBytes(16)
				if p.tenc.defaultIsProtected == 1 && p.tenc.defaultPerSampleIVSize == 0 {
					p.tenc.defaultConstantIVSize = r.ReadByte()
					p.tenc.defaultConstantIV, _, _ = r.ReadBytes(p.tenc.defaultConstantIVSize)
				}
				return nil
			default:
			}
		}
	}
	return nil
}
