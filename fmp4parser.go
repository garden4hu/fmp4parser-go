package main

import (
	"io"
	"os"
)

type Packet struct {
	Duration        uint32 // in ms
	DTS             uint64
	PTS             uint64
	Data            []byte
	Size            uint32
	DescriptorIndex int
	offset          uint64
}

// Track is the struct of track in a media source file.
// It contains the overall information about the track.
type Track struct {
	Type    TrackType // Audio/Video/Subtitle
	TrackID uint32    // The unique Id of a track
	Codec   CodecType // The codec
	Format  string    // If track is encrypted, it's from enca/encv; If not, it's sampleEntry

	Duration  uint64 // If track is VOD, it's duration by timeScale of the track; If not, this parameter is not accurate
	TimeScale uint32 //
	// for audio
	ChannelCount uint16 // For audio track
	SampleSize   uint32 // Default sample size
	SampleRate   uint32 // For audio track
	//StartTime    uint64

	// for video
	Width  uint16 // picture width
	Height uint16 // picture height

	DTS uint64 // decoding timestamp
	PTS uint64 // presentation timestamp

	ExtraRawData         map[CodecType][]byte  // audio descriptor OR video codec configuration record
	EncryptedInformation *ProtectedInformation // Track encryption information
}

type Movie struct {
	Duration  uint64            // The duration of longest track
	TimeScale uint32            // The unit of duration
	Tracks    map[uint32]*Track // Key is TrackId of Track, Value is Track
	PSSHs     []*PSSH           // Protection system specific header
	// sidx []boxSidx
}

// A Parser reads and parse BMFF from an input stream.
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
	// p.m.readSeeker.ResetReader(readSeeker)
}

// Parse performs parsing operations
func (p *Parser) Parse() error {
	_ = p.m.parseInternal()
	return nil
}

func (p *Parser) GetTracks() []Track {
	return nil
}

func (p *Parser) GetTrackCounts() int {
	return 0
}

func (p *Parser) GetAudioTracks() []Track {
	return nil
}

func (p *Parser) GetVideoTracks() []Track {
	return nil
}

func (p *Parser) GetSubtitleTracks() []Track {
	return nil
}

func (p *Parser) GetPacket(trackID int) Packet {
	return Packet{}
}

func (p *Parser) GetMetaData() {

}
