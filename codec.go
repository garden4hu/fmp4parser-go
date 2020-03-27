package fmp4parser

func getTrackType(box uint32) int {
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
		return videoTrak
	} else if box == flaCSampleEntry ||
		box == opusSampleEntry ||
		box == mp4aSampleEntry ||
		box == encaSampleEntry ||
		box == mp3SampleEntry ||
		box == lpcmSampleEntry ||
		box == alacSampleEntry ||
		box == ac_3SampleEntry ||
		box == ac_4SampleEntry ||
		box == ec_3SampleEntry ||
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
		return audioTrack
	} else if box == tx3gSampleEntry ||
		box == stppSampleEntry ||
		box == wvttSampleEntry ||
		box == TTMLSampleEntry ||
		box == c608SampleEntry {
		return subtitleTrack
	} else {
		return unknowTrack
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
	pcmDVD
	pcmF32BE
	pcmF32LE
	pcmF64BE
	pcmF64LE
	pcmBluray
	pcmLXF
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

var set = map[string]void{}

// codecId
const (
	videoUNKNOW int = iota
	videoMP4
	videoWEBM
	videoH263
	videoH264
	videoH265
	videoVP8
	videoVP9
	videoMP4V
	videoMPEG
	videoMPEG2
	videoVC1

	audioUNKNOW
	audioMP4
	audioAAC
	audioWEBM
	audioMPEG
	audioMPEGL1
	audioMPEGL2
	audioRAW
	audioALAW
	audioMLAW
	audioAC3
	audioEAC3
	audioEAC3JOC
	audioTRUEHD
	audioDTS
	audioDTSHD
	audioDTSEXPRESS
	audioVORBIS
	audioOPUS
	audioAMRNB
	audioAMRWB
	audioFLAC
	audioALAC
	audioMSGSM

	subtitleVTT
	subtitleSSA
)

func getMediaTypeFromObjectType(objectType int) int {
	switch objectType {
	case 0x20:
		return videoMP4V
	case 0x21:
		return videoH264
	case 0x23:
		return videoH265
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
		return videoMPEG2
	case 0x6A:
		return videoMPEG
	case 0x69:
		fallthrough
	case 0x6B:
		return audioMPEG
	case 0xA3:
		return videoVC1
	case 0xB1:
		return videoVP9
	case 0x40:
		fallthrough
	case 0x41:
		fallthrough
	case 0x66:
		fallthrough
	case 0x67:
		fallthrough
	case 0x68:
		return audioAAC
	case 0xA5:
		return audioAC3
	case 0xA6:
		return audioEAC3
	case 0xA9:
		fallthrough
	case 0xAC:
		return audioDTS
	case 0xAA:
		fallthrough
	case 0xAB:
		return audioDTSHD
	case 0xAD:
		return audioOPUS
	default:
		return int(0xFFFF)
	}
}
