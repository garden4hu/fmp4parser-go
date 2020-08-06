package fmp4parser

import (
	"io"
	"os"
)

// Fmp4Parser is a object type for fmp4parser
type Track struct {
	Type         TrackType // Audio/Video/Subtitle
	TrackId      uint32    // The unique Id of a track
	Format       string    // If track is encrypted, it's from enca/encv; If not, it's sampleEntry
	Duration     uint64    // If track is VOD, it's duration by timeScale of the track; If not, this parameter is not accurate
	TimeScale    uint64    //
	ChannelCount uint32    // For audio track
	SampleSize   uint32    // Default sample size
	SampleRate   uint32    // For audio track
	StartTime    uint64
	FrameRate    uint32
	Dts          uint64
	Codec        CodecType // The codec

	// If there is movie fragment
	FragmentDuration              uint64 // it's the longest duration of fragment track. If stream is real-time, this parameter is not accurate
	DefaultSampleDescriptionIndex uint32
	DefaultSampleDuration         uint32
	DefaultSampleSize             uint32
	DefaultSampleFlags            uint32

	ExtraData            interface{}          // audio descriptor OR video codec configuration record
	EncryptedInformation ProtectedInformation // Track encryption information
}

type Movie struct {
	Tracks map[uint32]Track // Key is TrackId of Track, Value is Track
	Psshs  []Pssh           // Protection system specific header
	// sidx []boxSidx
}

// A Parser reads and parse IBFF from an input stream.
type Parser struct {
	m *mediaInfo
}

// NewFmp4Parser return the object for handle the fmp4parser
func NewFmp4Parser(r io.ReadSeeker) *Parser {
	m := newMediaInfo(r)
	newLog(os.Stdout) // create log obj
	return &Parser{m: m}
}

func (p *Parser) GetMediaInformation() (*Movie, error) {

	return nil, nil
}

func (p *Parser) GetVerboseMediaInformation() (*MovieInfo, error) {

	return nil, nil
}

// GetSample return the parsed sample.
// Warn: Need re-consider the signature of function.
func (p *Parser) GetSample() error {
	return nil
}

// SetReader allow to reset the io br without changing the status of internal
func (p *Parser) SetReader(r io.ReadSeeker) {
	p.m.r.ResetReader(r)
}

// Parse performs parsing operations
func (p *Parser) Parse() error {
	_ = p.m.parse()
	return nil
}

func (p *Parser) GGG() {

}
