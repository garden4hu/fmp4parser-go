package fmp4parser

import "fmt"

// MovieInfo is more like a general structure.
// It may be represent a 'moov' or a 'moof', but not both.
// It's decided by the topLevelType.
// Refer to : ISO/IEC 14496-12 Table 1
type MovieInfo struct {
	// For 'moov'
	ftyp *boxFtyp
	ssix []*boxSsix // 0 or more
	sidx []*boxSidx // 0 or more
	mvhd *boxMvhd
	trak []*boxTrak //  1 or more
	pssh []*Pssh    // 0 or more
	mvex *boxMvex

	// For 'moof'
	movieHeader    *MovieInfo // The pointer of parsed moov
	sequenceNumber uint32
	traf           []*boxTraf // 0 or more

	// For 'meta'
	hdlr *boxHdlr

	hasFragment bool // if there is mvex box in moov box, it shows that there are fragment boxes in this file
	// Representation of the type of this MovieInfo. 'moov'/'moof'
	topLevelType uint32
}

func (p *MovieInfo) String() string {
	if p.topLevelType == fourCCmoov {
		return fmt.Sprintf("[MovieInfo]\n TopLevelType:%s trak number:%d",
			int2String(p.topLevelType), len(p.trak))
	} else if p.topLevelType == fourCCmoof {
		return fmt.Sprintf("[MovieInfo]\n TopLevelType:%s traf number:%d",
			int2String(p.topLevelType), len(p.traf))
	}
	return "nil"
}

/* ------------- Audio Codec Specific Boxes------------- */

// ElementaryStreamDescriptor
type EsDescriptor struct {
	AudioCodec              CodecType
	AudioObjectType         int
	ExtendedAudioObjectType int
	SampleRate              uint32
	ChannelCount            uint16
	DecoderSpecificInfo     []byte
}

// OpusDescriptor
type OpusDescriptor struct {
	DecoderSpecificInfo []byte
}

// FlaCDescriptor
type FlacDescriptor struct {
	SampleRate          int
	ChannelCount        int
	BitPerSample        int
	StreamInfo          []byte // the first FLACMetadataBlock, size is 34
	DecoderSpecificInfo []byte
}

// Dolby AC-3/E-AC-3 Descriptor
type Ac3Descriptor struct {
	Fscod        uint8 // sampling frequency code
	Bsid         uint8 // Bit Stream Information
	Acmod        uint8 // audio coding mode
	SampleRate   uint32
	ChannelCount uint16
}

type Ac4Descriptor struct {
	SampleRate          uint32
	DecoderSpecificInfo []byte
}

// DTS* Descriptor
type DtsDescriptor struct {
	SamplingRate        uint32 // DTSSampling Frequency
	MaxBiterate         uint32
	AvgBiterate         uint32
	PcmSampleDepth      uint8  // value is 16 or 24 bits
	FrameDuration       uint16 // 0 = 512, 1 = 1024, 2 = 2048, 3 = 4096
	StreamConstruction  uint8  //
	CoreLFEPresent      uint8  // 0 = none; 1 = LFE exists
	CoreLayout          uint8  //
	CoreSize            uint16
	StereoDownmix       uint8 // 0 = none; 1 = embedded downmix present
	RepresentationType  uint8
	ChannelLayout       uint16
	MultiAssetFlag      uint8 // 0 = single asset, 1 = multiple asset
	LBRDurationMod      uint8 // 0 = ignore, 1 = Special LBR duration modifier
	ReservedBoxPresent  uint8 // 0 = no ReservedBox, 1 = Reserved present
	DecoderSpecificInfo []byte
}

// Dolby TrueHD MlpaDescriptor
type MlpaDescriptor struct {
	FormatInfo          uint32
	PeakDataRate        uint16
	DecoderSpecificInfo []byte
}

/* ------------- Video Codec Specific Boxes------------- */

type AvcConfig struct {
	Version              uint8
	ProfileIndication    uint8
	ProfileCompatibility uint8
	AvcLevel             uint8
	LengthSize           uint8
	NumSPS               uint8
	ListSPS              [][]byte
	NumPPS               uint8
	ListPPS              [][]byte
	DecoderSpecificInfo  []byte // need by decoder
}

type NalUnitInfo struct {
	ArrayCompleteness uint8 // 1 bit lsb
	NALUnitType       uint8 // 6 bits lsb
	NumNalus          uint16
	NalUnitLength     []uint16
	NalUnit           [][]byte // length(nalUint[i]) == NalUnitLength
}

type HevcConfig struct {
	GeneralProfileSpace              uint8 // 2 bits lsb
	GeneralTierFlag                  uint8 // 1 bit lsb
	GeneralProfileIdc                uint8 // 5 bits lsb
	GeneralProfileCompatibilityFlags uint32
	GeneralConstraintIndicatorFlags  uint64 // 48 bit lsb
	GeneralLevelIdc                  uint8
	MinSpatialSegmentationIdc        uint16 // 12 bits lsb
	ParallelismType                  uint8  // 2 bits lsb
	ChromaFormatIdc                  uint8  // 2 bits lsb
	BitDepthLumaMinus8               uint8  // 3 bits lsb
	BitDepthChromaMinus8             uint8  // 3 bits lsb
	AvgFrameRate                     uint16
	ConstantFrameRate                uint8
	NumTemporalLayers                uint8
	TemporalIdNested                 uint8
	LengthSizeMinusOne               uint8
	NumOfArrays                      uint8
	NalUnitArrays                    []NalUnitInfo
	DecoderSpecificInfo              []byte // need by decoder
}

type Av1cConfig struct {
	SeqProfile                       uint8  // 3 bits lsb
	SeqLevelIdx0                     uint8  // 5 bits lsb
	SeqTier0                         uint8  // 1 bit lsb
	HighBitdepth                     uint8  // 1 bit lsb
	TwelveBit                        uint8  // 1 bit lsb
	Monochrome                       uint8  // 1 bit lsb
	ChromaSubsamplingX               uint8  // 1 bit lsb
	ChromaSubsamplingY               uint8  // 1 bit lsb
	ChromaSamplePosition             uint8  // 2 bits lsb
	InitialPresentationDelayPresent  uint8  // 1 bit lsb
	InitialPresentationDelayMinusOne uint8  // 4bits lsb
	DecoderSpecificInfo              []byte // need by decoder
}

type VpcConfig struct {
	Profile                    uint8
	Level                      uint8
	BitDepth                   uint8 // 4 bits lsb
	ChromaSubsampling          uint8 // 3 bits lsb
	VideoFullRangeFlag         uint8 // 1 bit lsb
	ColourPrimaries            uint8
	TransferCharacteristics    uint8
	MatrixCoefficients         uint8
	CodecIntializationDataSize uint16
	CodecIntializationData     []byte
	DecoderSpecificInfo        []byte // need by decoder
}

type DvcConfig struct {
	DvVersionMajor            uint8
	DvVersionMinor            uint8
	DvProfile                 uint8  // 7 bits lsb
	DvLevel                   uint8  // 6 bits lsb
	RpuPresentFlag            uint8  // 1 bit lsb
	ElPresentFlag             uint8  // 1 bit lsb
	BlPresentFlag             uint8  // 1 bit lsb
	DvBlSingalCompatibilityId uint8  // 4 bits lsb
	DecoderSpecificInfo       []byte // need by decoder
}
