package main

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
		integerPart := r.Read2S()
		fractionPart := r.Read2()
		edts.mediaRate = append(edts.mediaRate, float32(integerPart)+float32(fractionPart)/100)
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

// parse moov/trak/mdia box
func (p *boxTrak) parseMdia(reader *atomReader) error {
	for {
		itemReader, err := reader.GetSubAtom()
		if err != nil {
			if err == ErrNoMoreAtom {
				break
			} else {
				return err
			}
		}
		switch itemReader.TypeCC() {
		case fourCChdlr:
			p.trackType = p.parseHdlr(itemReader)
			break
		case fourCCmdhd:

			p.parseMdhd(itemReader)
			break
		case fourCCelng:
			p.parseElng(itemReader)
			break
		case fourCCminf:
			err = p.parseMinf(itemReader)
			if err != nil {
				return nil
			}
			break
		default:
			break
		}
	}
	return nil
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

func (p *boxTrak) parseElng(r *atomReader) {
	r.Move(4)
	extLanguage := make([]byte, r.a.bodySize-4)
	r.ReadBytes(extLanguage)
	p.extLanguage = string(extLanguage)
}

// parse trak/mdia/minf box
// Notice: dinf box and media header box(vmhd/smhd/nmhd/sthd) are omitted
func (p *boxTrak) parseMinf(reader *atomReader) error {
	for {
		itemReader, err := reader.GetSubAtom()
		if err != nil {
			if err == ErrNoMoreAtom {
				break
			} else {
				return err
			}
		}
		if itemReader.TypeCC() != fourCCstbl {
			continue
		}
		err = p.parseStbl(itemReader)
		if err != nil {
			return err
		}
	}
	return nil
}

// parse trak/mdia/minf/stbl box
func (p *boxTrak) parseStbl(reader *atomReader) (err error) {
	var sencAtomReader *atomReader = nil // parsing "senc" box depends on "sbgp" and "spgd"
	for {
		itemReader, err := reader.GetSubAtom()
		if err != nil {
			if err == ErrNoMoreAtom {
				break
			} else {
				return err
			}
		}
		switch itemReader.TypeCC() {
		case fourCCstsd:
			_ = p.parseStsd(itemReader)

		case fourCCstts: // Decoding time to sample
			p.parseStts(itemReader)

		case fourCCctts:
			p.parseCtts(itemReader)

		case fourCCcslg:
			p.parseCslg(itemReader)

		case fourCCstsc:
			p.parseStsc(itemReader)

		case fourCCstsz:
			fallthrough
		case fourCCstz2:
			p.parseStsz(itemReader)

		case fourCCstco:
			fallthrough
		case fourCCco64:
			p.parseStco(itemReader)

		case fourCCstss:
			p.parseStss(itemReader)

		case fourCCstsh:
			p.parseStsh(itemReader)

		case fourCCpadb:
			// sample padding bits

		case fourCCstdp:
			p.parseStdp(itemReader) // sample degradation priority

		case fourCCsdtp:
			p.parseSdtp(itemReader)

		case fourCCsbgp:
			p.sbgp = parseSbgp(itemReader)

		case fourCCsgpd:
			p.sgpd, _ = parseSgpd(itemReader)

		case fourCCsubs:
			p.subs = parseSubs(itemReader)

		case fourCCsaiz:
			if p.encrypted {
				p.saiz = parseSaiz(itemReader)
			}

		case fourCCsaio:
			if p.encrypted {
				p.saio = parseSaio(itemReader, p)
			}

		case fourCCsenc:
			sencAtomReader = itemReader

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

	for {
		itemReader, err := r.GetSubAtom()
		if err != nil {
			if err == ErrNoMoreAtom {
				break
			}
			if err == ErrNoEnoughData {
				return err
			}
		}
		trackType := getTrackType(itemReader.TypeCC())
		switch trackType {
		case AudioTrack:
			err = p.parseAudioSampleEntry(itemReader)
			break
		case VideoTrack:
			err = p.parseVideoSampleEntry(itemReader)
			break
		case SubtitleTrack:
			// TODO. parse subtitle sample entry
			// err = p.parseSubtitleSampleEntry(ar)
			break
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
	sampleNumber := uint64(0)
	duration := uint64(0)
	for i := uint32(0); i < stts.entryCount; i++ {
		stts.sampleCount = append(stts.sampleCount, r.Read4())
		stts.sampleDelta = append(stts.sampleDelta, r.Read4())
		duration += uint64(stts.sampleCount[i]) * uint64(stts.sampleDelta[i])
		sampleNumber += uint64(stts.sampleCount[i])
	}
	p.duration = min(p.movie.duration, duration)
	p.sampleNumber = sampleNumber
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
		logD.Printf("stsc first Chunk=%d  samples=%d index=%d\n", stsc.firstChunk, stsc.samplePerChunk, stsc.sampleDescriptionIndex)
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
	_, _ = r.ReadVersionFlags()
	entries := r.Read4()
	if entries <= 0 {
		return
	}
	p.syncSamples = make([]uint32, entries)
	for i := uint32(0); i < entries; i++ {
		p.syncSamples = append(p.syncSamples, r.Read4())
	}
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
		kIdCount := r.Read4()
		for i := uint32(0); i < kIdCount; i++ {
			var kId [16]byte
			_, _ = r.ReadBytes(kId[:])
			pssh.KId = append(pssh.KId, kId)
		}
	}
	pssh.Data = make([]byte, r.Read4())
	_, _ = r.ReadBytes(pssh.Data)
	p.pssh = append(p.pssh, pssh)
	return nil
}

// parse saio box
// ISO/IEC 14496-12:2020(E) 8.7.9.1
func parseSaio(r *atomReader, container interface{}) *boxSaio {
	saio := new(boxSaio)
	version, flags := r.ReadVersionFlags()
	if flags&1 != 0 {
		saio.auxInfoType = r.Read4()
		saio.auxInfoTypeParameter = r.Read4()
	} else {
		return nil
	}

	saio.entryCount = r.Read4()
	if saio.entryCount != 1 {
		logE.Println("saio entry count must be 1")
		return nil
	}
	if version == 0 {
		for i := uint32(0); i < saio.entryCount; i++ {
			if version == 0 {
				saio.offset = append(saio.offset, uint64(r.Read4()))
			} else {
				saio.offset = append(saio.offset, r.Read8())
			}
		}
	}
	// parse the CencSampleAuxiliaryDataFormat which is
	// defined in ISO/IEC 23001-7:2016(E) 7.1

	// In fact, the content pointed by the offset is contained by 'senc' box
	// So, we don't need to parse the offset here.

	return saio
}

// parse saiz box
func parseSaiz(r *atomReader) *boxSaiz {
	saiz := new(boxSaiz)
	_, flags := r.ReadVersionFlags()
	if flags&1 != 0 {
		saiz.auxInfoType = r.Read4()
		saiz.auxInfoTypeParameter = r.Read4()
	} else {
		return nil
	}
	saiz.defaultSampleInfoSize = r.ReadUnsignedByte()
	saiz.sampleCount = r.Read4()
	if saiz.defaultSampleInfoSize == 0 {
		for i := uint32(0); i < saiz.sampleCount; i++ {
			saiz.sampleInfoSize = append(saiz.sampleInfoSize, r.ReadUnsignedByte())
		}
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
	// Not sure it works properly. If perSampleIVSize is 0, in specs, the constantIV in tenc box
	// should be used as IV.
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
				subSample := new(subSampleInfo)
				if version == 1 {
					subSample.subSampleSize = r.Read4()
				} else {
					subSample.subSampleSize = uint32(r.Read2())
				}
				subSample.subSamplePriority = r.ReadUnsignedByte()
				subSample.discardable = r.ReadUnsignedByte()
				subSample.codecSpecificParameters = r.Read4()
				sampleEntry.subSamples = append(sampleEntry.subSamples, subSample)
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
		ar, e := r.GetSubAtom()
		if e != nil {
			if errors.Is(ErrNoMoreAtom, e) {
				break
			}
			if errors.Is(ErrNoEnoughData, e) {
				return e
			}
		}
		switch ar.a.atomType {
		case fourCCtfhd:
			fragment.parseTfhd(ar)

		case fourCCtfdt:
			fragment.parseTfdt(ar)

		case fourCCtrun:
			fragment.parseTrun(ar)

		case fourCCsaio:
			fragment.saio = parseSaio(ar, p)

		case fourCCsaiz:
			fragment.saiz = parseSaiz(ar)

		case fourCCsbgp:
			fragment.sbgp = parseSbgp(ar)

		case fourCCsgpd:
			fragment.sgpd, _ = parseSgpd(ar)

		case fourCCsubs:
			fragment.subs = append(fragment.subs, parseSubs(ar))

		case fourCCsenc:
			sencAtomReader = ar

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
	// data-offset-present
	if flags&0x000001 != 0 {
		trun.dataOffset = new(uint32)
		*trun.dataOffset = r.Read4()
	}
	// first-sample-flags-present
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

// parseTfhd will parse track fragment decode time
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
