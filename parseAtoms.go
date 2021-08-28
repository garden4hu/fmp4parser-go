package fmp4parser

import "fmt"

// parse ftyp box
func parseFtyp(p *MovieInfo, r *atomReader) error {
	err := error(nil)
	logD.Print("parseConfig ftyp box")
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
	return err
}

// parse ssix box (SubSegment Index box)
func parseSsix(p *MovieInfo, r *atomReader) {
	_ = r.Move(8) // version + flags
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

// parse moov box
// The 2-th parameter is mp4Reader instead of atomReader is because of there is
// a chance the "moov" is huge. Sending a reading pointer without buffer is a wise move.
func parseMoov(p *MovieInfo, r *mp4Reader, a *atom) error {
	logD.Println("parseConfig moov box")
	logD.Println(a)
	left := a.bodySize
	for {
		if left < 8 {
			break
		}
		atomHeader, err := r.ReadAtomHeader()
		if err != nil {
			return err
		}
		ar, err := r.GetAtomReader(atomHeader)
		if err != nil {
			return fmt.Errorf("%w failed to get data for parsing %s", err, atomHeader.Type())
		}
		left -= atomHeader.Size()
		if err = moovParseTable[atomHeader.atomType](p, ar); err != nil {
			return fmt.Errorf("%w failed to parseConfig atom %s", err, atomHeader.Type())
		}
	}
	return nil
}

// parse mvhd box
func parseMvhd(p *MovieInfo, r *atomReader) error {
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
	_ = r.Move(70)
	// 10 bytes reserved. 36 bytes matrix. 24 bytes pre_defined
	p.mvhd.nextTrackId = r.Read4()
	return nil
}

// parse mvex box
func parseMvex(p *MovieInfo, r *atomReader) error {

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

	logD.Print("parsing moov.mvex, ", r.a)
	p.mvex = new(boxMvex)
	p.hasFragment = true
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
		if ar.a.atomType == fourCCmehd {
			parseMehd(p.mvex, ar)
		} else if ar.a.atomType == fourCCtrex {
			parseTrex(p.mvex, ar)
		} else if ar.a.atomType == fourCCleva {
			parseLeva(p.mvex, ar)
		} else {
			continue
		}
	}
}

// parse trak box
func parseTrak(p *MovieInfo, r *atomReader) error {

	trak := new(boxTrak)
	trak.quickTimeFormat = p.ftyp.isQuickTimeFormat
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
		if ar.a.atomType == fourCCtkhd {
			trak.parseTkhd(ar)
		} else if ar.a.atomType == fourCCedts {
			trak.parseEdts(ar)
		} else if ar.a.atomType == fourCCmdia {
			if err := trak.parseMdia(ar); err != nil {
				return err
			}
		} else {
			continue
		}
	}

}

// parse moof box
func parseMoof(p *movieFragment, r *atomReader) error {
	var e error = nil
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
