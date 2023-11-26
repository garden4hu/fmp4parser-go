package main

import (
	"errors"
)

// parse ftyp box
func parseFtyp(p *MovieInfo, r *atomReader) error {
	err := error(nil)
	p.ftyp = new(boxFtyp)
	p.ftyp.majorBrand = r.Read4()
	// check QuickTime format
	if p.ftyp.majorBrand == 0x71742020 {
		p.ftyp.isQuickTimeFormat = true
	}
	p.ftyp.minorVersion = r.Read4()
	for brandCount := (r.Size() - 4) / 4; brandCount > 0; brandCount-- {
		compatibleBrand := r.Read4()
		if compatibleBrand == 0x71742020 {
			p.ftyp.isQuickTimeFormat = true
		}
		p.ftyp.compatibleBrands = append(p.ftyp.compatibleBrands, compatibleBrand)
	}
	logD.Println(p.ftyp)
	return err
}

// parse ssix box (SubSegment Index box)
func parseSsix(p *MovieInfo, r *atomReader) {
	_ = r.Move(8) // version + flags
	ssix := new(boxSsix)
	ssix.subSegmentCount = r.Read4()
	for i := 0; i < int(ssix.subSegmentCount); i++ {
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
}

// parse sdix box (Segment Index box)
func parseSidx(p *MovieInfo, r *atomReader) {
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
	_ = r.Move(2) // reserved
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
}

// parseMoov parse moov box
func parseMoov(movie *MovieInfo, reader *atomReader) error {
	logD.Println("parsing moov box")
	for {
		itemReader, err := reader.GetSubAtom()
		if err != nil {
			if err == ErrNoMoreAtom {
				return nil
			}
			return nil
		}
		logD.Println("MOOV type=", itemReader.a.Type())
		switch itemReader.TypeCC() {
		case fourCCmvhd:
			movie.parseMvhd(itemReader)
		case fourCCmvex:
			err = movie.parseMvex(itemReader)
			break
		case fourCCtrak:
			err = movie.parseTrak(itemReader)
			break
		}
		if err != nil {
			return nil
		}
	}
	if movie.hasFragment == false {
		// if moov doesn't have fragment, then we can build the sample list
	}
	return nil
}

// parse mvhd box
func (movie *MovieInfo) parseMvhd(r *atomReader) {
	version, _ := r.ReadVersionFlags()
	if version == 1 {
		movie.creationTime = r.Read8()
		movie.modificationTime = r.Read8()
		movie.timeScale = r.Read4()
		movie.duration = r.Read8()
	} else {
		movie.creationTime = uint64(r.Read4())
		movie.modificationTime = uint64(r.Read4())
		movie.timeScale = r.Read4()
		movie.duration = uint64(r.Read4())
	}
	_ = r.Move(70)
	// 10 bytes reserved. 36 bytes matrix. 24 bytes pre_defined
	movie.nextTrackId = r.Read4()
}

// parse mvex box
func (movie *MovieInfo) parseMvex(reader *atomReader) error {
	parseTrex := func(p *boxMvex, r *atomReader) {
		logD.Print("parsing moov.mvex.trex, ", r.a)
		trex := new(boxTrex)
		_ = r.Move(8)
		// Version + flags
		trex.trackId = r.Read4()
		trex.defaultSampleDescriptionIndex = r.Read4()
		trex.defaultSampleDuration = r.Read4()
		trex.defaultSampleSize = r.Read4()
		trex.defaultSampleFlags = r.Read4()
		p.trex = append(p.trex, *trex)
	}

	parseMehd := func(p *boxMvex, r *atomReader) {
		v, _ := r.ReadVersionFlags()
		if v == 1 {
			p.fragmentDuration = r.Read8()
		} else {
			p.fragmentDuration = uint64(r.Read4())
		}
	}

	parseLeva := func(p *boxMvex, r *atomReader) {
		logD.Print("parsing moov.mvex.leva, ", r.a)
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
		p.leva = leva
	}

	parseTrep := func(p *boxMvex, r *atomReader) {
		logD.Print("parsing moov.mvex.trep", r.a)
		trep := new(boxTrep)
		trep.trackId = r.Read4()
		_, e := r.FindSubAtom(fourCCcslg)
		if e == nil {
			// found cslg box. There is no
			//normative processing associated with this box.
			logW.Print("trep box is supported")
		}
	}

	movie.mvex = new(boxMvex)
	movie.hasFragment = true
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
		case fourCCmehd:
			parseMehd(movie.mvex, itemReader)
			break
		case fourCCtrex:
			parseTrex(movie.mvex, itemReader)
			break
		case fourCCleva:
			parseLeva(movie.mvex, itemReader)
			break
		case fourCCtrep:
			parseTrep(movie.mvex, itemReader)
			break
		}
	}
	return nil
}

// parse trak box
func (movie *MovieInfo) parseTrak(reader *atomReader) error {
	trak := new(boxTrak)
	trak.quickTimeFormat = movie.ftyp.isQuickTimeFormat
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
		case fourCCtkhd:
			trak.parseTkhd(itemReader)
			break
		case fourCCedts:
			trak.parseEdts(itemReader)
			break
		case fourCCmdia:
			err = trak.parseMdia(itemReader)
			break
		default:
			break
		}
		if err != nil {
			return err
		}
	}
	//trak.constructPacketList()
	return nil
}

// parse moof box
func parseMoof(p *movieFragment, r *atomReader) error {
	var e error = nil
	for {
		ar, e := r.GetSubAtom()
		if e != nil {
			if errors.Is(ErrNoMoreAtom, e) {
				return nil
			}
			if errors.Is(ErrNoEnoughData, e) {
				return e
			}
		}
		switch ar.a.atomType {
		case fourCCmfhd:
			_ = ar.Move(4) // version + flags
			p.sequenceNumber = ar.Read4()
			break
		case fourCCtraf:
			e := p.parseTraf(ar)
			if e != nil {
				return e
			}
		default:
			break
		}

	}
	return e
}
