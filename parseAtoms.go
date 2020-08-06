package fmp4parser

import (
	"errors"
	"fmt"
)

// parse ftyp box
func parseFtyp(p *MovieInfo, r *deMuxReader, a *atom) error {
	err := r.CheckEnoughAtomData(a.atomSize)
	if err != nil {
		return err
	}
	logD.Print("parse ftyp box")
	p.ftyp = new(boxFtyp)
	p.ftyp.majorBrand = r.Read4()
	// check QuickTime format
	if p.ftyp.majorBrand == 0x71742020 {
		p.ftyp.isQuickTimeFormat = true
	}
	p.ftyp.minorVersion = r.Read4()
	for brandCount := a.atomSize / 4; brandCount > 0; brandCount-- {
		compatibleBrand := r.Read4()
		if compatibleBrand == 0x71742020 {
			p.ftyp.isQuickTimeFormat = true
		}
		p.ftyp.compatibleBrands = append(p.ftyp.compatibleBrands, compatibleBrand)
	}

	return nil
}

// parse ssix box (SubSegment Index box)
func parseSsix(p *MovieInfo, r *deMuxReader, a *atom) error {
	err := r.CheckEnoughAtomData(a.atomSize)
	if err != nil {
		return err
	}
	r.Move(8) // 0,0
	ssix := new(boxSsix)
	ssix.sugSegmentCount = r.Read4()
	for i := 0; i < int(ssix.sugSegmentCount); i++ {
		var tmpRange struct {
			rangeCount uint32 // is rangeSize's len
			rangeSize  []struct {
				level uint8
				size  uint32
			}
		}
		tmpRange.rangeCount = r.Read4()
		for j := 0; j < int(tmpRange.rangeCount); j++ {
			tmp := r.Read4()
			var tmpRangeSize struct {
				level uint8
				size  uint32
			}
			tmpRangeSize.level = uint8(tmp >> 24 & 0xFF)
			tmpRangeSize.size = tmp & 0xFFFFFF
			tmpRange.rangeSize = append(tmpRange.rangeSize, tmpRangeSize)
		}
		ssix.ranges = append(ssix.ranges, tmpRange)
	}
	p.ssix = append(p.ssix, ssix)
	return nil
}

// parse sdix box (Segment Index box)
func parseSidx(p *MovieInfo, r *deMuxReader, a *atom) error {
	err := r.CheckEnoughAtomData(a.atomSize)
	if err != nil {
		return err
	}
	version, _ := r.ReadVersionFlags()
	sidx := new(boxSidx)
	sidx.referenceID = r.Read4()
	sidx.timeScale = r.Read4()
	if version == 0 {
		sidx.earlistPresentationTime = uint64(r.Read4())
		sidx.firstTime = uint64(r.Read4())
	} else {
		sidx.earlistPresentationTime = r.Read8()
		sidx.firstTime = r.Read8()
	}
	r.Move(2) // reserved
	sidx.referenceCount = r.Read2()
	for i := uint16(0); i < sidx.referenceCount; i++ {
		var reference struct {
			referenceType      uint8  // reference_type 1 bit
			referenceSize      uint32 // reference_size 31 bit
			subSegmentDuration uint32
			startWithSAP       uint8  // starts_with_SAP 1 bit
			sapType            uint8  // SAP_type 3 bit
			sapDeltaTime       uint32 // SAP_delta_time 28 bit
		}
		typeSize := r.Read4()
		reference.referenceType = uint8(typeSize >> 31 & 0x1)
		reference.referenceSize = typeSize & 0x7FFFFFFF
		reference.subSegmentDuration = r.Read4()
		sap := r.Read4()
		reference.startWithSAP = uint8(sap >> 31 & 0x1)
		reference.sapType = uint8(sap >> 24 & 0x7)
		reference.sapDeltaTime = sap & 0xFFFFFFF
		sidx.reference = append(sidx.reference, reference)
	}
	p.sidx = append(p.sidx, sidx)
	// logD.Print(sidx)
	return nil
}

// parse moov box
func parseMoov(p *MovieInfo, r *deMuxReader, a *atom) error {
	logD.Print("parse moov box")
	stopPosition := r.Position() + a.atomSize
	var err error = nil
	for nextAtom := r.Position(); r.Position() < stopPosition; _ = r.MoveTo(nextAtom) {
		err = r.GetNextAtomData()
		if err != nil {
			return err
		}
		a, _ := r.GetAtomHeader()

		nextAtom += a.Size()
		if _, ok := moovParseTable[a.atomType]; ok {
			err = moovParseTable[a.atomType](p, r, a)
			if err != nil {
				logE.Printf("error encountered when parsing atom %s, error type: %T", a.Type(), err)
				return err
			}
		} else {
			logW.Printf("atom:%s will not be parsed", a.Type())
		}
	}
	logD.Print("DDDDD track size is ", len(p.trak))
	return nil
}

// parse mvhd box
func parseMvhd(p *MovieInfo, r *deMuxReader, a *atom) error {
	if r.CheckEnoughAtomData(a.atomSize) != nil {
		return ErrNoEnoughData
	}
	p.mvhd = new(boxMvhd)
	version, _ := r.ReadVersionFlags()
	if version == 1 {
		p.mvhd.creationTime = r.Read8()
		p.mvhd.modificationTime = r.Read8()
		p.mvhd.timeScale = r.Read4()
		p.mvhd.duration = r.Read8()
	} else {
		p.mvhd.creationTime = uint64(r.Read4())
		p.mvhd.modificationTime = uint64(r.Read4())
		p.mvhd.timeScale = r.Read4()
		p.mvhd.duration = uint64(r.Read4())
	}
	r.Move(70) // 10 bytes reserved. 36 bytes matrix. 24 bytes pre_defined
	p.mvhd.nextTrackId = r.Read4()
	return nil
}

// parse mvex box
func parseMvex(p *MovieInfo, r *deMuxReader, a *atom) error {
	if r.CheckEnoughAtomData(a.atomSize) != nil {
		return ErrNoEnoughData
	}
	logD.Print("parsing moov.mvex, ", a)
	p.mvex = new(boxMvex)
	p.hasFragment = true
	stopPosition := r.Position() + a.atomSize
	var err error = nil
	var nextAtom int64
	for ; r.Position() < stopPosition; _ = r.MoveTo(nextAtom) {
		a, _ := r.GetAtomHeader()
		nextAtom = r.Position() + a.atomSize
		switch a.atomType {
		case fourCCmehd:
			err = parseMehd(p, r, a)
			if err != nil {
				goto T
			}
			break
		case fourCCtrex:
			err = parseTrex(p, r, a)
			if err != nil {
				goto T
			}
			break
		case fourCCleva:
			err = parseLeva(p, r, a)
			if err != nil {
				goto T
			}
			break
		default:
			err = fmt.Errorf("parseTrak: unsupported/unparsed atom type in trak, atom type is %s", a.Type())
			goto T
		}
	}
T:
	return err
}

// parse mvhd box
func parseMehd(p *MovieInfo, r *deMuxReader, a *atom) error {
	logD.Print("parsing moov.mvex.mehd, ", a)
	mehd := new(boxMehd)
	version, _ := r.ReadVersionFlags()
	if version == 0 {
		mehd.fragmentDuration = uint64(r.Read4())
	} else {
		mehd.fragmentDuration = r.Read8()
	}
	p.mvex.mehd = *mehd
	return nil
}

// parse trex box
func parseTrex(p *MovieInfo, r *deMuxReader, a *atom) error {
	logD.Print("parsing moov.mvex.trex, ", a)
	trex := new(boxTrex)
	r.Move(8) // Version + flags
	trex.trackId = r.Read4()
	trex.defaultSampleDescriptionIndex = r.Read4()
	trex.defaultSampleDuration = r.Read4()
	trex.defaultSampleSize = r.Read4()
	trex.defaultSampleFlags = r.Read4()
	p.mvex.trex = append(p.mvex.trex, *trex)
	return nil
}

// parse leva box
func parseLeva(p *MovieInfo, r *deMuxReader, a *atom) error {
	logD.Print("parsing moov.mvex.leva, ", a)
	leva := new(boxLeva)
	leva.levelCount = r.ReadUnsignedByte()
	for i := 0; i < int(leva.levelCount); i++ {
		var level struct {
			trackId               uint32
			paddingFlag           uint8  // 1 bit
			assignmentType        uint8  // 7bit
			groupingType          uint32 // assignmentType == 0 || 1
			groupingTypeParameter uint32 // assignmentType == 1
			subTrackId            uint32 // assignmentType == 4
		}
		level.trackId = r.Read4()
		tmp := r.ReadUnsignedByte()
		level.paddingFlag = tmp >> 7 & 0x1
		level.assignmentType = tmp & 0x7F
		if level.assignmentType == 0 {
			level.groupingType = r.Read4()
		} else if level.assignmentType == 1 {
			level.groupingType = r.Read4()
			level.groupingTypeParameter = r.Read4()
		} else if level.assignmentType == 4 {
			level.subTrackId = r.Read4()
		}
		leva.levels = append(leva.levels, level)
	}
	p.mvex.leva = leva
	return nil
}

// parse trak box
func parseTrak(p *MovieInfo, r *deMuxReader, a *atom) error {
	if r.CheckEnoughAtomData(a.atomSize) != nil {
		logE.Print("when parsing moov.trak, no enough data")
		return ErrNoEnoughData
	}
	trak := new(boxTrak)
	p.trak = append(p.trak, trak)
	stopPosition := r.Position() + a.atomSize
	var err error = nil
	for nextAtom := r.Position(); r.Position() < stopPosition; _ = r.MoveTo(nextAtom) {
		a, _ := r.GetAtomHeader()
		nextAtom += a.Size()
		logD.Print("parse moov.trak: current box is ", a)
		switch a.atomType {
		case fourCCtkhd:
			err = parseTkhd(p, r, a)
			if err != nil {
				goto T
			}
			break
		case fourCCedts:
			err = parseEdts(p, r, a)
			if err != nil {
				goto T
			}
			break
		case fourCCmdia:
			err = parseMdia(p, r, a)
			if err != nil {
				logD.Print("error encountered when parsing moov.trak.mdia ,reason is ", err)
				goto T
			}
			break
		default:
			// err = fmt.Errorf("parseTrak: unsupported/unparsed atom type in trak, atom type is %s",a.Type())
			break
		}
	}
T:
	if err != nil {
		p.trak = p.trak[:len(p.trak)-1]
	}
	return err
}

// parse tkhd box
func parseTkhd(p *MovieInfo, r *deMuxReader, _ *atom) error {
	version, flags := r.ReadVersionFlags()
	tkhd := new(boxTkhd)
	if version == 1 {
		tkhd.creationTime = r.Read8()
		tkhd.modificationTime = r.Read8()
		tkhd.trackId = r.Read4()
		r.Move(4) // reversed 0
		tkhd.duration = r.Read8()
	} else {
		tkhd.creationTime = uint64(r.Read4())
		tkhd.modificationTime = uint64(r.Read4())
		tkhd.trackId = r.Read4()
		r.Move(4) // reversed 0
		tkhd.duration = uint64(r.Read4())
	}
	r.Move(8) // reversed 8 bytes
	r.Move(4) // layer, alternate_group 0
	tkhd.volume = r.Read2()
	r.Move(2)
	r.Move(36) // matrix= { 0x00010000,0,0,0,0x00010000,0,0,0,0x40000000 };
	tkhd.width = r.Read4()
	tkhd.height = r.Read4()

	flag := flags & 0xFF
	tkhd.flagTrackEnabled = flag&0x01 != 0x00
	tkhd.flagTrackInMovie = flag&0x02 != 0x00
	tkhd.flagTrackInPreview = flag&0x04 != 0x00
	p.trak[len(p.trak)-1].tkhd = tkhd
	return nil
}

// parse edts box
func parseEdts(p *MovieInfo, r *deMuxReader, _ *atom) error {
	edts := new(boxEdts)
	r.Move(4) // fourCCelts
	version, _ := r.ReadVersionFlags()
	edts.entryCount = r.Read4()
	for i := uint32(0); i < edts.entryCount; i++ {
		if version == 1 {
			edts.entrySegmentDuration = append(edts.entrySegmentDuration, r.Read8())
			edts.entryMediaTime = append(edts.entryMediaTime, r.Read8())
		} else {
			edts.entrySegmentDuration = append(edts.entrySegmentDuration, uint64(r.Read4()))
			edts.entryMediaTime = append(edts.entryMediaTime, uint64(r.Read4()))
		}
		edts.mediaRateInteger = append(edts.mediaRateInteger, r.Read2())
		r.Move(2) // media_rate_fraction == 0
	}
	p.trak[len(p.trak)-1].edts = edts
	return nil
}

// parse mdia box
func parseMdia(p *MovieInfo, r *deMuxReader, a *atom) error {
	logD.Print("parsing moov.trak.mdia ", a)
	mdia := new(boxMdia)
	p.trak[len(p.trak)-1].mdia = mdia
	stopPosition := r.Position() + a.atomSize
	var err error = nil
	for nextAtom := r.Position(); r.Position() < stopPosition; _ = r.MoveTo(nextAtom) {
		a, _ := r.GetAtomHeader()
		nextAtom += a.Size()
		switch a.atomType {
		case fourCCmdhd:
			err = parseMdhd(p, r, a)
			if err != nil {
				goto T
			}
			break
		case fourCChdlr:
			err = parseHdlr(p, r, a)
			if err != nil {
				goto T
			}
			break
		case fourCCminf:
			err = parseMinf(p, r, a)
			if err != nil {
				goto T
			}
			break
		case fourCCelng:
			r.Move(4)
			lang, _, _ := r.ReadBytes(int(a.atomSize - 4))
			mdia.extLanguageTag = string(lang)
			break
		default:
			err = fmt.Errorf("parsemdia: unsupported/unparsed atom type in mdia, atom type is %s", a.Type())
			goto T
		}
	}
T:
	return err
}

// parse mdhd box
func parseMdhd(p *MovieInfo, r *deMuxReader, _ *atom) error {
	mida := p.trak[len(p.trak)-1].mdia
	mdhd := new(boxMdhd)
	mida.mdhd = mdhd
	version, _ := r.ReadVersionFlags()
	if version == 1 {
		mdhd.creationTime = r.Read8()
		mdhd.modificationTime = r.Read8()
		mdhd.timeScale = r.Read4()
		mdhd.duration = r.Read8()
	} else { // Version == 0
		mdhd.creationTime = uint64(r.Read4())
		mdhd.modificationTime = uint64(r.Read4())
		mdhd.timeScale = r.Read4()
		mdhd.duration = uint64(r.Read4())
	}
	lang := r.Read2()
	mdhd.language = lang & 0x7FFF
	return nil
}

// parse hdlr box
func parseHdlr(p *MovieInfo, r *deMuxReader, a *atom) error {
	hdlr := new(boxHdlr)
	r.Move(4) // Version , flags 00 00 00 00
	r.Move(4) // pre_defined 0
	hdlr.handlerType = r.Read4()
	r.Move(12) // reversed. unsigned int[32] *3
	name, _, _ := r.ReadBytes(int(a.atomSize - 24))
	hdlr.name = string(name)
	if p.topLevelType == fourCCmoov {
		(p.trak[len(p.trak)-1].mdia).hldr = hdlr
	} else if a.atomType == fourCCmeta {
		p.hdlr = hdlr
	}
	return nil
}

// parse minf box
func parseMinf(p *MovieInfo, r *deMuxReader, a *atom) error {
	minf := new(boxMinf)
	p.trak[len(p.trak)-1].mdia.minf = minf
	stopPosition := r.Position() + a.atomSize
	var err error = nil
	for nextAtom := r.Position(); r.Position() < stopPosition; _ = r.MoveTo(nextAtom) {
		a, _ := r.GetAtomHeader()
		nextAtom += a.Size()
		logD.Print("parsing moov.trak.mdia.minf, current box is ", a)
		switch a.atomType {
		case fourCCvmhd:
			fallthrough
		case fourCCsmhd:
			fallthrough
		case fourCChmhd:
			err = parseXmhd(p, r, a)
			if err != nil {
				goto T
			}
			break
		case fourCCdinf:
			err = parseDinf(p, r, a)
			if err != nil {
				goto T
			}
			break
		case fourCCstbl:
			err = parseStbl(p, r, a)
			if err != nil {
				goto T
			}
			break
		default:
			break
		}
	}
T:
	return err
}

// parse vmhd/smhd/hmhd/nmhd box
func parseXmhd(p *MovieInfo, _ *deMuxReader, a *atom) error {
	p.trak[len(p.trak)-1].mdia.minf.mediaInfoHeader = a.atomType
	return nil
}

// parse dinf box
func parseDinf(p *MovieInfo, r *deMuxReader, _ *atom) error {
	dinf := new(boxDinf)
	p.trak[len(p.trak)-1].mdia.dinf = dinf
	_, _ = r.GetAtomHeader() // dref box
	_, _ = r.ReadVersionFlags()
	dinf.entryCount = r.Read4()
	dinf.dataEntries = make(map[uint32]*dataEntry)
	for i := uint32(0); i < dinf.entryCount; i++ {
		dataAtom, _ := r.GetAtomHeader()
		entryFlag := r.Read4()
		tmpDataEntry := new(dataEntry)
		tmpDataEntry.entryFlag = entryFlag
		content, _, _ := r.ReadBytes(int(dataAtom.atomSize) - 12)
		tmpDataEntry.content = string(content)
		dinf.dataEntries[dataAtom.atomType] = tmpDataEntry
	}
	return nil
}

// parse stbl box
func parseStbl(p *MovieInfo, r *deMuxReader, a *atom) error {
	stbl := new(boxStbl)
	p.trak[len(p.trak)-1].mdia.stbl = stbl
	logD.Print("parse stbl box")
	stopPosition := r.Position() + a.atomSize
	var err error = nil
	for nextAtom := r.Position(); r.Position() < stopPosition; _ = r.MoveTo(nextAtom) {
		err = r.GetNextAtomData()
		if err != nil {
			return err
		}
		a, _ := r.GetAtomHeader()

		nextAtom += a.Size()
		if _, ok := stblParseTable[a.atomType]; ok {
			err = stblParseTable[a.atomType](p, r, a)
			if err != nil {
				logE.Printf("error encountered when parsing atom %s, error type: %T", a.Type(), err)
				return err
			}
		} else {
			logW.Printf("atom:%s will not be parsed", a.Type())
		}
	}
	return nil
}

// parse stsd box
func parseStsd(p *MovieInfo, r *deMuxReader, a *atom) error {
	stsdBoxSize := a.atomSize
	// endPosition := anchor + stsdBoxSize
	stsd := new(boxStsd)
	p.trak[len(p.trak)-1].mdia.stbl.stsd = stsd
	stsd.version, _ = r.ReadVersionFlags()
	stsd.entryCount = r.Read4()
	if stsd.entryCount <= 0 || stsd.entryCount >= uint32(stsdBoxSize)/8 {
		// logs.errLog.Println(" invalid stsd entry Data")
		return errors.New("invalid stsd entry Data")
	}
	var err error
	nextEntryPosition := r.Position()
	for i := 0; i < int(stsd.entryCount); i++ {
		_ = r.MoveTo(nextEntryPosition)
		entryAtom, _ := r.GetAtomHeader()
		nextEntryPosition = r.Position() + entryAtom.atomSize
		trackType := getTrackType(entryAtom.atomType)
		logD.Printf("parsing moov.trak.mdia.minf.stbl.stsd, track type(audio=0, video=1, subtitle=2) is %d, sample entry is %s", trackType, entryAtom)
		switch trackType {
		case AudioTrack:
			err = stsd.parseAudioSampleEntry(r, entryAtom, p.ftyp.isQuickTimeFormat)
			break
		case VideoTrak:
			err = stsd.parseVideoSampleEntry(r, entryAtom)
			break
		case SubtitleTrack:
			// errLog = stsd.parseSubtitleSampleEntry(readSeeker, entryAtom)
			break
		default:
			{
				err = errors.New("unsupported sample entry")
				break
			}
		}
		if err != nil {
			return err
		}
	}

	return err
}

// parse stts box
func parseStts(p *MovieInfo, r *deMuxReader, _ *atom) error {
	stts := new(boxStts)
	p.trak[len(p.trak)-1].mdia.stbl.stts = stts
	_, _ = r.ReadVersionFlags()
	stts.entryCount = r.Read4()
	for i := uint32(0); i < stts.entryCount; i++ {
		stts.sampleCount = append(stts.sampleCount, r.Read4())
		stts.sampleDelta = append(stts.sampleDelta, r.Read4())
	}
	return nil
}

// parse stsc box
func parseStsc(p *MovieInfo, r *deMuxReader, _ *atom) error {
	stsc := new(boxStsc)
	p.trak[len(p.trak)-1].mdia.stbl.stsc = stsc
	_, _ = r.ReadVersionFlags()
	stsc.entryCount = r.Read4()
	for i := uint32(0); i < stsc.entryCount; i++ {
		stsc.firstChunk = append(stsc.firstChunk, r.Read4())
		stsc.samplePerChunk = append(stsc.samplePerChunk, r.Read4())
		stsc.sampleDescriptionIndex = append(stsc.sampleDescriptionIndex, r.Read4())
	}
	return nil
}

// parse stsz box
func parseStsz(p *MovieInfo, r *deMuxReader, a *atom) error {
	stsz := new(boxStsz)
	p.trak[len(p.trak)-1].mdia.stbl.stsz = stsz
	stsz.atomType = a.atomType
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
			buf, _, _ := r.ReadBytes(int(byteSize))
			br := newBitReaderFromSlice(buf)
			stsz.entrySize = append(stsz.entrySize, uint32(br.ReadBitsLE16(byteSize)))
		}
	}
	return nil
}

// parse stco box
func parseStco(p *MovieInfo, r *deMuxReader, a *atom) error {
	stco := new(boxStco)
	p.trak[len(p.trak)-1].mdia.stbl.stco = stco
	_, _ = r.ReadVersionFlags()
	stco.entryCount = r.Read4()
	for i := uint32(0); i < stco.entryCount; i++ {
		if a.atomType == fourCCstco {
			stco.chunkOffset = append(stco.chunkOffset, uint64(r.Read4()))
		} else if a.atomType == fourCCco64 {
			stco.chunkOffset = append(stco.chunkOffset, r.Read8())
		}
	}
	return nil
}

// parse stss box
func parseStss(p *MovieInfo, r *deMuxReader, _ *atom) error {
	stss := new(boxStss)
	p.trak[len(p.trak)-1].mdia.stbl.stss = stss
	_, _ = r.ReadVersionFlags()
	stss.entryCount = r.Read4()
	for i := uint32(0); i < stss.entryCount; i++ {
		stss.sampleNumber = append(stss.sampleNumber, r.Read4())
	}
	return nil
}

// parse pssh box
func parsePssh(p *MovieInfo, r *deMuxReader, a *atom) error {
	if r.CheckEnoughAtomData(a.atomSize) != nil {
		return ErrNoEnoughData
	}
	pssh := new(Pssh)
	version, _ := r.ReadVersionFlags()
	pssh.SystemId, _, _ = r.ReadBytes(16)
	if version > 0 {
		pssh.KIdCount = r.Read4()
		for i := uint32(0); i < pssh.KIdCount; i++ {
			kid, _, _ := r.ReadBytes(16)
			pssh.KId = append(pssh.KId, kid)
		}
	}
	pssh.DataSize = r.Read4()
	pssh.Data, _, _ = r.ReadBytes(int(pssh.DataSize))
	p.pssh = append(p.pssh, pssh)
	return nil
}

// parse moof box
func parseMoof(p *MovieInfo, r *deMuxReader, a *atom) error {
	logD.Print("parse moof box", p.movieHeader)
	stopPosition := r.Position() + a.atomSize
	// p.topLevelType = fourCCmoov
	// return parseStblBoxes(p, readSeeker, a)
	var err error = nil
	for nextAtom := r.Position(); r.Position() < stopPosition; _ = r.MoveTo(nextAtom) {
		err = r.GetNextAtomData()
		if err != nil {
			return err
		}
		a, _ := r.GetAtomHeader()
		nextAtom += a.Size()
		switch a.atomType {
		case fourCCmfhd:
			r.Move(4)
			p.sequenceNumber = r.Read4()
			logD.Print("     parsing moof.mfhd")
			break
		case fourCCtraf:
			err = parseTraf(p, r, a)
			if err != nil {
				logD.Print("when parsing traf , error returned")
				goto T
			}
			break
		case fourCCpssh:
			err = parsePssh(p, r, a)
			if err != nil {
				goto T
			}
			break
		case fourCCmeta:
			break
		default:
			break
		}
	}
T:
	return err
}

// parse traf box
func parseTraf(p *MovieInfo, r *deMuxReader, a *atom) error {
	traf := new(boxTraf)
	p.traf = append(p.traf, traf)
	logD.Print("     parsing moof.traf")
	stopPosition := r.Position() + a.atomSize
	var err error = nil
	for nextAtom := r.Position(); r.Position() < stopPosition; _ = r.MoveTo(nextAtom) {
		a, _ := r.GetAtomHeader()
		nextAtom += a.Size()
		if _, ok := trafParseTable[a.atomType]; ok {
			err = trafParseTable[a.atomType](p, r, a)
			if err != nil {
				logE.Printf("error encountered when parsing atom %s, error type: %T", a.Type(), err)
				goto T
			}
		} else {
			logW.Printf("atom:%s will not be parsed", a.Type())
		}
	}
T:
	if err != nil {
		p.traf = p.traf[:len(p.traf)-1]
	}
	return err
}

// parse tfhd box
func parseTfhd(p *MovieInfo, r *deMuxReader, _ *atom) error {
	tfhd := new(boxtfhd)
	tfhd.tfFlags = r.Read4() & 0x00FFFFFF
	tfhd.trackId = r.Read4()
	if tfhd.tfFlags&0x000001 != 0 {
		tfhd.baseDataOffset = new(uint64)
		*tfhd.baseDataOffset = r.Read8()
	} else if tfhd.tfFlags&0x000002 != 0 {
		tfhd.sampleDescriptionIndex = new(uint32)
		*tfhd.sampleDescriptionIndex = r.Read4()
	} else if tfhd.tfFlags&0x000008 != 0 {
		tfhd.defaultSampleDuration = new(uint32)
		*tfhd.defaultSampleDuration = r.Read4()
	} else if tfhd.tfFlags&0x000010 != 0 {
		tfhd.defaultSampleSize = new(uint32)
		*tfhd.defaultSampleSize = r.Read4()
	} else if tfhd.tfFlags&0x000020 != 0 {
		tfhd.defaultSampleFlags = new(uint32)
		*tfhd.defaultSampleFlags = r.Read4()
	} else if tfhd.tfFlags&0x000001 == 0 && tfhd.tfFlags&0x020000 != 0 {
		tfhd.defaultBaseIsMoof = true
	}
	p.traf[len(p.traf)-1].tfhd = tfhd
	return nil
}

// parse trun box
func parseTrun(p *MovieInfo, r *deMuxReader, _ *atom) error {
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
				bytesInt, _, _ := r.ReadBytes(4)
				*sampleTrun.sampleCompositionTimeOffset = int32(bytesInt[0])<<24 | int32(bytesInt[1])<<16 | int32(bytesInt[2])<<8 | int32(bytesInt[3])
			}
		}
		trun.samples = append(trun.samples, sampleTrun)
	}
	p.traf[len(p.traf)-1].trun = append(p.traf[len(p.traf)-1].trun, trun)
	return nil
}

// parse tfdt box
func parseTfdt(p *MovieInfo, r *deMuxReader, _ *atom) error {
	version, _ := r.ReadVersionFlags()
	if version != 0 {
		p.traf[len(p.traf)-1].baseMediaDecodeTime = r.Read8()
	} else {
		p.traf[len(p.traf)-1].baseMediaDecodeTime = uint64(r.Read4())
	}
	return nil
}

// parse saio box
func parseSaio(p *MovieInfo, r *deMuxReader, _ *atom) error {
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
	if p.topLevelType == fourCCmoov {
		p.trak[len(p.trak)-1].mdia.stbl.saio = append(p.trak[len(p.trak)-1].mdia.stbl.saio, saio)
	} else if p.topLevelType == fourCCmoof {
		p.traf[len(p.traf)-1].saio = append(p.traf[len(p.traf)-1].saio, saio)
	}
	return nil
}

// parse saiz box
func parseSaiz(p *MovieInfo, r *deMuxReader, _ *atom) error {
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
	if p.topLevelType == fourCCmoov {
		p.trak[len(p.trak)-1].mdia.stbl.saiz = append(p.trak[len(p.trak)-1].mdia.stbl.saiz, saiz)
	} else if p.topLevelType == fourCCmoof {
		p.traf[len(p.traf)-1].saiz = append(p.traf[len(p.traf)-1].saiz, saiz)
	}
	return nil
}

// parse sbgp box
func parseSbgp(p *MovieInfo, r *deMuxReader, _ *atom) error {
	sbgp := new(boxSbgp)
	version, _ := r.ReadVersionFlags()
	sbgp.groupingType = r.Read4()
	if version == 1 {
		sbgp.groupingTypeParameter = new(uint32)
		*sbgp.groupingTypeParameter = r.Read4()
	}
	sbgp.entryCount = r.Read4()
	for i := uint32(0); i < sbgp.entryCount; i++ {
		sampleCountT := r.Read4()
		groupDescriptionIndexT := r.Read4()
		sbgp.sampleGroupDescriptionIndexes = append(sbgp.sampleGroupDescriptionIndexes, sampleGroupDescriptionIndex{sampleCount: sampleCountT, groupDescriptionIndex: groupDescriptionIndexT})
	}
	if p.topLevelType == fourCCmoov {
		p.trak[len(p.trak)-1].mdia.stbl.sbgp = sbgp
	} else if p.topLevelType == fourCCmoof {
		p.traf[len(p.traf)-1].sbgp = sbgp
	}
	return nil
}

// parse sgpd box
func parseSgpd(p *MovieInfo, r *deMuxReader, _ *atom) error {
	sgpd := new(boxSgpd)
	version, _ := r.ReadVersionFlags()
	sgpd.groupingType = r.Read4()
	if int2String(sgpd.groupingType) != "seig" {
		return ErrUnsupportedSampleGroupType
	}
	if version == 1 {
		sgpd.defaultLength = new(uint32)
		*sgpd.defaultLength = r.Read4()
		if *sgpd.defaultLength == 0 || *sgpd.defaultLength > 20 {
			return ErrUnsupportedVariableSampleGroupLength
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
					return ErrInvalidLengthOfSampleGroup
				}
			}
		}
		// CencSampleEncryptionInformationGroupEntry
		// only support "cenc" scheme currently
		cencGroupEntry := new(cencSampleEncryptionInformationGroupEntry)
		r.Move(1) // reversed 1 byte
		byteT := r.ReadUnsignedByte()
		cencGroupEntry.cryptByteBlock = byteT >> 4
		cencGroupEntry.skipByteBlock = byteT & 0x0F
		cencGroupEntry.isProtected = r.ReadUnsignedByte() != 0
		cencGroupEntry.perSampleIVSize = r.ReadUnsignedByte()
		if cencGroupEntry.perSampleIVSize != 8 && cencGroupEntry.perSampleIVSize != 16 {
			return ErrInvalidLengthOfIVInSampleGroup
		}
		cencGroupEntry.kID, _, _ = r.ReadBytes(16)
		if cencGroupEntry.isProtected && cencGroupEntry.perSampleIVSize == 0 {
			constIVSize := r.ReadUnsignedByte()
			if constIVSize != 8 && constIVSize != 16 {
				return ErrInvalidLengthOfIVInSampleGroup
			}
			cencGroupEntry.constantIV, _, _ = r.ReadBytes(int(constIVSize))
		}
		sgpd.cencGroupEntries = append(sgpd.cencGroupEntries, cencGroupEntry)
	}
	if p.topLevelType == fourCCmoov {
		p.trak[len(p.trak)-1].mdia.stbl.sgpd = sgpd
	} else if p.topLevelType == fourCCmoof {
		p.traf[len(p.traf)-1].sgpd = sgpd
	}
	return nil
}

// parse senc box
func parseSenc(p *MovieInfo, r *deMuxReader, a *atom) error {
	senc := new(boxSenc)
	_, senc.flags = r.ReadVersionFlags()
	senc.sampleCount = r.Read4()
	currentTraf := p.traf[len(p.traf)-1]
	// get the trak info from moov
	trak, err := findTrak(p.movieHeader, currentTraf.tfhd.trackId)
	var iVSize uint8
	if err != nil {
		logW.Print("cannot find the trak/moov, so the default IV size cannot be set. Maybe the parsing of senc will be fault. error type:", err)
		iVSize = 0
	} else {
		tenc := trak.mdia.stbl.stsd.protectedInfo
		iVSize = tenc.DefaultPerSampleIVSize
	}
	sampleGroupDescriptionIndexList := currentTraf.getSampleGroupDescriptionIndexList()
	for i := uint32(0); i < senc.sampleCount; i++ {
		if currentTraf.sgpd != nil && currentTraf.sbgp != nil && i < uint32(len(sampleGroupDescriptionIndexList)) {
			index := sampleGroupDescriptionIndexList[i]
			if index != 0 {
				if index > 65536 {
					index -= 65536
				}
				if index < uint32(len(currentTraf.sgpd.cencGroupEntries)) {
					cencSampleEncryptionInformationEntry := currentTraf.sgpd.cencGroupEntries[index-1]
					if cencSampleEncryptionInformationEntry != nil {
						iVSize = cencSampleEncryptionInformationEntry.perSampleIVSize
					}
				}
			}
		}
		if iVSize == 0 {
			// try to detect the iv size
			n, err := senc.tryToDetectIVSize(r, uint32(a.atomSize)-8)
			if err != nil {
				return errors.New("failed to parse senc box, because the IV size is invalid")
			}
			iVSize = n
		}
		sampleEnc := new(sampleEncryption)
		sampleEnc.IV, _, _ = r.ReadBytes(int(iVSize))
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
	currentTraf.senc = senc
	return nil
}

// boxSenc.tryToDetectIVSize try to detect the IV size in the absence of movie header or sample-group information.
func (p *boxSenc) tryToDetectIVSize(r *deMuxReader, left uint32) (uint8, error) {
	resumePosition := r.Position()
	if left <= 8 {
		return 0, errors.New("too small for left size")
	}
	sencSampleUnitSize := left / p.sampleCount
	if p.flags&0x000002 == 0 {
		if sencSampleUnitSize != 8 && sencSampleUnitSize != 16 {
			return 0, errors.New("failed to detect the IV size")
		}
		return uint8(sencSampleUnitSize), nil
	}
	// has subSample, IV'size is 8 or 16
	increaseStep := 8
	ivSize := 0
DETECTOR:
	_ = r.MoveTo(resumePosition)
	ivSize += increaseStep
	if ivSize > 16 {
		return 0, errors.New("failed to detect the IV size")
	}

	r.Move(int64(ivSize))
	subSampleCount := r.Read2()
	if uint32(subSampleCount)*(4+2) != (left - uint32(ivSize) - 4) {
		goto DETECTOR
	} else {
		_ = r.MoveTo(resumePosition)
		return uint8(ivSize), nil
	}
}

// parse Subs Box
func parseSubs(p *MovieInfo, r *deMuxReader, _ *atom) error {
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
	if p.topLevelType == fourCCmoov {
		p.trak[len(p.trak)-1].mdia.stbl.subs = subs
	} else if p.topLevelType == fourCCmoof {
		p.traf[len(p.traf)-1].subs = subs
	}
	return nil
}
