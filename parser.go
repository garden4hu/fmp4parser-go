package main

import (
	"errors"
	"fmt"
	"io"
)

type mediaInfo struct {
	r     *mp4Reader
	movie *MovieInfo
	moof  *movieFragment // current

	dataPos int64

	// for internal usage
	currentState parsingState
	leftAtomSize uint64
}

func newMediaInfo(r io.ReadSeeker) *mediaInfo {
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
	stateParsingEnd                      // parsing finished
)

func (p *mediaInfo) parseInternal() (err error) {

	for p.currentState != stateParsingEnd {
		if p.currentState == stateParsingIDLE {
			_, err = p.checkStatus() // get the state and atomReader
			if err != nil {
				logD.Println(err)
				return err
			}
		}
		switch p.currentState {
		case stateParsingFTYP:
			ftypReader, e := p.r.GetAtom()
			if e != nil {
				logE.Println(e)
				return e
			}
			_ = parseFtyp(p.movie, ftypReader)
			p.currentState = stateParsingIDLE
			break
		case stateParsingMOOV:
			movieReader, e := p.r.GetAtom()
			if e != nil {
				logE.Println(e)
				return e
			}
			e = parseMoov(p.movie, movieReader)
			if e != nil {
				return e
			}
			p.currentState = stateParsingIDLE
			break
		case stateParsingMOOF:
			moofReader, e := p.r.GetAtom()
			if e != nil {
				return e
			}
			p.moof = newMovieFragment(p.movie)
			e = parseMoof(p.moof, moofReader)
			if e != nil {
				return e
			}
			p.currentState = stateParsingIDLE
			break
		case stateParsingSIDX:
			sidxReader, e := p.r.GetAtom()
			if e != nil {
				return e
			}
			parseSidx(p.movie, sidxReader)
			p.currentState = stateParsingIDLE
			break
		case stateParsingSSIX:
			ssixReader, e := p.r.GetAtom()
			if e != nil {
				return e
			}
			parseSsix(p.movie, ssixReader)
			p.currentState = stateParsingIDLE
			break
		case stateParsingMDAT:
			p.dataPos = p.r.GetAtomPosition()
			if p.movie == nil || !p.movie.parsedProfile {
				err = p.r.SkipCurrentAtom()
				if err != nil {
					return err
				}
			}
			p.currentState = stateParsingIDLE
			break
		}
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
		a, e := p.r.PeekAtomHeader()
		if e != nil {
			if e == io.EOF {
				p.currentState = stateParsingEnd
			}
			return nil, e
		}
		fmt.Println(a.Type())
		if fourCCftyp == a.atomType || fourCCstyp == a.atomType || fourCCmoov == a.atomType || fourCCmoof == a.atomType || fourCCsidx == a.atomType || fourCCmdat == a.atomType || fourCCssix == a.atomType {
			if p.movie == nil {
				// when meeting the FourCC above, the MovieInfo will be created.
				p.movie = new(MovieInfo)
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
		} else if fourCCskip == a.atomType || fourCCfree == a.atomType || fourCCpdin == a.atomType || fourCCprft == a.atomType || fourCCmeta == a.atomType {
			_ = p.r.SkipCurrentAtom()
			continue
		} else {
			_ = p.r.SkipCurrentAtom()
		}
	}
	//return nil, ErrUnsupportedAtomType
}
