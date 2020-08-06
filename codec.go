package fmp4parser

func getTrackType(box uint32) TrackType {
	if box == avc1SampleEntry ||
		box == avc2SampleEntry ||
		box == avc3SampleEntry ||
		box == avc4SampleEntry ||
		box == encvSampleEntry ||
		box == hev1SampleEntry ||
		box == hvc1SampleEntry ||
		box == hVC1SampleEntry ||
		box == dvavSampleEntry ||
		box == dva1SampleEntry ||
		box == dvheSampleEntry ||
		box == dvh1SampleEntry ||
		box == vp08SampleEntry ||
		box == vp09SampleEntry ||
		box == av01SampleEntry ||
		box == s263SampleEntry ||
		box == h263SampleEntry ||
		box == s264SampleEntry ||
		box == mp4vSampleEntry ||
		box == jpegSampleEntry ||
		box == jPEGSampleEntry ||
		box == div3SampleEntry ||
		box == dIV3SampleEntry {
		return VideoTrak
	} else if box == flaCSampleEntry ||
		box == opusSampleEntry ||
		box == mp4aSampleEntry ||
		box == encaSampleEntry ||
		box == mp3SampleEntry ||
		box == lpcmSampleEntry ||
		box == alacSampleEntry ||
		box == ac3SampleEntry ||
		box == ac4SampleEntry ||
		box == ec3SampleEntry ||
		box == mlpaSampleEntry ||
		box == dtscSampleEntry ||
		box == dtseSampleEntry ||
		box == dtshSampleEntry ||
		box == dtslSampleEntry ||
		box == samrSampleEntry ||
		box == sawbSampleEntry ||
		box == sowtSampleEntry ||
		box == alawSampleEntry ||
		box == ulawSampleEntry ||
		box == sounSampleEntry {
		return AudioTrack
	} else if box == tx3gSampleEntry ||
		box == stppSampleEntry ||
		box == wvttSampleEntry ||
		box == TTMLSampleEntry ||
		box == c608SampleEntry {
		return SubtitleTrack
	} else {
		return UnknowTrack
	}
}

type lpcmCodecId int

const (
	None lpcmCodecId = iota
	pcmS16LE
	pcmS16BE
	pcmU16LE
	pcmU16BE
	pcmS8
	pcmU8
	pcmMULaw
	pcmALaw
	pcmS32LE
	pcmS32BE
	pcmU32LE
	pcmU32BE
	pcmS24LE
	pcmS24BE
	pcmU24LE
	pcmU24BE
	pcmS24DAUD
	pcmZORK
	pcmS16LEPlanar
	// pcmDVD
	pcmF32BE
	pcmF32LE
	pcmF64BE
	pcmF64LE
	// pcmBluray
	// pcmLXF
	pcmS8Planar
	pcmS24LEPlanar
	pcmS32LEPlanar
	pcmS16BEPlanar
	pcmS64LE
	pcmS64BE
	pcmF16LE
	pcmF24LE
	pcmVIDC
)

// codecId

type CodecType uint32

const (
	CodecUNKNOW CodecType = iota
	VideoCodecH264
	VideoCodecHEVC
	VideoCodecVP8
	VideoCodecVP9
	VideoCodecMP4V
	VideoCodecMPEG
	VideoCodecMPEG2
	VideoCodecVC1
	VideoCodecAV1
	VideoCodecDolbyVision
	VideoCodecMJPEG
	VideoCodecPNG
	VideoCodecJPG2000
	VideoCodecDIRAC

	AudioCodecAAC CodecType = iota + 20
	AudioCodecMP3
	AudioCodecRAW
	AudioCodecALAW
	AudioCodecMULAW
	AudioCodecAC3
	AudioCodecEAC3
	AudioCodecAC4
	AudioCodecMLP // Dolby TrueHD
	AudioCodecDTS
	AudioCodecDTSHD
	AudioCodecDTSEXPRESS
	AudioCodecOPUS
	AudioCodecAMRNB
	AudioCodecAMRWB
	AudioCodecFLAC
	AudioCodecALAC

	// subtitleCodecVTT
	// subtitleCodecSSA
)

// human-readable codec
var codecString map[CodecType]string = map[CodecType]string{
	CodecUNKNOW:           "unknown codec",
	VideoCodecH264:        "h264",
	VideoCodecHEVC:        "hevc",
	VideoCodecVP8:         "vp8",
	VideoCodecVP9:         "vp9",
	VideoCodecMP4V:        "mp4v",
	VideoCodecMPEG:        "mpeg",
	VideoCodecMPEG2:       "mpeg2",
	VideoCodecVC1:         "vc1",
	VideoCodecAV1:         "av1",
	VideoCodecDolbyVision: "dolby vision",
	VideoCodecMJPEG:       "mjpeg",
	VideoCodecPNG:         "png",
	VideoCodecJPG2000:     "jpg2000",
	VideoCodecDIRAC:       "dirac",

	AudioCodecAAC:        "aac",
	AudioCodecMP3:        "mp3",
	AudioCodecRAW:        "raw",
	AudioCodecALAW:       "a-law",
	AudioCodecMULAW:      "m-law",
	AudioCodecAC3:        "ac-3",
	AudioCodecEAC3:       "e-ac-3",
	AudioCodecAC4:        "ac-4",
	AudioCodecMLP:        "dolby mlp",
	AudioCodecDTS:        "dts",
	AudioCodecDTSHD:      "dts-hd",
	AudioCodecDTSEXPRESS: "dts-express",
	AudioCodecOPUS:       "opus",
	AudioCodecAMRNB:      "amr-nb",
	AudioCodecAMRWB:      "amr-wb",
	AudioCodecFLAC:       "flac",
	AudioCodecALAC:       "alac",
}

func getMediaTypeFromObjectType(objectType uint8) CodecType {
	switch objectType {
	case 0x20:
		return VideoCodecMP4V
	case 0x21:
		return VideoCodecH264
	case 0x23:
		return VideoCodecHEVC
	case 0x60:
		fallthrough
	case 0x61:
		fallthrough
	case 0x62:
		fallthrough
	case 0x63:
		fallthrough
	case 0x64:
		fallthrough
	case 0x65:
		return VideoCodecMPEG2
	case 0x6A:
		return VideoCodecMPEG
	case 0x69:
		fallthrough
	case 0x6B:
		return AudioCodecMP3
	case 0xB1:
		return VideoCodecVP9
	case 0x40:
		fallthrough
	case 0x41:
		fallthrough
	case 0x66:
		fallthrough
	case 0x67:
		fallthrough
	case 0x68:
		return AudioCodecAAC
	case 0xA3:
		return VideoCodecVC1
	case 0xA4:
		return VideoCodecDIRAC
	case 0xA5:
		return AudioCodecAC3
	case 0xA6:
		return AudioCodecEAC3
	case 0xA9:
		fallthrough
	case 0xAC:
		return AudioCodecDTS
	case 0xAA:
		fallthrough
	case 0xAB:
		return AudioCodecDTSHD
	case 0xAD:
		return AudioCodecOPUS
	case 0x6C:
		return VideoCodecMJPEG
	case 0x6D:
		return VideoCodecPNG
	case 0x6E:
		return VideoCodecJPG2000
	default:
		return CodecUNKNOW
	}
}
