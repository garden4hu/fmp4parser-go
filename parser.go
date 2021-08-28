package fmp4parser

import (
	"errors"
	"io"
)

type mediaInfo struct {
	r    *mp4Reader
	moov *MovieInfo
	moof *movieFragment // current

	mediaNotifier MediaNotifier
	// for internal usage
	currentState parsingState
	leftAtomSize uint64
}

func newMediaInfo(r io.Reader) *mediaInfo {
	return &mediaInfo{r: newMp4Reader(r),
		currentState: stateParsingIDLE,
	}
}

type parsingState uint32

const (
	stateParsingIDLE parsingState = iota //  idle
	stateParsingFTYP                     // parsing "ftyp"/"styp"
	stateParsingMOOV                     // parsing "moov"
	stateParsingMOOF                     // parsing "moof"
	stateParsingSIDX                     // parsing "sidx"
	stateParsingSSIX                     // parsing "ssix"
	stateParsingMDAT                     // parsing "mdat"
)

func (p *mediaInfo) parseInternal() (err error) {
	var curAtom *atom
	switch p.currentState {
	case stateParsingIDLE:
		curAtom, err = p.checkStatus()
		if err != nil {
			return nil
		}
		fallthrough
	case stateParsingFTYP:
		ftypReader, e := p.r.GetAtomReader(curAtom)
		if e != nil {
			return e
		}
		_ = parseFtyp(p.moov, ftypReader)
		break
	case stateParsingMOOV:
		e := parseMoov(p.moov, p.r, curAtom)
		if e != nil {
			return e
		}

		break
	case stateParsingMOOF:
		moofReader, e := p.r.GetAtomReader(curAtom)
		if e != nil {
			return e
		}
		p.moof = newMovieFragment(p.moov)
		e = parseMoof(p.moof, moofReader)
		if e != nil {
			return e
		}
		break
	case stateParsingSIDX:
		sidxReader, e := p.r.GetAtomReader(curAtom)
		if e != nil {
			return e
		}
		parseSidx(p.moov, sidxReader)
		break
	case stateParsingSSIX:
		ssixReader, e := p.r.GetAtomReader(curAtom)
		if e != nil {
			return e
		}
		parseSsix(p.moov, ssixReader)
		break
	case stateParsingMDAT:
		break
	}
	return nil
}

func (p *MovieInfo) GenerateMovie() (*Movie, error) {
	if p == nil || p.topLevelType != fourCCmoov {
		return nil, errors.New("failed to generate Movie information, because the MovieInfo  is null or not moov information")
	}
	// movie  := new(Movie)
	return nil, nil
}

func (p *mediaInfo) checkStatus() (*atom, error) {
	for {
		a, e := p.r.ReadAtomHeader()
		if e != nil {
			return nil, e
		}
		if fourCCftyp == a.atomType || fourCCstyp == a.atomType || fourCCmoov == a.atomType || fourCCmoof == a.atomType || fourCCsidx == a.atomType || fourCCssix == a.atomType {
			if p.moov == nil {
				// when meeting the FourCC above, the MovieInfo will be created.
				p.moov = new(MovieInfo)
			}
			switch a.atomType {
			case fourCCftyp:
				fallthrough
			case fourCCstyp:
				p.currentState = stateParsingFTYP
				break
			case fourCCmoov:
				p.currentState = stateParsingMOOV
				break
			case fourCCmoof:
				p.currentState = stateParsingMOOF
				break
			case fourCCsidx:
				p.currentState = stateParsingSIDX
				break
			case fourCCssix:
				p.currentState = stateParsingSSIX
				break
			case fourCCmdat:
				p.currentState = stateParsingMDAT
				break
			}
			return a, nil
		} else if fourCCskip == a.atomType || fourCCfree == a.atomType || fourCCpdin == a.atomType || fourCCprft == a.atomType {
			b := make([]byte, a.bodySize)
			if e := p.r.ReadAtomBodyFull(b); e != nil {
				return nil, e
			}
			continue
		} else {
			break
		}
	}
	return nil, ErrInvalidMP4Format
}
