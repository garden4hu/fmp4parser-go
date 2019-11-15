package fmp4parser

// ISO/IEC 14496-12 Part 12: ISO base media file format
// basic copy from https://github.com/mozilla/mp4parse-rust/blob/master/mp4parse/src/boxes.rs
var (
	ftypBox uint32 = 0x66747970 // "ftyp"
	stypBox uint32 = 0x73747970 // "styp"
	moovBox uint32 = 0x6d6f6f76 // "moov"
	sidxBox uint32 = 0x73696478 // "sidx"
	ssixBox uint32 = 0x73736978 // "ssix"
	mdatBox uint32 = 0x6D646174 // "mdat"
	metaBox uint32 = 0x6d657461 // "meta"
	ilocBox uint32 = 0x696C6F63 // "iloc"
	trexBox uint32 = 0x74726578 // "trex"
	moofBox uint32 = 0x6D6F6F66 // "moof" 	fragment-dash box ->
	mfhdBox uint32 = 0x6D666864 // "mfhd"
	trafBox uint32 = 0x74726166 // "traf"
	tfhdBox uint32 = 0x74666864 // "tfhd"
	trunBox uint32 = 0x7472756E // "trun"
	sbgpBox uint32 = 0x73626770 // "sbgp"
	sgpdBox uint32 = 0x73677064 // "sgpd"
	subsBox uint32 = 0x73756273 // "subs"
	saizBox uint32 = 0x7361697A // "saiz"
	saioBox uint32 = 0x7361696F // "saio"
	tfdtBox uint32 = 0x74666474 // "tfdt"  <- fragment-dash box
	mfraBox uint32 = 0x6D667261 // "mfra"
	mvhdBox uint32 = 0x6d766864 // "mvhd"
	trakBox uint32 = 0x7472616b // "trak"
	tkhdBox uint32 = 0x746b6864 // "tkhd"
	edtsBox uint32 = 0x65647473 // "edts"
	mdiaBox uint32 = 0x6d646961 // "mdia"
	elstBox uint32 = 0x656c7374 // "elst"
	mdhdBox uint32 = 0x6d646864 // "mdhd"
	hdlrBox uint32 = 0x68646c72 // "hdlr"
	minfBox uint32 = 0x6d696e66 // "minf"
	vmhdBox uint32 = 0x766D6864 // "vmhd"
	smhdBox uint32 = 0x736D6864 // "smhd"
	hmhdBox uint32 = 0x686D6864 // "hmhd"
	nmhdBox uint32 = 0x6E6D6864 // "nmhd"
	dinfBox uint32 = 0x64696E66 // "dinf"
	drefBox uint32 = 0x64726566 // "dref"
	stblBox uint32 = 0x7374626c // "stbl"
	stsdBox uint32 = 0x73747364 // "stsd"
	sttsBox uint32 = 0x73747473 // "stts"
	stscBox uint32 = 0x73747363 // "stsc"
	stszBox uint32 = 0x7374737a // "stsz"
	stcoBox uint32 = 0x7374636f // "stco"
	co64Box uint32 = 0x636f3634 // "co64"
	stssBox uint32 = 0x73747373 // "stss"

	avc1SampleEntry uint32 = 0x61766331 // "avc1"   video sample entry ->
	avc2SampleEntry uint32 = 0x61766332 // "avc2"
	avc3SampleEntry uint32 = 0x61766333 // "avc3"
	avc4SampleEntry uint32 = 0x61766334 // "avc4"
	encvSampleEntry uint32 = 0x656e6376 // "encv"
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

	av1cConfigurationBox uint32 = 0x61763143 // "av1C"
	avcconfigurationBox  uint32 = 0x61766343 // "avcC"
	vpccConfigurationBox uint32 = 0x76706343 // "vpcC"

	flaCSampleEntry uint32 = 0x664c6143 // "fLaC"	audio sample entry ->
	opusSampleEntry uint32 = 0x4f707573 // "Opus"
	mp4aSampleEntry uint32 = 0x6d703461 // "mp4a"
	encaSampleEntry uint32 = 0x656e6361 // "enca"
	mp3SampleEntry  uint32 = 0x2e6d7033 // ".mp3"
	lpcmSampleEntry uint32 = 0x6c70636d // "lpcm"
	alacSampleEntry uint32 = 0x616c6163 // "alac"
	ac_3SampleEntry uint32 = 0x61632d33 // "ac-3"
	ac_4SampleEntry uint32 = 0x61632d34 // "ac-4"
	ec_3SampleEntry uint32 = 0x65632d33 // "ec-3"
	dtscSampleEntry uint32 = 0x64747363 // "dtsc"
	dtseSampleEntry uint32 = 0x64747365 // "dtse"
	dtshSampleEntry uint32 = 0x64747368 // "dtsh"
	dtslSampleEntry uint32 = 0x6474736c // "dtsl"
	samrSampleEntry uint32 = 0x73616d72 // "samr"
	sawbSampleEntry uint32 = 0x73617762 // "sawb"
	sowtSampleEntry uint32 = 0x736f7774 // "sowt"
	alawSampleEntry uint32 = 0x616c6177 // "alaw"
	ulawSampleEntry uint32 = 0x756c6177 // "ulaw"
	sounSampleEntry uint32 = 0x736f756e // "soun"	<- audio sample entry

	tx3gSampleEntry uint32 = 0x74783367 // "tx3g"	subtitle sample entry ->
	stppSampleEntry uint32 = 0x73747070 // "stpp"
	wvttSampleEntry uint32 = 0x77767474 // "wvgtt"
	TTMLSampleEntry uint32 = 0x54544d4c // "TTML"
	c608SampleEntry uint32 = 0x63363038 // "c608"	<- subtitle sample entry

	esdsBox uint32 = 0x65736473 // "esds"
	dflaBox uint32 = 0x64664c61 // "dfLa"
	dopsBox uint32 = 0x644f7073 // "dOps"
	mvexBox uint32 = 0x6d766578 // "mvex"
	mehdBox uint32 = 0x6d656864 // "mehd"
	waveBox uint32 = 0x77617665 // "wave" - quicktime atom

	sinfBox uint32 = 0x73696e66 // "sinf"
	frmaBox uint32 = 0x66726d61 // "frma"
	schmBox uint32 = 0x7363686d // "schm"
	psshBox uint32 = 0x70737368 // "pssh"
	schiBox uint32 = 0x73636869 // "schi"
	tencBox uint32 = 0x74656e63 // "tenc"

	cttsBox   uint32 = 0x63747473 // "ctts"
	alacBox   uint32 = 0x616C6163 // "alac" - Also used by ALACSampleEntry
	uuidBox   uint32 = 0x75756964 // "uuid"
	mhdrBox   uint32 = 0x6d686472 // "mhdr"
	keysBox   uint32 = 0x6b657973 // "keys"
	ilstEntry uint32 = 0x696c7374 // "ilst"
	dataEntry uint32 = 0x64617461 // "data"
	nameBox   uint32 = 0x6e616d65 // "name"
	itifBox   uint32 = 0x69746966 // "itif"
	udtaBox   uint32 = 0x75647461 // "udta"

	AlbumEntry           uint32 = 0xa9616c62 // "©alb"
	ArtistEntry          uint32 = 0xa9415254 // "©ART"
	ArtistLowercaseEntry uint32 = 0xa9617274 // "©art"
	AlbumArtistEntry     uint32 = 0x61415254 // "aART"
	CommentEntry         uint32 = 0xa9636d74 // "©cmt"
	DateEntry            uint32 = 0xa9646179 // "©day"
	TitleEntry           uint32 = 0xa96e616d // "©nam"
	CustomGenreEntry     uint32 = 0xa967656e // "©gen"
	StandardGenreEntry   uint32 = 0x676e7265 // "gnre"
	TrackNumberEntry     uint32 = 0x74726b6e // "trkn"
	DiskNumberEntry      uint32 = 0x6469736b // "disk"
	ComposerEntry        uint32 = 0xa9777274 // "©wrt"
	EncoderEntry         uint32 = 0xa9746f6f // "©too"
	EncodedByEntry       uint32 = 0xa9656e63 // "©enc"
	TempoEntry           uint32 = 0x746d706f // "tmpo"
	CopyrightEntry       uint32 = 0x63707274 // "cprt"
	CompilationEntry     uint32 = 0x6370696c // "cpil"
	CoverArtEntry        uint32 = 0x636f7672 // "covr"
	AdvisoryEntry        uint32 = 0x72746e67 // "rtng"
	RatingEntry          uint32 = 0x72617465 // "rate"
	GroupingEntry        uint32 = 0xa9677270 // "©grp"
	MediaTypeEntry       uint32 = 0x7374696b // "stik"
	PodcastEntry         uint32 = 0x70637374 // "pcst"
	CategoryEntry        uint32 = 0x63617467 // "catg"
	KeywordEntry         uint32 = 0x6b657977 // "keyw"
	PodcastUrlEntry      uint32 = 0x7075726c // "purl"
	PodcastGuidEntry     uint32 = 0x65676964 // "egid"
	DescriptionEntry     uint32 = 0x64657363 // "desc"
	LongDescriptionEntry uint32 = 0x6c646573 // "ldes"
	LyricsEntry          uint32 = 0xa96c7972 // "©lyr"
	TVNetworkNameEntry   uint32 = 0x74766e6e // "tvnn"
	TVShowNameEntry      uint32 = 0x74767368 // "tvsh"
	TVEpisodeNameEntry   uint32 = 0x7476656e // "tven"
	TVSeasonNumberEntry  uint32 = 0x7476736e // "tvsn"
	TVEpisodeNumberEntry uint32 = 0x74766573 // "tves"
	PurchaseDateEntry    uint32 = 0x70757264 // "purd"
	GaplessPlaybackEntry uint32 = 0x70676170 // "pgap"
	OwnerEntry           uint32 = 0x6f776e72 // "ownr"
	HDVideoEntry         uint32 = 0x68647664 // "hdvd"
	SortNameEntry        uint32 = 0x736f6e6d // "sonm"
	SortAlbumEntry       uint32 = 0x736f616c // "soal"
	SortArtistEntry      uint32 = 0x736f6172 // "soar"
	SortAlbumArtistEntry uint32 = 0x736f6161 // "soaa"
	SortComposerEntry    uint32 = 0x736f636f // "soco"
)

const (
	audioTrack = iota
	videoTrak
	subtitleTrack
	unknowTrack
)

type boxFtyp struct {
	majorBrand       uint32
	minorVersion     uint32
	compatibleBrands []uint32
}

type boxStyp struct {
}

type boxMoov struct {
	mvhd   *boxMvhd
	Mvex   *boxMvex
	tracks []boxTrak
	Meta   []boxMeta
}

type boxMeta struct {
	hldr boxHdlr
	iloc boxIloc
}

type boxSidx struct {
}

type boxSsix struct {
}
type boxMoof struct {
}

type boxMfra struct {
}

type boxMvhd struct {
	version          int
	creationTime     uint64 // uint32 : version == 0
	modificationTime uint64 // uint32 : version == 0
	timescale        uint32
	duration         uint64    // uint32 : version == 0
	rate             uint32    // 0x00010000
	volume           uint16    // 0x0100
	reserved1        [10]uint8 // bit(16) reserved = 0; int(32)[2] reserved = 0; int(32)[9]
	matrix           [9]uint32 // int(32)[9] matrix = { 0x00010000,0,0,0,0x00010000,0,0,0,0x40000000 };
	reserved2        [24]uint8 // bit(32)[6] pre_defined = 0;
	nextTrackId      uint32
}

type boxMvex struct {
	mehdbox boxMehd
	trexbox []boxTrex
	pssh    []boxPssh
}

type boxMehd struct {
	fragmentDuration uint64
}

type boxTrex struct {
	trackId                       uint32
	defaultSampleDescriptionIndex uint32
	defaultSampleDuration         uint32
	defaultSampleSize             uint32
	defaultSampleFlags            uint32
}

type boxTrak struct {
	id               uint32
	creationTime     uint64 // in seconds since midnight, Jan. 1, 1904, in UTC time
	modificationTime uint64 // in seconds since midnight, Jan. 1, 1904, in UTC time
	duration         uint64

	tkhd *boxTkhd
	mdia *boxMdia
}

type boxMdia struct {
	mdhd *boxMdhd
	hldr *boxHdlr
	minf *boxMinf
}

type boxMdhd struct {
	creationTime     uint64 // in seconds since midnight, Jan. 1, 1904, in UTC time
	modificationTime uint64 // in seconds since midnight, Jan. 1, 1904, in UTC time
	timeScale        uint32
	duration         uint64 // in timescale
	language         []byte //  unsigned int(5)[3], ISO-639-2/T language code
}

type boxHdlr struct {
	handlerType uint32
	name        string // a null-terminated string in UTF-8 characters
}

type boxIloc struct {
}

type boxPssh struct {
	version int
	// if version > 0
	systemId []byte   // uuid, 128 bits (16 bytes)
	kIdCount uint32   // number of kId
	kId      [][]byte // unsigned int(8)[16] KID

	dataSize uint32
	data     []byte // len(data) == dataSize
}

type boxTkhd struct {
	trackId          uint32
	creationTime     uint64 // if version == 1  esle uint32
	modificationTime uint64
	duration         uint64
	volume           int // in fact is 2 byte; if track_is_audio 0x0100 else 0
	width            uint32
	hight            uint32
}

type boxMinf struct {
	dinf *boxDinf
	stbl *boxStbl
}

type boxDinf struct {
}

type boxStbl struct {
	stsd *boxStsd
	stts *boxStts
	stsc *boxStsc
	stco *boxStco
	stsz *boxStsz
}

type boxStsd struct {
	version            int
	audioSampleEntries []audioSampleEntry
}

type audioSampleEntry struct {
	qttf	bool
	qttfVersion                   int // quick time format version
	qttfBytesPerSample     int
	codecId             uint32 // FourCC
	channelCount           int
	sampleRate             int
	sampleSize             int
	qttfSamplesPerPacket int
	qttfBytesPerPacket   int
	qttfBytesPerFrame    int
	qttfBytesPerSample   int
}

type boxEnca struct {
	sinf *boxSinf
	
}

type boxSinf struct {

}

type boxStts struct {
}

type boxStsc struct {
}

type boxStsz struct {
}

type boxStco struct {
}

