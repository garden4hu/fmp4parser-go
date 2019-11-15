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
