package fmp4parser

import (
	"errors"
	"io"
)

// internal function of fmp4parser

type mediaInfo struct {
	r    *deMuxReader
	moov *MovieInfo
	moof []MovieInfo
}

func newMediaInfo(r io.ReadSeeker) *mediaInfo {
	// logD.Print(readSeeker)
	de := NewDeMuxReader(r)
	return &mediaInfo{r: de}
}

//
func (p *mediaInfo) parse() error {
	var err error = nil
	nextAtom := p.r.Position()
	// Read the first atom
	// if p.readSeeker.GetNextAtomData() != nil{
	// 	return io.ErrUnexpectedEOF
	// }
	for {
		_ = p.r.MoveTo(nextAtom)
		a, err := p.r.GetAtomHeader()
		if err != nil {
			logW.Print("break of parsing top level atoms, error type: ", err)
			break
		}
		nextAtom += a.Size()
		// "mdat" box will be omitted if it shows in the top level.
		// TODO. should record the start position of mdat for getting sample
		if a.atomType == fourCCmdat {
			logD.Printf("find mdat atom(size=%d) in movie level, skip it.", a.atomSize)
			_ = p.r.Skip(a.atomSize)
			continue
		}
		logD.Print("current parsing the box ", a)

		switch a.atomType {
		case fourCCftyp:
			fallthrough
		case fourCCstyp:
			fallthrough
		case fourCCmoov:
			fallthrough
		case fourCCssix:
			fallthrough
		case fourCCsidx:
			if p.moov == nil {
				p.moov = new(MovieInfo)
			}
			p.moov.topLevelType = fourCCmoov
			if _, ok := topLevelParseTable[a.atomType]; ok {
				err = topLevelParseTable[a.atomType](p.moov, p.r, a)
				if err != nil {
					goto TError
				}
			} else {
				logW.Printf("atom:%s will not be parsed", a.Type())
			}
			break
		case fourCCmoof:
			moof := new(MovieInfo)
			moof.topLevelType = fourCCmoof
			moof.movieHeader = p.moov
			//dbg
			logD.Print(p.moov)
			err = parseMoof(moof, p.r, a)
			if err != nil {
				goto TError
			}
			p.moof = append(p.moof, *moof)
			break
		case fourCCmeta:
			meta := new(MovieInfo)
			err = metaParseTable[a.atomType](meta, p.r, a)
			if err != nil {
				goto TError
			}
			break
		case fourCCmfra:
			fallthrough
		case fourCCfree:
			fallthrough
		case fourCCskip:
			fallthrough
		case fourCCpdin:
			break
		default:
			// unsupported/unparsed atom
			break
		}
	}

TError:
	return err

}

func findTrak(p *MovieInfo, trakId uint32) (*boxTrak, error) {
	if p == nil {
		logE.Print("movie header not set, pointer=nil")
		return nil, errors.New("not moov atom")
	}
	if len(p.trak) == 1 {
		return p.trak[0], nil
	}
	for i := 0; i < len(p.trak); i++ {
		if p.trak[i].tkhd.trackId == trakId {
			return p.trak[i], nil
		}
	}
	logW.Print("not find the specific track")
	return nil, ErrNotFoundTrak
}

func (p *boxTraf) getSampleGroupDescriptionIndexList() []uint32 {
	var sampleGroupDescriptionIndexList []uint32
	if p.sbgp != nil {
		for j := uint32(0); j < p.sbgp.entryCount; j++ {
			for k := uint32(0); k < p.sbgp.sampleGroupDescriptionIndexes[j].sampleCount; k++ {
				sampleGroupDescriptionIndexList = append(sampleGroupDescriptionIndexList, p.sbgp.sampleGroupDescriptionIndexes[j].groupDescriptionIndex)
			}
		}
	}
	return sampleGroupDescriptionIndexList
}
func (p *MovieInfo) GenerateMovie() (*Movie, error) {
	if p == nil || p.topLevelType != fourCCmoov {
		return nil, errors.New("failed to generate Movie information, because the MovieInfo  is null or not moov information")
	}
	// movie  := new(Movie)
	return nil, nil
}
