package fmp4parser

import (
	"errors"
	"io"
)

// parse tkhd box
func (p *boxTrak) parseTkhd(r *atomReader) {
	version, flags := r.ReadVersionFlags()
	p.trackEnabled = flags&0x00000001 != 0
	// p.flagTrackInMovie = flags & 0x00000002 != 0
	// p.flagTrackInPreview = flags & 0x00000004 != 0
	p.flagTrackSizeIsAspectRatio = flags&0x00000008 != 0
	if version == 1 {
		p.creationTime = r.Read8()
		p.modificationTime = r.Read8()
		p.id = r.Read4() // track id
		_ = r.Move(4)    // reversed 0
		p.duration = r.Read8()
	} else {
		p.creationTime = uint64(r.Read4())
		p.modificationTime = uint64(r.Read4())
		p.id = r.Read4()
		_ = r.Move(4) // reversed 0
		p.duration = uint64(r.Read4())
	}
	_ = r.Move(12) // reversed, layer, alternate_group 0 8+4 bytes
	_ = r.Move(2)  // Volume {if track_is_audio 0x0100 else 0}
	_ = r.Move(2)  // reserved = 0
	_ = r.Move(36) // matrix= { 0x00010000,0,0,0,0x00010000,0,0,0,0x40000000 };
	p.width = r.Read4()
	p.height = r.Read4()
}

// parse edts box
func (p *boxTrak) parseEdts(r *atomReader) {
	edts := new(boxEdts)
	_ = r.Move(4) // fourCCelts
	version, _ := r.ReadVersionFlags()
	edts.entryCount = r.Read4()
	for i := uint32(0); i < edts.entryCount; i++ {
		if version == 1 {
			edts.editDuration = append(edts.editDuration, r.Read8())
			edts.mediaTime = append(edts.mediaTime, r.Read8S())
		} else {
			edts.editDuration = append(edts.editDuration, uint64(r.Read4()))
			edts.mediaTime = append(edts.mediaTime, int64(r.Read4S()))
		}
		edts.mediaRateInteger = append(edts.mediaRateInteger, r.Read2S())
		_ = r.Move(2) // media_rate_fraction == 0
	}
	p.edts = edts
}

// parse google spatial media. Extra
func parseUuid(p *MovieInfo, r *atomReader) error {
	// check if is spatial-media ref: https://github.com/google/spatial-media
	if r.a.Size() < 16 {
		return nil
	}
	sphericalMedia := r.Read4() == 0xffcc8263 && r.Read4() == 0xf8554a93 && r.Read4() == 0x8814587a && r.Read4() == 0x02521fdd
	if sphericalMedia {
		logD.Println("This movie is a spatial media")
		rdfData := make([]byte, r.a.Size()-16)
		_, _ = r.ReadBytes(rdfData)
		logD.Println(string(rdfData))
	}
	return nil
}

// parse trak/mdia box
func (p *boxTrak) parseMdia(r *atomReader) error {
	hdlrAtom, err := r.FindSubAtom(fourCChdlr) // get track type
	if err != nil {
		return err
	}
	p.trackType = p.parseHdlr(hdlrAtom)
	for {
		ar, e := r.GetNextAtom()
		if e != nil {
			if e == ErrEndOfAtom {
				return nil
			}
			if e == ErrBadAtom {
				return e
			}
		}
		if ar.a.atomType == fourCCmdhd {
			p.parseMdhd(ar)
		} else if ar.a.atomType == fourCCminf {
			err = p.parseMinf(ar)
			if err != nil {
				return err
			}
		} else {
			continue
		}
	}
}

// parse trak/mdia/mdhd box
func (p *boxTrak) parseMdhd(r *atomReader) {
	version, _ := r.ReadVersionFlags()
	if version == 1 {
		p.creationTime = r.Read8()
		p.modificationTime = r.Read8()
		p.timeScale = r.Read4()
		p.duration = r.Read8()
	} else { // Version == 0
		p.creationTime = uint64(r.Read4())
		p.modificationTime = uint64(r.Read4())
		p.timeScale = r.Read4()
		p.duration = uint64(r.Read4())
	}
	lang := r.Read2()
	p.language = lang & 0x7FFF
}

// parse trak/mdia/hdlr box
func (p *boxTrak) parseHdlr(r *atomReader) TrackType {
	_ = r.Move(8) // Version + flags
	_ = r.Move(4) // pre_defined 0
	handlerType := r.Read4()
	switch handlerType {
	case string2int("vide"):
		return VideoTrack
	case string2int("soun"):
		return AudioTrack
	case string2int("subt"):
		return SubtitleTrack
	default:
		/*
			there are some other type of handler type, such as,
				"meta" : Timed metadata media, metadata tracks use a NullMediaHeaderBox
				"hint" : Hint media, hint tracks use a HintMediaHeaderBox
				"text" : Timed text media, Timed text tracks use a NullMediaHeaderBox
				"fdsm" : Font media, font tracks use a NullMediaHeaderBox
			Those type of track wouldn't be parsed.
		*/
		return UnknownTrack
	}
}

// parse trak/mdia/minf box
// Notice: dinf box and media header box(vmhd/smhd/nmhd/sthd) are omitted
func (p *boxTrak) parseMinf(r *atomReader) error {
	// read extend-language
	elng, err := r.FindSubAtom(fourCCelng)
	if err != nil {
		return err
	}
	lang := make([]byte, elng.Size()-8)
	_, _ = elng.ReadBytes(lang)
	p.extLanguageTag = string(lang[4:])

	// parse "stbl" atom
	stbl, err := r.FindSubAtom(fourCCstbl)
	if err != nil {
		return err
	}
	err = p.parseStbl(stbl)
	if err != nil {
		return err
	}
	return nil
}

// parse trak/mdia/minf/stbl box
func (p *boxTrak) parseStbl(r *atomReader) (err error) {
	var sencAtomReader *atomReader = nil // parsing "senc" box depends on "sbgp" and "spgd"
	for {
		ar, err := r.GetNextAtom()
		if err != nil {
			if err == ErrEndOfAtom {
				break
			}
			if err == ErrBadAtom {
				return err
			}
		}
		switch ar.a.atomType {
		case fourCCstsd:
			err = p.parseStsd(ar)
			break
		case fourCCstts: // Decoding time to sample
			p.parseStts(ar)
			break
		case fourCCctts:
			p.parseCtts(ar)
			break
		case fourCCcslg:
			p.parseCslg(ar)
			break
		case fourCCstsc:
			p.parseStsc(ar)
			break
		case fourCCstsz:
			fallthrough
		case fourCCstz2:
			p.parseStsz(ar)
			break
		case fourCCstco:
			fallthrough
		case fourCCco64:
			p.parseStco(ar)
			break
		case fourCCstss:
			p.parseStss(ar)
			break
		case fourCCstsh:
			p.parseStsh(ar)
			break
		case fourCCpadb:
			// sample padding bits
			break
		case fourCCstdp:
			p.parseStdp(ar) // sample degradation priority
			break
		case fourCCsdtp:
			p.parseSdtp(ar)
			break
		case fourCCsbgp:
			p.sbgp = parseSbgp(ar)
			break
		case fourCCsgpd:
			p.sgpd, _ = parseSgpd(ar)
			break
		case fourCCsubs:
			p.subs = parseSubs(ar)
			break
		case fourCCsaiz:
			p.saiz = parseSaiz(ar)
			break
		case fourCCsaio:
			p.saio = parseSaio(ar)
			break
		case fourCCsenc:
			sencAtomReader = ar
		default:
			break
		}
	}
	if sencAtomReader != nil && p.sbgp != nil && p.sgpd != nil && len(p.protection) > 0 {
		p.senc, _ = parseSenc(sencAtomReader, p.sbgp, p.sgpd, p.protection[0].DefaultPerSampleIVSize)
	}
	return nil
}

// parse stsd box
func (p *boxTrak) parseStsd(r *atomReader) (err error) {
	stsdBoxSize := r.a.bodySize
	stsd := new(boxStsd)
	stsd.version, _ = r.ReadVersionFlags()
	stsd.entryCount = r.Read4()
	// check validity
	if stsd.entryCount <= 0 || stsd.entryCount >= uint32(stsdBoxSize)/8 {
		// logs.errLog.Println(" invalid stsd entry Data")
		return errors.New("invalid stsd entry Data")
	}

	for i := 0; i < int(stsd.entryCount); i++ {
		ar, err := r.GetNextAtom()
		if err != nil {
			if err == ErrEndOfAtom {
				break
			}
			if err == ErrBadAtom {
				return err
			}
		}
		trackType := getTrackType(ar.a.atomType)
		switch trackType {
		case AudioTrack:
			err = p.parseAudioSampleEntry(ar)
			break
		case VideoTrack:
			err = p.parseVideoSampleEntry(ar)
			break
		case SubtitleTrack:
			// TODO. parse subtitle sample entry
			// err = p.parseSubtitleSampleEntry(ar)
			break
		default:
			{
				err = errors.New("unsupported sample entry")
				break
			}
		}
		if err != nil {
			break
		}
	}
	return err
}

// parse stts box
func (p *boxTrak) parseStts(r *atomReader) {
	stts := new(boxStts)
	_, _ = r.ReadVersionFlags()
	stts.entryCount = r.Read4()
	for i := uint32(0); i < stts.entryCount; i++ {
		stts.sampleCount = append(stts.sampleCount, r.Read4())
		stts.sampleDelta = append(stts.sampleDelta, r.Read4())
	}
	p.stts = stts
}

// parse stsc box
func (p *boxTrak) parseStsc(r *atomReader) {
	stsc := new(boxStsc)
	_, _ = r.ReadVersionFlags()
	stsc.entryCount = r.Read4()
	for i := uint32(0); i < stsc.entryCount; i++ {
		stsc.firstChunk = append(stsc.firstChunk, r.Read4())
		stsc.samplePerChunk = append(stsc.samplePerChunk, r.Read4())
		stsc.sampleDescriptionIndex = append(stsc.sampleDescriptionIndex, r.Read4())
	}
	p.stsc = stsc
}

// parse stsz box
func (p *boxTrak) parseStsz(r *atomReader) {
	stsz := new(boxStsz)
	stsz.atomType = r.a.atomType
	_, _ = r.ReadVersionFlags()
	if stsz.atomType == fourCCstsz {
		stsz.sampleSize = r.Read4()
		stsz.sampleCount = r.Read4()
		if stsz.sampleSize == 0 {
			for i := uint32(0); i < stsz.sampleCount; i++ {
				stsz.entrySize = append(stsz.entrySize, r.Read4())
			}
		}
	}
	if stsz.atomType == fourCCstz2 {
		stsz.fieldSize = uint8(r.Read4() & 0x000000FF)
		stsz.sampleCount = r.Read4()
		for i := uint32(0); i < stsz.sampleCount; i++ {
			byteSize := func() uint {
				if stsz.fieldSize == 4 {
					return uint(stsz.sampleCount+1) / 2
				} else if stsz.fieldSize == 8 {
					return uint(stsz.sampleCount)
				} else {
					return uint(stsz.sampleCount * 2) // stsz.fieldSize == 16
				}
			}()
			buf := make([]byte, byteSize)
			_, _ = r.ReadBytes(buf)
			br := newBitReaderFromSlice(buf)
			stsz.entrySize = append(stsz.entrySize, uint32(br.ReadBitsLE16(byteSize)))
		}
	}
	p.stsz = stsz
}

// parse stco box
func (p *boxTrak) parseStco(r *atomReader) {
	stco := new(boxStco)
	_, _ = r.ReadVersionFlags()
	stco.entryCount = r.Read4()
	for i := uint32(0); i < stco.entryCount; i++ {
		if r.a.atomType == fourCCstco {
			stco.chunkOffset = append(stco.chunkOffset, uint64(r.Read4()))
		} else if r.a.atomType == fourCCco64 {
			stco.chunkOffset = append(stco.chunkOffset, r.Read8())
		}
	}
	p.stco = stco
}

// parse stss box
func (p *boxTrak) parseCtts(r *atomReader) {
	ctts := new(boxCtts)
	_, _ = r.ReadVersionFlags()
	ctts.entryCount = r.Read4()
	for i := uint32(0); i < ctts.entryCount; i++ {
		ctts.sampleCount = append(ctts.sampleCount, r.Read4())
		ctts.sampleOffset = append(ctts.sampleOffset, r.Read4S())
	}
	p.ctts = ctts
}

// parse composition to decode timeline mapping
func (p *boxTrak) parseCslg(r *atomReader) {
	cslg := new(boxCslg)
	v, _ := r.ReadVersionFlags()
	if v == 0 {
		cslg.compositionToDTSShift = int64(r.Read4S())
		cslg.leastDecodeToDisplayDelta = int64(r.Read4S())
		cslg.greatestDecodeToDisplayDelta = int64(r.Read4S())
		cslg.compositionStartTime = int64(r.Read4S())
		cslg.compositionEndTime = int64(r.Read4S())
	} else {
		cslg.compositionToDTSShift = r.Read8S()
		cslg.leastDecodeToDisplayDelta = r.Read8S()
		cslg.greatestDecodeToDisplayDelta = r.Read8S()
		cslg.compositionStartTime = r.Read8S()
		cslg.compositionEndTime = r.Read8S()
	}
	p.cslg = cslg
}

// parse stss box
func (p *boxTrak) parseStss(r *atomReader) {
	stss := new(boxStss)
	_, _ = r.ReadVersionFlags()
	stss.entryCount = r.Read4()
	for i := uint32(0); i < stss.entryCount; i++ {
		stss.sampleNumber = append(stss.sampleNumber, r.Read4())
	}
	p.stss = stss
}

func (p *boxTrak) parseStsh(r *atomReader) {
	stsh := new(boxStsh)
	_ = r.Move(4) // version + flags
	stsh.entryCount = r.Read4()
	for i := 0; i < int(stsh.entryCount); i++ {
		stsh.shadowedSampleNumber = append(stsh.shadowedSampleNumber, r.Read4())
		stsh.syncSampleNumber = append(stsh.syncSampleNumber, r.Read4())
	}
	p.stsh = stsh
}

func (p *boxTrak) parseStdp(r *atomReader) {
	_ = r.Move(4) // version + flags
	sampleCount := (r.Size() - 4) / 2
	for i := 0; i < sampleCount; i++ {
		p.samplePriority = append(p.samplePriority, r.Read2())
	}
}

func (p *boxTrak) parseSdtp(r *atomReader) {
	_ = r.Move(4) // version + flags
	sampleCount := r.Size() - 4
	sdtp := new(boxSdtp)
	for i := 0; i < sampleCount; i++ {
		i := r.ReadUnsignedByte()
		sdtp.isLeading = append(sdtp.isLeading, i>>6)
		sdtp.sampleDependsOn = append(sdtp.sampleDependsOn, (i>>4)&0x03)
		sdtp.sampleIsDependedOn = append(sdtp.sampleIsDependedOn, (i>>2)&0x03)
		sdtp.sampleHasRedundancy = append(sdtp.sampleHasRedundancy, i&0x03)
	}
	p.sampleDependency = sdtp
}

// parse pssh box
func parsePssh(p *MovieInfo, r *atomReader) error {
	pssh := new(PSSH)
	version, _ := r.ReadVersionFlags()
	pssh.SystemId = make([]byte, 16)
	_, _ = r.ReadBytes(pssh.SystemId)
	if version > 0 {
		pssh.KIdCount = r.Read4()
		for i := uint32(0); i < pssh.KIdCount; i++ {
			kid := make([]byte, 16)
			_, _ = r.ReadBytes(kid)
			pssh.KId = append(pssh.KId, kid)
		}
	}
	pssh.Data = make([]byte, r.Read4())
	_, _ = r.ReadBytes(pssh.Data)
	p.pssh = append(p.pssh, pssh)
	return nil
}

// parse saio box
func parseSaio(r *atomReader) *boxSaio {
	saio := new(boxSaio)
	version, flags := r.ReadVersionFlags()
	if flags&1 != 0 {
		saio.auxInfoType = new(uint32)
		*saio.auxInfoType = r.Read4()
		saio.auxInfoTypeParameter = new(uint32)
		*saio.auxInfoTypeParameter = r.Read4()
	}
	saio.entryCount = r.Read4()
	if version == 0 {
		for i := uint32(0); i < saio.entryCount; i++ {
			if version == 0 {
				saio.offset = append(saio.offset, uint64(r.Read4()))
			} else {
				saio.offset = append(saio.offset, r.Read8())
			}
		}
	}
	return saio
}

// parse saiz box
func parseSaiz(r *atomReader) *boxSaiz {
	saiz := new(boxSaiz)
	_, flags := r.ReadVersionFlags()
	if flags&1 != 0 {
		saiz.auxInfoType = new(uint32)
		*saiz.auxInfoType = r.Read4()
		saiz.auxInfoTypeParameter = new(uint32)
		*saiz.auxInfoTypeParameter = r.Read4()
	}
	saiz.defaultSampleInfoSize = r.ReadUnsignedByte()
	saiz.sampleCount = r.Read4()
	for i := uint32(0); i < saiz.sampleCount; i++ {
		saiz.sampleInfoSize = append(saiz.sampleInfoSize, r.ReadUnsignedByte())
	}
	return saiz
}

// parse sbgp box
func parseSbgp(r *atomReader) *boxSbgp {
	sbgp := new(boxSbgp)
	version, _ := r.ReadVersionFlags()
	sbgp.groupingType = r.Read4()
	if version == 1 {
		sbgp.groupingTypeParameter = new(uint32)
		*sbgp.groupingTypeParameter = r.Read4()
	}
	sbgp.entryCount = r.Read4()
	if r.Len() < int(8*sbgp.entryCount) {
		return nil // check
	}
	for i := uint32(0); i < sbgp.entryCount; i++ {
		sbgp.sampleCount = append(sbgp.sampleCount, r.Read4())
		sbgp.groupDescriptionIndex = append(sbgp.groupDescriptionIndex, r.Read4())
	}
	return sbgp
}

// parse sgpd box
func parseSgpd(r *atomReader) (*boxSgpd, error) {
	sgpd := new(boxSgpd)
	version, _ := r.ReadVersionFlags()
	sgpd.groupingType = r.Read4()
	if int2String(sgpd.groupingType) != "seig" {
		return nil, ErrUnsupportedSampleGroupType
	}
	if version == 1 {
		sgpd.defaultLength = new(uint32)
		*sgpd.defaultLength = r.Read4()
		if *sgpd.defaultLength == 0 || *sgpd.defaultLength > 20 {
			return nil, ErrUnsupportedVariableSampleGroupLength
		}
	} else if version >= 2 {
		sgpd.defaultSampleDescriptionIndex = new(uint32)
		*sgpd.defaultSampleDescriptionIndex = r.Read4()
	}
	sgpd.entryCount = r.Read4()
	for i := uint32(0); i < sgpd.entryCount; i++ {
		if version == 1 {
			if *sgpd.defaultLength == 0 {
				*sgpd.descriptionLength = r.Read4()
				if *sgpd.descriptionLength > 16 {
					return nil, ErrInvalidLengthOfSampleGroup
				}
			}
		}
		// CencSampleEncryptionInformationGroupEntry
		// only support "cenc" scheme currently
		cencGroupEntry := new(cencSampleEncryptionInformationGroupEntry)
		_ = r.Move(1) // reversed 1 byte
		byteT := r.ReadUnsignedByte()
		cencGroupEntry.cryptByteBlock = byteT >> 4
		cencGroupEntry.skipByteBlock = byteT & 0x0F
		cencGroupEntry.isProtected = r.ReadUnsignedByte() != 0
		cencGroupEntry.perSampleIVSize = r.ReadUnsignedByte()
		if cencGroupEntry.perSampleIVSize != 8 && cencGroupEntry.perSampleIVSize != 16 {
			return nil, ErrInvalidLengthOfIVInSampleGroup
		}
		cencGroupEntry.kID = make([]byte, 16)
		_, _ = r.ReadBytes(cencGroupEntry.kID)
		if cencGroupEntry.isProtected && cencGroupEntry.perSampleIVSize == 0 {
			constIVSize := r.ReadUnsignedByte()
			if constIVSize != 8 && constIVSize != 16 {
				return nil, ErrInvalidLengthOfIVInSampleGroup
			}
			cencGroupEntry.constantIV = make([]byte, constIVSize)
			_, _ = r.ReadBytes(cencGroupEntry.constantIV)
		}
		sgpd.cencGroupEntries = append(sgpd.cencGroupEntries, cencGroupEntry)
	}
	return sgpd, nil
}

// parse senc box
func parseSenc(r *atomReader, sbgp *boxSbgp, sgpd *boxSgpd, defaultPerSampleIVSize uint8) (*boxSenc, error) {
	senc := new(boxSenc)
	_, senc.flags = r.ReadVersionFlags()
	senc.sampleCount = r.Read4()
	iVSize := defaultPerSampleIVSize
	sampleGroupDescriptionIndexList := func() []uint32 {
		var sampleGroupDescriptionIndexList []uint32
		if sbgp != nil {
			for j := uint32(0); j < sbgp.entryCount; j++ {
				for k := uint32(0); k < sbgp.sampleCount[j]; k++ {
					sampleGroupDescriptionIndexList = append(sampleGroupDescriptionIndexList, sbgp.groupDescriptionIndex[j])
				}
			}
		}
		return sampleGroupDescriptionIndexList
	}()

	// tryToDetectIVSize try to detect the IV's size in the absence of movie header or sample-group information.
	tryToDetectIVSize := func(r *atomReader, sampleCount uint32, flags uint32) (uint8, error) {
		left := uint32(r.Len())
		resumePosition, _ := r.r.Seek(0, io.SeekCurrent)
		if left <= 8 || sampleCount == 0 {
			return 0, errors.New("too small for left size or sampleCount in senc box is 0")
		}
		sencSampleUnitSize := left / sampleCount
		if flags&0x000002 == 0 {
			if sencSampleUnitSize != 8 && sencSampleUnitSize != 16 {
				return 0, errors.New("failed to detect the IV size")
			}
			return uint8(sencSampleUnitSize), nil
		}
		// has subSample, IV's size is 8 or 16
		increaseStep := 8
		ivSize := 0
	DETECTOR:
		_, _ = r.r.Seek(resumePosition, io.SeekStart)
		ivSize += increaseStep
		if ivSize > 16 {
			return 0, errors.New("failed to detect the IV size")
		}
		_ = r.Move(ivSize)
		subSampleCount := r.Read2()
		if uint32(subSampleCount)*(4+2) != (left - uint32(ivSize) - 4) {
			goto DETECTOR
		} else {
			_, _ = r.r.Seek(resumePosition, io.SeekStart)
			return uint8(ivSize), nil
		}
	}

	for i := uint32(0); i < senc.sampleCount; i++ {
		if sgpd != nil && sbgp != nil && i < uint32(len(sampleGroupDescriptionIndexList)) {
			index := sampleGroupDescriptionIndexList[i]
			if index != 0 {
				if index > 65536 {
					index -= 65536
				}
				if index < uint32(len(sgpd.cencGroupEntries)) {
					cencSampleEncryptionInformationEntry := sgpd.cencGroupEntries[index-1]
					if cencSampleEncryptionInformationEntry != nil {
						iVSize = cencSampleEncryptionInformationEntry.perSampleIVSize
					}
				}
			}
		}
		if iVSize == 0 {
			// try to detect the iv size
			n, err := tryToDetectIVSize(r, senc.sampleCount, senc.flags)
			if err != nil {
				return nil, errors.New("failed to parseConfig senc box, because the IV size is invalid")
			}
			iVSize = n
		}
		sampleEnc := new(sampleEncryption)
		sampleEnc.IV = make([]byte, iVSize)
		_, _ = r.ReadBytes(sampleEnc.IV)
		if senc.flags&0x000002 != 0 {
			sampleEnc.subSampleCount = r.Read2()
			for j := uint16(0); j < sampleEnc.subSampleCount; j++ {
				clearData := r.Read2()
				protectedData := r.Read4()
				sampleEnc.subSamples = append(sampleEnc.subSamples, subSampleEncryption{bytesOfClearData: clearData, bytesOfProtectedData: protectedData})
			}
		}
		senc.samples = append(senc.samples, sampleEnc)
	}
	return senc, nil
}

// parse Subs Box
func parseSubs(r *atomReader) *boxSubs {
	subs := new(boxSubs)
	version, flags := r.ReadVersionFlags()
	subs.flags = flags
	subs.entryCount = r.Read4()
	for i := uint32(0); i < subs.entryCount; i++ {
		sampleEntry := new(subSampleEntry)
		sampleEntry.sampleDelta = r.Read4()
		sampleEntry.subSampleCount = r.Read2()
		if sampleEntry.subSampleCount > 0 {
			for j := uint16(0); j < sampleEntry.subSampleCount; j++ {
				subSampleInfo := new(subSampleInfo)
				if version == 1 {
					subSampleInfo.subSampleSize = r.Read4()
				} else {
					subSampleInfo.subSampleSize = uint32(r.Read2())
				}
				subSampleInfo.subSamplePriority = r.ReadUnsignedByte()
				subSampleInfo.discardable = r.ReadUnsignedByte()
				subSampleInfo.codecSpecificParameters = r.Read4()
				sampleEntry.subSamples = append(sampleEntry.subSamples, subSampleInfo)
			}
			subs.entries = append(subs.entries, sampleEntry)
		}
	}
	return subs
}

// parse traf box
func (p *movieFragment) parseTraf(r *atomReader) (err error) {
	fragment := new(trackFragment)
	fragment.movie = p.movie
	fragment.moof = p
	var sencAtomReader *atomReader = nil
	for {
		ar, e := r.GetNextAtom()
		if e != nil {
			if e == ErrEndOfAtom {
				break
			}
			if e == ErrBadAtom {
				return e
			}
		}
		switch ar.a.atomType {
		case fourCCtfhd:
			fragment.parseTfhd(ar)
			break
		case fourCCtfdt:
			fragment.parseTfdt(ar)
			break
		case fourCCtrun:
			fragment.parseTrun(ar)
			break
		case fourCCsaio:
			fragment.saio = parseSaio(ar)
			break
		case fourCCsaiz:
			fragment.saiz = parseSaiz(ar)
			break
		case fourCCsbgp:
			fragment.sbgp = parseSbgp(ar)
			break
		case fourCCsgpd:
			fragment.sgpd, _ = parseSgpd(ar)
			break
		case fourCCsubs:
			fragment.subs = append(fragment.subs, parseSubs(ar))
			break
		case fourCCsenc:
			sencAtomReader = ar
			break
		}
	}
	if fragment.sgpd != nil && fragment.sbgp != nil && fragment.trackInfo() != nil && len(fragment.trackInfo().protection) != 0 {
		fragment.senc, err = parseSenc(sencAtomReader, fragment.sbgp, fragment.sgpd, fragment.trackInfo().protection[0].DefaultPerSampleIVSize)
	}
	p.fragment = append(p.fragment, fragment)
	return err
}

// parse trun box
func (p *trackFragment) parseTrun(r *atomReader) {
	// without checking atom's size
	trun := new(boxTrun)
	version, flags := r.ReadVersionFlags()
	trun.sampleCount = r.Read4()
	if flags&0x000001 != 0 {
		trun.dataOffset = new(uint32)
		*trun.dataOffset = r.Read4()
	}
	if flags&0x000004 != 0 {
		trun.firstSampleFlags = new(uint32)
		*trun.firstSampleFlags = r.Read4()
	}
	for i := uint32(0); i < trun.sampleCount; i++ {
		sampleTrun := new(trunSample)
		if flags&0x000100 != 0 {
			sampleTrun.sampleDuration = new(uint32)
			*sampleTrun.sampleDuration = r.Read4()
		}
		if flags&0x000200 != 0 {
			sampleTrun.sampleSize = new(uint32)
			*sampleTrun.sampleSize = r.Read4()
		}
		if flags&0x000400 != 0 {
			sampleTrun.sampleFlags = new(uint32)
			*sampleTrun.sampleFlags = r.Read4()
		}
		if flags&0x000800 != 0 {
			sampleTrun.sampleCompositionTimeOffset = new(int32)

			if version == 0 {
				*sampleTrun.sampleCompositionTimeOffset = int32(r.Read4())
			} else {
				*sampleTrun.sampleCompositionTimeOffset = r.Read4S()
			}
		}
		trun.samples = append(trun.samples, sampleTrun)
	}
	p.trun = append(p.trun, trun)
}

// parse tfdt box
func (p *trackFragment) parseTfdt(r *atomReader) {
	version, _ := r.ReadVersionFlags()
	p.baseMediaDecodeTime = new(uint64)
	if version != 0 {
		*p.baseMediaDecodeTime = r.Read8()
	} else {
		*p.baseMediaDecodeTime = uint64(r.Read4())
	}
}

func (p *trackFragment) parseTfhd(r *atomReader) {
	p.flags = r.Read4() & 0x00FFFFFF
	p.trackID = r.Read4()
	if p.flags&0x000001 != 0 {
		p.baseDataOffset = new(uint64)
		*p.baseDataOffset = r.Read8()
	} else if p.flags&0x000002 != 0 {
		p.sampleDescriptionIndex = new(uint32)
		*p.sampleDescriptionIndex = r.Read4()
	} else if p.flags&0x000008 != 0 {
		p.defaultSampleDuration = new(uint32)
		*p.defaultSampleDuration = r.Read4()
	} else if p.flags&0x000010 != 0 {
		p.defaultSampleSize = new(uint32)
		*p.defaultSampleSize = r.Read4()
	} else if p.flags&0x000020 != 0 {
		p.defaultSampleFlags = new(uint32)
		*p.defaultSampleFlags = r.Read4()
	} else if p.flags&0x000001 == 0 && p.flags&0x020000 != 0 {
		p.defaultBaseIsMoof = true
	}
}
