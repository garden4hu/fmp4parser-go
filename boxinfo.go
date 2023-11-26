package main

import (
	"fmt"
)

type TrackType uint32

const (
	UnknownTrack TrackType = iota
	AudioTrack
	VideoTrack
	SubtitleTrack
)

// encryption scheme type
var (
	encryptionSchemeTypeCENC uint32 = 0x63656E63 // "cenc"
	encryptionSchemeTypeCENS uint32 = 0x63656E73 // "cens"
	encryptionSchemeTypeCBCS uint32 = 0x63626373 // "cecs"
	encryptionSchemeTypeCBC1 uint32 = 0x63626331 // "cbc1"
)

type atom struct {
	atomType   uint32
	bodySize   int64 // body size
	headerSize uint32
}

func (a *atom) String() string {

	return fmt.Sprintf("Atom type:%s. Atom size:%d", a.Type(), a.Size())
}

func (a *atom) Type() string {
	return int2String(a.atomType)
}

func (a *atom) Size() int64 {
	return a.bodySize + int64(a.headerSize)
}

// ISO/IEC 14496-12 Part 12: ISO base media file format
// basic copy from https://github.com/mozilla/mp4parse-rust/blob/master/mp4parse/src/boxes.rs
var (
	fourCCftyp uint32 = 0x66747970 // "ftyp"
	fourCCstyp uint32 = 0x73747970 // "styp"
	fourCCmoov uint32 = 0x6d6f6f76 // "moov"
	fourCCsidx uint32 = 0x73696478 // "sidx"
	fourCCssix uint32 = 0x73736978 // "ssix"
	fourCCimda uint32 = 0x696D6461 // "imda"
	fourCCmdat uint32 = 0x6D646174 // "mdat"

	fourCCmvex uint32 = 0x6d766578 // "mvex"
	fourCCmehd uint32 = 0x6d656864 // "mehd"
	fourCCmeta uint32 = 0x6d657461 // "meta"
	fourCCtrep uint32 = 0x74726570 // "trep"
	fourCCtrex uint32 = 0x74726578 // "trex"
	fourCCleva uint32 = 0x6c657661 // "leva"

	fourCCmoof uint32 = 0x6D6F6F66 // "moof" 	fragment-movie    ->
	fourCCmfhd uint32 = 0x6D666864 // "mfhd"
	fourCCtraf uint32 = 0x74726166 // "traf"
	fourCCtfhd uint32 = 0x74666864 // "tfhd"
	fourCCtrun uint32 = 0x7472756E // "trun"
	fourCCsbgp uint32 = 0x73626770 // "sbgp"
	fourCCsgpd uint32 = 0x73677064 // "sgpd"
	fourCCsenc uint32 = 0x73656e63 // "senc"
	fourCCsubs uint32 = 0x73756273 // "subs"
	fourCCsaiz uint32 = 0x7361697A // "saiz"
	fourCCsaio uint32 = 0x7361696F // "saio"
	fourCCtfdt uint32 = 0x74666474 // "tfdt"  <- fragment-movie

	fourCCmfra uint32 = 0x6D667261 // "mfra"
	fourCCfree uint32 = 0x66726565 // "free"
	fourCCskip uint32 = 0x736b6970 // "skip"
	fourCCpdin uint32 = 0x7064696e // "pdin"
	fourCCuuid uint32 = 0x75756964 // "uuid"
	fourCCudta uint32 = 0x75647461 // "udta"
	fourCCprft uint32 = 0x70726674 // "prft"

	fourCCmvhd uint32 = 0x6d766864 // "mvhd"
	fourCCtrak uint32 = 0x7472616b // "trak"
	fourCCtkhd uint32 = 0x746b6864 // "tkhd"
	fourCCedts uint32 = 0x65647473 // "edts"
	fourCCmdia uint32 = 0x6d646961 // "mdia"
	fourCCmdhd uint32 = 0x6d646864 // "mdhd"
	fourCChdlr uint32 = 0x68646c72 // "hdlr"
	fourCCminf uint32 = 0x6d696e66 // "minf"
	fourCCelng uint32 = 0x656c6e67 // "elng"
	fourCCvmhd uint32 = 0x766D6864 // "vmhd"
	fourCCsmhd uint32 = 0x736D6864 // "smhd"
	// fourCChmhd uint32 = 0x686D6864 // "hmhd"
	// fourCCnmhd uint32 = 0x6E6D6864 // "nmhd"
	fourCCdinf uint32 = 0x64696E66 // "dinf"
	fourCCstbl uint32 = 0x7374626c // "stbl"
	fourCCstsd uint32 = 0x73747364 // "stsd"
	fourCCstts uint32 = 0x73747473 // "stts"
	fourCCstsc uint32 = 0x73747363 // "stsc"
	fourCCstsz uint32 = 0x7374737a // "stsz"
	fourCCstz2 uint32 = 0x73747a32 // "stz2"
	fourCCstco uint32 = 0x7374636f // "stco"
	fourCCco64 uint32 = 0x636f3634 // "co64"
	fourCCstss uint32 = 0x73747373 // "stss"
	fourCCctts uint32 = 0x63747473 // "ctts"
	fourCCcslg uint32 = 0x63736c67 // "cslg"
	fourCCstsh uint32 = 0x73747368 // "stsh"
	fourCCpadb uint32 = 0x70616462 // "padb"
	fourCCstdp uint32 = 0x73746470 // "stdp"
	fourCCsdtp uint32 = 0x73647470 // "sdtp"
	fourCCcolr uint32 = 0x636f6c72 // "colr"
	fourCCclap uint32 = 0x636c6170 // "clap"
	fourCCpasp uint32 = 0x70617370 // "pasp"

	avc1SampleEntry uint32 = 0x61766331 // "avc1"   video sample entry ->
	avc2SampleEntry uint32 = 0x61766332 // "avc2"
	avc3SampleEntry uint32 = 0x61766333 // "avc3"
	avc4SampleEntry uint32 = 0x61766334 // "avc4"
	encvSampleEntry uint32 = 0x656e6376 // "protectedInfo"  encrypted video sample entry
	hev1SampleEntry uint32 = 0x68657631 // "hev1"
	hvc1SampleEntry uint32 = 0x68766331 // "hvc1"
	hVC1SampleEntry uint32 = 0x48564331 // "HVC1"
	dvavSampleEntry uint32 = 0x64766176 // "dvav"
	dva1SampleEntry uint32 = 0x64766131 // "dva1"
	dvheSampleEntry uint32 = 0x64766865 // "dvhe"
	dvh1SampleEntry uint32 = 0x64766831 // "dvh1"
	vp08SampleEntry uint32 = 0x76703038 // "vp08"
	vp09SampleEntry uint32 = 0x76703039 // "vp09"
	av01SampleEntry uint32 = 0x61763031 // "av01"
	s263SampleEntry uint32 = 0x73323633 // "s263"
	h263SampleEntry uint32 = 0x48323633 // "H263"
	s264SampleEntry uint32 = 0x73323634 // "s264"
	mp4vSampleEntry uint32 = 0x6d703476 // "mp4v"
	jpegSampleEntry uint32 = 0x6a706567 // "jpeg"
	jPEGSampleEntry uint32 = 0x4a504547 // "JPEG"
	div3SampleEntry uint32 = 0x64697633 // "div3"
	dIV3SampleEntry uint32 = 0x44495633 // "DIV3"	<- video sample entry

	fourCCav1c uint32 = 0x61763143 // "av1C"  -> video codec configuration record
	fourCCavcC uint32 = 0x61766343 // "avcC"
	fourCCdvcC uint32 = 0x64766343 // "dvcC"
	fourCCdvvC uint32 = 0x64767643 // "dvvC"
	fourCCvpcC uint32 = 0x76706343 // "vpcC"
	fourCChvcC uint32 = 0x68766343 // "hvcC"  <- video codec configuration record

	flaCSampleEntry uint32 = 0x664c6143 // "fLaC"	audio sample entry ->
	opusSampleEntry uint32 = 0x4f707573 // "Opus"
	mp4aSampleEntry uint32 = 0x6d703461 // "mp4a"
	encaSampleEntry uint32 = 0x656e6361 // "enca"  encrypted audio sample entry
	mp3SampleEntry  uint32 = 0x2e6d7033 // ".mp3"
	lpcmSampleEntry uint32 = 0x6c70636d // "lpcm"
	alacSampleEntry uint32 = 0x616c6163 // "alac"
	ac3SampleEntry  uint32 = 0x61632d33 // "ac-3"
	ac4SampleEntry  uint32 = 0x61632d34 // "ac-4"
	ec3SampleEntry  uint32 = 0x65632d33 // "ec-3"
	mlpaSampleEntry uint32 = 0x6D6C7061 // "mlpa"
	dtscSampleEntry uint32 = 0x64747363 // "dtsc"
	dtseSampleEntry uint32 = 0x64747365 // "dtse"
	dtshSampleEntry uint32 = 0x64747368 // "dtsh"
	dtslSampleEntry uint32 = 0x6474736c // "dtsl"
	samrSampleEntry uint32 = 0x73616d72 // "samr"
	sawbSampleEntry uint32 = 0x73617762 // "sawb"
	sowtSampleEntry uint32 = 0x736f7774 // "sowt"
	twosSampleEntry uint32 = 0x74776f73 // "twos"
	alawSampleEntry uint32 = 0x616c6177 // "alaw"
	ulawSampleEntry uint32 = 0x756c6177 // "ulaw"
	sounSampleEntry uint32 = 0x736f756e // "soun"	<- audio sample entry

	tx3gSampleEntry uint32 = 0x74783367 // "tx3g"	subtitle sample entry ->
	stppSampleEntry uint32 = 0x73747070 // "stpp"
	wvttSampleEntry uint32 = 0x77767474 // "wvgtt"
	TTMLSampleEntry uint32 = 0x54544d4c // "TTML"
	c608SampleEntry uint32 = 0x63363038 // "c608"	<- subtitle sample entry

	fourCCesds uint32 = 0x65736473 // "esds" audio sample descriptors ->
	fourCCdfla uint32 = 0x64664c61 // "dfLa"
	fourCCdops uint32 = 0x644f7073 // "dOps"
	fourCCalac uint32 = 0x616C6163 // "alac" - Also used by ALACSampleEntry
	fourCCddts uint32 = 0x64647473 // "ddts"
	fourCCdac3 uint32 = 0x64616333 // "dac3"
	fourCCdec3 uint32 = 0x64656333 // "dec3"
	fourCCdac4 uint32 = 0x64616334 // "dac4"
	fourCCwave uint32 = 0x77617665 // "wave" - quicktime atom
	fourCCdmlp uint32 = 0x646D6C70 // "dmlp"  <- audio sample descriptors

	// protection information boxes
	fourCCpssh uint32 = 0x70737368 // "pssh"
	fourCCsinf uint32 = 0x73696e66 // "sinf"
	fourCCfrma uint32 = 0x66726d61 // "frma"
	fourCCschm uint32 = 0x7363686d // "schm"
	fourCCschi uint32 = 0x73636869 // "schi"
	// fourCCtenc uint32 = 0x74656e63 // "tenc"

	// fourCCctts uint32 = 0x63747473 // "ctts"
	// fourCCuuid uint32 = 0x75756964 // "uuid"
	// fourCCmhdr uint32 = 0x6d686472 // "mhdr"
	// fourCCkeys uint32 = 0x6b657973 // "keys"
	// fourCCilst uint32 = 0x696c7374 // "ilst"
	// fourCCdata uint32 = 0x64617461 // "Data"
	// fourCCname uint32 = 0x6e616d65 // "name"
	// fourCCitif uint32 = 0x69746966 // "itif"
	// fourCCudta uint32 = 0x75647461 // "udta"

	// AlbumEntry           uint32 = 0xa9616c62 // "©alb"
	// ArtistEntry          uint32 = 0xa9415254 // "©ART"
	// ArtistLowercaseEntry uint32 = 0xa9617274 // "©art"
	// AlbumArtistEntry     uint32 = 0x61415254 // "aART"
	// CommentEntry         uint32 = 0xa9636d74 // "©cmt"
	// DateEntry            uint32 = 0xa9646179 // "©day"
	// TitleEntry           uint32 = 0xa96e616d // "©nam"
	// CustomGenreEntry     uint32 = 0xa967656e // "©gen"
	// StandardGenreEntry   uint32 = 0x676e7265 // "gnre"
	// TrackNumberEntry     uint32 = 0x74726b6e // "trkn"
	// DiskNumberEntry      uint32 = 0x6469736b // "disk"
	// ComposerEntry        uint32 = 0xa9777274 // "©wrt"
	// EncoderEntry         uint32 = 0xa9746f6f // "©too"
	// EncodedByEntry       uint32 = 0xa9656e63 // "©enc"
	// TempoEntry           uint32 = 0x746d706f // "tmpo"
	// CopyrightEntry       uint32 = 0x63707274 // "cprt"
	// CompilationEntry     uint32 = 0x6370696c // "cpil"
	// CoverArtEntry        uint32 = 0x636f7672 // "covr"
	// AdvisoryEntry        uint32 = 0x72746e67 // "rtng"
	// RatingEntry          uint32 = 0x72617465 // "rate"
	// GroupingEntry        uint32 = 0xa9677270 // "©grp"
	// MediaTypeEntry       uint32 = 0x7374696b // "stik"
	// PodcastEntry         uint32 = 0x70637374 // "pcst"
	// CategoryEntry        uint32 = 0x63617467 // "catg"
	// KeywordEntry         uint32 = 0x6b657977 // "keyw"
	// PodcastUrlEntry      uint32 = 0x7075726c // "purl"
	// PodcastGuidEntry     uint32 = 0x65676964 // "egid"
	// DescriptionEntry     uint32 = 0x64657363 // "desc"
	// LongDescriptionEntry uint32 = 0x6c646573 // "ldes"
	// LyricsEntry          uint32 = 0xa96c7972 // "©lyr"
	// TVNetworkNameEntry   uint32 = 0x74766e6e // "tvnn"
	// TVShowNameEntry      uint32 = 0x74767368 // "tvsh"
	// TVEpisodeNameEntry   uint32 = 0x7476656e // "tven"
	// TVSeasonNumberEntry  uint32 = 0x7476736e // "tvsn"
	// TVEpisodeNumberEntry uint32 = 0x74766573 // "tves"
	// PurchaseDateEntry    uint32 = 0x70757264 // "purd"
	// GaplessPlaybackEntry uint32 = 0x70676170 // "pgap"
	// OwnerEntry           uint32 = 0x6f776e72 // "ownr"
	// HDVideoEntry         uint32 = 0x68647664 // "hdvd"
	// SortNameEntry        uint32 = 0x736f6e6d // "sonm"
	// SortAlbumEntry       uint32 = 0x736f616c // "soal"
	// SortArtistEntry      uint32 = 0x736f6172 // "soar"
	// SortAlbumArtistEntry uint32 = 0x736f6161 // "soaa"
	// SortComposerEntry    uint32 = 0x736f636f // "soco"
)

type boxFtyp struct {
	majorBrand        uint32
	minorVersion      uint32
	compatibleBrands  []uint32
	isQuickTimeFormat bool
}

type boxSidx struct {
	referenceID             uint32
	timeScale               uint32
	earlistPresentationTime uint64
	firstTime               uint64
	referenceCount          uint16
	reference               []struct {
		referenceType      uint8  // reference_type 1 bit
		referenceSize      uint32 // reference_size 31 bit
		subSegmentDuration uint32
		startWithSAP       uint8  // starts_with_SAP 1 bit
		sapType            uint8  // SAP_type 3 bit
		sapDeltaTime       uint32 // SAP_delta_time 28 bit
	}
}

func (p *boxSidx) String() string {
	return fmt.Sprintf("\n[Segment Index]:\n{   ReferenceID:%d\n    Time Scale:%d    EarlistPresentationTime:%d\n    "+
		"firstTime:%d\n%s}", p.referenceID, p.timeScale, p.earlistPresentationTime, p.firstTime,
		func() string {
			var retString string
			for i := uint16(0); i < p.referenceCount; i++ {
				retString += fmt.Sprintf("    referenceType:%-2d referenceSize:%-10d subSegmentDuration:%-10d startWithSAP:%-1d sapType:%-2d sapDeltaTime:%-10d\n",
					p.reference[i].referenceType, p.reference[i].referenceSize, p.reference[i].subSegmentDuration, p.reference[i].startWithSAP,
					p.reference[i].sapType, p.reference[i].sapDeltaTime)
			}
			return retString
		}())
}

type boxSsix struct {
	subSegmentCount uint32 // is ranges' len
	ranges          []struct {
		rangeCount uint32 // is rangeSize's len
		rangeSize  []struct {
			level uint8
			size  uint32
		}
	}
}

// type boxMfra struct {
// 	trfa []boxTraf
// 	mfro *boxMfro
// }

type boxMvhd struct {
	version          int
	creationTime     uint64 // uint32 : Version == 0
	modificationTime uint64 // uint32 : Version == 0
	timeScale        uint32
	duration         uint64    // uint32 : Version == 0
	rate             uint32    // 0x00010000
	volume           uint16    // 0x0100
	reserved1        [10]uint8 // bit(16) reserved = 0; int(32)[2] reserved = 0; int(32)[9]
	matrix           [9]uint32 // int(32)[9] matrix = { 0x00010000,0,0,0,0x00010000,0,0,0,0x40000000 };
	reserved2        [24]uint8 // bit(32)[6] pre_defined = 0;
	nextTrackId      uint32
}

type boxMvex struct {
	fragmentDuration uint64 // uint32 if Version == 0
	trex             []boxTrex
	leva             *boxLeva
}

type boxTrex struct {
	trackId                       uint32
	defaultSampleDescriptionIndex uint32
	defaultSampleDuration         uint32
	defaultSampleSize             uint32
	defaultSampleFlags            uint32
}

type boxLeva struct {
	levelCount uint8
	levels     []struct {
		trackId               uint32
		paddingFlag           uint8  // 1 bit
		assignmentType        uint8  // 7bit
		groupingType          uint32 // assignmentType == 0 || 1
		groupingTypeParameter uint32 // assignmentType == 1
		subTrackId            uint32 // assignmentType == 4
	}
}

type boxTrep struct {
	trackId uint32
}

type boxTrak struct {
	packets         []Packet
	id              uint32 // track id
	trackEnabled    bool   // is track enabled
	trackType       TrackType
	quickTimeFormat bool // only for audio

	movie *MovieInfo

	creationTime     uint64 // in seconds since midnight, Jan. 1, 1904, in UTC time
	modificationTime uint64 // in seconds since midnight, Jan. 1, 1904, in UTC time

	// the duration of media. If edit list box exist, the value of this field is equal to
	// the sum of the durations of all the track’s edits.
	duration     uint64
	sampleNumber uint64
	timeOffset   int64

	timeScale   uint32
	language    uint16 // ISO-639-2/T language code
	extLanguage string

	// for visual tracks
	flagTrackSizeIsAspectRatio bool
	width                      uint32
	height                     uint32

	format uint32 // fourCC format, i.e. unencrypted sample entry/ Coding name

	encrypted bool

	protection []*ProtectedInformation

	edts *boxEdts
	// mdia *boxMdia

	audioEntry *audioSampleEntry
	videoEntry *videoSampleEntry

	stts             *boxStts
	ctts             *boxCtts
	cslg             *boxCslg
	stsc             *boxStsc // sample to chunk
	stsz             *boxStsz // sample size
	stco             *boxStco // chunk offset
	syncSamples      []uint32
	stss             *boxStss
	stsh             *boxStsh
	samplePriority   []uint16 // degradation priority of each sample. If existed, len(samplePriority) == sample_count of stsz box
	sampleDependency *boxSdtp

	subs *boxSubs

	sbgp *boxSbgp
	sgpd *boxSgpd

	saio *boxSaio
	saiz *boxSaiz
	senc *boxSenc
}

func (p *boxTrak) getTrackType() TrackType {
	return p.trackType
}

func (p *boxTrak) getTrackId() uint32 {
	return p.id
}

func (p *boxTrak) isTrackEncrypted() bool {
	return len(p.protection) > 0 && p.protection[0].DefaultIsProtected == 1
}

func (p *boxTrak) getTrackDuration() uint64 {
	return p.duration
}

func (p *boxTrak) getTrackTimeScale() uint32 {
	return p.timeScale
}

func (p *boxTrak) getProtectedInformation() *ProtectedInformation {
	if len(p.protection) > 0 {
		return p.protection[0]
	}
	return nil
}

func (p *boxTrak) getProtectedInformationBySchemeType(scheme uint32) *ProtectedInformation {
	for i := 0; i < len(p.protection); i++ {
		if p.protection[i].SchemeType == scheme {
			return p.protection[i]
		}
	}
	return nil
}

type boxMdia struct {
	// media header
	creationTime     uint64 // in seconds since midnight, Jan. 1, 1904, in UTC time
	modificationTime uint64 // in seconds since midnight, Jan. 1, 1904, in UTC time
	timeScale        uint32
	duration         uint64 // in timeScale
	language         uint16 //  unsigned int(5)[3], ISO-639-2/T language code

	stbl           *boxStbl
	ctts           *boxCtts
	extLanguageTag string
}

type PSSH struct {
	SystemId []byte     // uuid, 128 bits (16 bytes)
	KId      [][16]byte // unsigned int(8)[16] KID
	Data     []byte     // len(Data) == DataSize
}

type boxTkhd struct {
	trackId          uint32
	creationTime     uint64 // if Version == 1  else uint32; in seconds since midnight, Jan. 1, 1904, in UTC time
	modificationTime uint64 // if Version == 1  else uint32; in seconds since midnight, Jan. 1, 1904, in UTC time
	duration         uint64
	volume           uint16 // if track_is_audio 0x0100 else 0
	width            uint32
	height           uint32

	flagTrackEnabled   bool
	flagTrackInMovie   bool
	flagTrackInPreview bool
}

type boxMinf struct {
	dinf *boxDinf
	stbl *boxStbl
}

type dataEntry struct {
	entryFlag uint32
	content   string
}

type boxDinf struct {
	entryCount  uint32
	dataEntries map[uint32]*dataEntry
}

type boxStbl struct {
	stsd *boxStsd
	stts *boxStts
	stsc *boxStsc
	stco *boxStco
	stsz *boxStsz
	stss *boxStss
	ctts *boxCtts
	saio []*boxSaio
	saiz []*boxSaiz
	sbgp *boxSbgp
	sgpd *boxSgpd
	subs *boxSubs
}

type boxStsd struct {
	version          uint8
	entryCount       uint32
	audioSampleEntry *audioSampleEntry
	videoSampleEntry *videoSampleEntry
	protectedInfo    *ProtectedInformation
}

type audioSampleEntry struct {
	qttfBytesPerSample   uint32
	qttfSamplesPerPacket uint32
	qttfBytesPerPacket   uint32
	qttfBytesPerFrame    uint32

	quickTimeVersion int
	codec            CodecType
	channelCount     uint16
	sampleRate       uint32
	sampleSize       uint16
	originalFormat   uint32
	protectedInfo    ProtectedInformation
	format           uint32 // need to be specific, now it represent the entryType

	descriptorsRawData map[CodecType][]byte      // raw Data of descriptor
	decoderDescriptors map[CodecType]interface{} // store the descriptor in specific struct
}

type videoSampleEntry struct {
	originalFormat     uint32
	codec              CodecType
	dataReferenceIndex uint16
	width              uint16
	height             uint16
	depth              uint16
	format             uint32 // need to be specific, now it represent the entryType
	// ColourInformationBox, if has
	colourType              uint32
	colorPrimaries          uint16
	transferCharacteristics uint16
	matrixCoefficients      uint16
	fullRangeFlag           bool
	iCCProfile              []byte
	// PixelAspectRatioBox, if has
	hSpacing uint32
	vSpacing uint32
	// CleanApertureBox, if has
	cleanApertureWidthN  uint32
	cleanApertureWidthD  uint32
	cleanApertureHeightN uint32
	cleanApertureHeightD uint32
	horizOffN            uint32
	horizOffD            uint32
	vertOffN             uint32
	vertOffD             uint32

	protectedInfo               *ProtectedInformation     // information of encv
	configurationRecordsRawData map[CodecType][]byte      // raw Data of decoderConfigurationRecord
	decoderConfigurationRecords map[CodecType]interface{} // key: codec type. value: parsed of decoderConfigurationRecord
}

// String return the human-readable format.
func (v *videoSampleEntry) String() string {
	return fmt.Sprintf("\n[Video Track Information]:\n{\n Original Format:%s\n "+
		"RealFormat:%s\n Codec:%s\n Width:%d, Height:%d\n hSpacing:%d, vSpacing:%d\n}\n",
		int2String(v.originalFormat), int2String(v.format), codecString[v.codec],
		v.width, v.height, v.hSpacing, v.vSpacing)
}

type ProtectedInformation struct {
	DataFormat             uint32 // coding name fourcc
	SchemeType             uint32 // 4CC identifying the scheme
	SchemeVersion          uint32 // scheme Version
	TencVersion            uint8  // Version if "tenc"
	DefaultCryptByteBlock  uint8  // 4 bits
	DefaultSkipByteBlock   uint8  // 4 bits
	DefaultIsProtected     uint8  //  least significant bit: 1 byte
	DefaultPerSampleIVSize uint8  //  least  significant bit 1 byte
	DefaultKID             []byte // 16 bytes
	// if DefaultIsProtected == 1 && DefaultPerSampleIVSize == 0 ->
	DefaultConstantIVSize uint8  //  least  significant bit 1 byte
	DefaultConstantIV     []byte // size: DefaultConstantIVSize bytes
}

func (p *ProtectedInformation) String() string {
	return fmt.Sprintf(" SchemeType:%s  is_protected:%d\n", int2String(p.SchemeType), p.DefaultIsProtected)
}

type boxStts struct {
	entryCount  uint32
	sampleCount []uint32
	sampleDelta []uint32
}

type boxStsc struct {
	entryCount             uint32
	firstChunk             []uint32
	samplePerChunk         []uint32
	sampleDescriptionIndex []uint32
}

type boxStsz struct {
	atomType    uint32 // fourCCstsz/forCCstz2
	sampleSize  uint32 // stsz
	fieldSize   uint8  // stz2
	sampleCount uint32
	entrySize   []uint32
}

type boxStco struct {
	entryCount  uint32
	chunkOffset []uint64
}

type boxTraf struct {
	tfhd                *boxtfhd
	subs                *boxSubs
	baseMediaDecodeTime uint64     // Track fragment decode time
	trun                []*boxTrun // 0 or more
	sbgp                *boxSbgp   // 0 or more
	sgpd                *boxSgpd   // 0 or more, with one for each 'sbgp'
	saio                []*boxSaio // 0 or more
	saiz                []*boxSaiz // 0 or more
	senc                *boxSenc
	psshs               []PSSH
}

type boxtfhd struct {
	trackId                uint32
	flags                  uint32
	baseDataOffset         *uint64 // if flags & 0x000001
	sampleDescriptionIndex *uint32 // if flags & 0x000002
	defaultSampleDuration  *uint32 // if flags & 0x000008
	defaultSampleSize      *uint32 // if flags & 0x000010
	defaultSampleFlags     *uint32 // if flags & 0x000020
	defaultBaseIsMoof      bool    // if flags & 0x000001 == 0
}

type trunSample struct {
	sampleDuration              *uint32
	sampleSize                  *uint32
	sampleFlags                 *uint32
	sampleCompositionTimeOffset *int32 // unsigned if version == 0
}

type boxTrun struct {
	sampleCount      uint32
	dataOffset       *uint32
	firstSampleFlags *uint32
	samples          []*trunSample
}

// sample to group
type boxSbgp struct {
	groupingType          uint32
	groupingTypeParameter *uint32 // if version == 1
	entryCount            uint32
	sampleCount           []uint32 // len(sampleCount) == entryCount
	groupDescriptionIndex []uint32 // len(groupDescriptionIndex) == entryCount
}

type cencSampleEncryptionInformationGroupEntry struct {
	cryptByteBlock  uint8
	skipByteBlock   uint8
	isProtected     bool
	perSampleIVSize uint8
	kID             []byte // 16 byte
	constantIV      []byte // if isProtected && perSampleIVSize == 0
}

// SampleGroupDescription
type boxSgpd struct {
	groupingType                  uint32  // only support "seig" currently
	defaultLength                 *uint32 // if version == 1
	defaultSampleDescriptionIndex *uint32 // if version >= 2
	entryCount                    uint32
	descriptionLength             *uint32                                      // if version ==1 && defaultLength == 0
	cencGroupEntries              []*cencSampleEncryptionInformationGroupEntry // len(cencGroupEntries) == entryCount
}

type subSampleEncryption struct {
	bytesOfClearData     uint16
	bytesOfProtectedData uint32
}

type sampleEncryption struct {
	IV             []byte
	subSampleCount uint16
	subSamples     []subSampleEncryption
}

type boxSenc struct {
	flags       uint32
	sampleCount uint32
	samples     []*sampleEncryption
}

// struct of "subs"
type subSampleInfo struct {
	subSampleSize           uint32
	subSamplePriority       uint8
	discardable             uint8
	codecSpecificParameters uint32
}
type subSampleEntry struct {
	sampleDelta    uint32
	subSampleCount uint16
	subSamples     []*subSampleInfo
}
type boxSubs struct {
	flags      uint32
	entryCount uint32
	entries    []*subSampleEntry
}

// boxSaio and boxSaiz support for encrypted sample entry only.
// i.e. entryCount must be 1 and auxInfoTypeParameter must be 0
// ISO/IEC 23001-7:2016(E) 7.1
type boxSaio struct {
	auxInfoType          uint32   // for encrypted track, ommit it
	auxInfoTypeParameter uint32   // for encrypted track, ommit it
	entryCount           uint32   // must be 1
	offset               []uint64 // len(offset) == entryCount
}

type boxSaiz struct {
	auxInfoType           uint32 // for encrypted track, ommit it
	auxInfoTypeParameter  uint32 // for encrypted track, ommit it
	defaultSampleInfoSize uint8
	sampleCount           uint32
	sampleInfoSize        []uint8 // len(sampleInfoSize) == sampleCount if defaultSampleInfoSize == 0
}

type boxEdts struct {
	entryCount   uint32
	editDuration []uint64 // if Version == 0, uint32
	mediaTime    []int64  // if Version == 0, int32
	mediaRate    []float32
}

type boxStss struct {
	entryCount   uint32
	sampleNumber []uint32
}

type boxCtts struct {
	entryCount   uint32
	sampleCount  []uint32
	sampleOffset []int32 // signed if version == 1
}

// CompositionToDecodeBox
type boxCslg struct {
	compositionToDTSShift        int64
	leastDecodeToDisplayDelta    int64
	greatestDecodeToDisplayDelta int64
	compositionStartTime         int64
	compositionEndTime           int64
}

// shadow sync table, for seeking or for similar purposes

type boxStsh struct {
	entryCount           uint32
	shadowedSampleNumber []uint32 // size is entryCount
	syncSampleNumber     []uint32 // size is entryCount
}

type boxSdtp struct {
	// all parameter's length is sample_count in stsz box
	isLeading           []uint8
	sampleDependsOn     []uint8
	sampleIsDependedOn  []uint8
	sampleHasRedundancy []uint8
}

//
// type sphatial struct {
// 	Spherical bool
// 	Stitched bool
// 	StitchingSoftware bool
// }
