package main

import (
	"errors"
	"fmt"
)

/*
	esds(Elementary Stream Descriptor) refer to:

https://developer.apple.com/library/archive/documentation/QuickTime/QTFF/QTFFChap3/qtff3.html#//apple_ref/doc/uid/TP40000939-CH205-124774
*/
func (p *EsDescriptor) parseDescriptor(r *atomReader) error {
	_ = r.Move(4) // Version(8 bits) + flags(24 bits)
	for i := 0; i < 3; i++ {
		_ = p.findDescriptor(r)
	}
	return nil
}

func (p *EsDescriptor) findDescriptor(r *atomReader) error {
	// defined by 14496-1, section:7.2.2.1
	var esdescrTag uint8 = 0x03
	var decoderConfigTag uint8 = 0x04
	var decoderSpecificTag uint8 = 0x05
	tag := r.ReadUnsignedByte() // tag's name
	// get the esds' length
	currentByte := r.ReadUnsignedByte()
	size := int(currentByte & 0x7f)
	for currentByte&0x80 == 0x80 {
		currentByte = r.ReadUnsignedByte()
		size = size<<7 | int(currentByte&0x7f)
	}
	if uint64(size) > 1<<30 {
		return errors.New("when get esds descriptor, the size is invalid")
	}
	// Start of the ES_Descriptor (defined in 14496-1)
	if tag == esdescrTag {
		_ = r.Move(2) // ES_ID
		flags := r.ReadUnsignedByte()
		if flags&0x80 != 0 { // streamDependenceFlag
			_ = r.Move(2)
		}
		if flags&0x40 != 0 { // uURL_Flag
			_ = r.Move(int(r.Read2()))
		}
		if flags&0x20 != 0 { // OCRstreamFlag
			_ = r.Move(2)
		}
	}
	// Start of the DecoderConfigDescriptor (defined in 14496-1)
	if tag == decoderConfigTag {
		objectProfile := r.ReadUnsignedByte()
		p.AudioCodec = getMediaTypeFromObjectType(objectProfile)
		_ = r.Move(12)
	}
	// Start of the DecoderSpecificInfo
	if tag == decoderSpecificTag {
		p.DecoderSpecificInfo = make([]byte, size)
		_, _ = r.ReadBytes(p.DecoderSpecificInfo)
		// For AAC
		frequencyTable := [13]uint32{96000, 88200, 64000, 48000, 44100, 32000, 24000, 22050, 16000, 12000, 11025, 8000, 7350}
		_ = r.Move(-size)
		b := make([]byte, size)
		_, _ = r.ReadBytes(b)
		br := newBitReaderFromSlice(b)
		getAudioObjectType := func() int {
			audioObjectType := br.ReadBitsLE32(5)
			if audioObjectType == 31 {
				// https://github.com/FFmpeg/FFmpeg/blob/a0ac49e38ee1d1011c394d7be67d0f08b2281526/libavcodec/mpeg4audio.h#L102
				audioObjectTypeExt := br.ReadBitsLE32(6)
				audioObjectType = 32 + audioObjectTypeExt
			}
			return int(audioObjectType)
		}
		audioObjectType := getAudioObjectType()
		frequencyIndex := br.ReadBitsLE32(4)
		frequency := func() uint32 {
			if frequencyIndex == 0x0F {
				return br.ReadBitsLE32(24)
			} else {
				if frequencyIndex < 13 {
					return frequencyTable[frequencyIndex]
				}
				return 0
			}
		}()
		channelConfiguration := br.ReadBitsLE32(4)
		extendedAudioObjectType := 0
		// audio object types: https://github.com/FFmpeg/FFmpeg/blob/b559a5882f54f6ab46b2d058b8526fae6b00ad0f/libavcodec/mpeg4audio.h#L87
		// 5 : Spectral Band Replication ;  29 : Parametric Stereo
		if audioObjectType == 5 || audioObjectType == 29 {
			extendedAudioObjectType = 5
			extendedFrequencyIndex := br.ReadBitsLE32(4)
			extendedFrequency := func() uint32 {
				if extendedFrequencyIndex == 0x0F { // FREQUENCY_INDEX_ARBITRARY
					return br.ReadBitsLE32(24)
				} else {
					if extendedFrequencyIndex < 13 {
						return frequencyTable[extendedFrequencyIndex]
					}
					return 0
				}
			}()
			frequency = extendedFrequency // Use the extendedFrequency.
			audioObjectType := getAudioObjectType()
			extendedChannelConfiguration := uint32(0)
			if audioObjectType == 22 {
				extendedChannelConfiguration = br.ReadBitsLE32(4)
			} else {
				extendedChannelConfiguration = channelConfiguration
			}
			channelConfiguration = extendedChannelConfiguration // use the extendedChannelConfiguration
		}
		gASpecificConfig := map[int]int{
			1:  0,
			2:  0,
			3:  0,
			4:  0,
			6:  0,
			7:  0,
			17: 0,
			19: 0,
			20: 0,
			21: 0,
			22: 0,
			23: 0,
		}
		if _, ok := gASpecificConfig[audioObjectType]; !ok {
			return errors.New("unsupported audio object type")
		}
		sampleFrequencyValue := func() uint32 {
			if frequency > 92016 {
				return 96000
			} else if frequency > 75131 {
				return 88200
			} else if frequency > 55425 {
				return 64000
			} else if frequency > 46008 {
				return 48000
			} else if frequency > 37565 {
				return 44100
			} else if frequency > 27712 {
				return 32000
			} else if frequency > 23003 {
				return 24000
			} else if frequency > 18782 {
				return 22050
			} else if frequency > 13885 {
				return 16000
			} else if frequency > 11501 {
				return 12000
			} else if frequency > 9390 {
				return 11025
			} else {
				return 8000
			}
		}()
		_ = br.ReadBitsLE32(1) // frameLengthFlag
		dependOnCoreOrder := br.ReadBitsLE32(1)
		if dependOnCoreOrder != 0 {
			_ = br.ReadBitsLE32(14) // codeCoderDelay
		}
		_ = br.ReadBool() // extensionFlag
		channelCount := func() uint16 {
			if channelConfiguration == 0 {
				_ = br.ReadBitsLE32(4) // element_instance_tag
				_ = br.ReadBitsLE32(2) // object_type
				_ = br.ReadBitsLE32(4) // sampling_frequency_index
				numFrontChannel := br.ReadBitsLE32(4)
				numSideChannel := br.ReadBitsLE32(4)
				numBackChannel := br.ReadBitsLE32(4)
				numLfeChannel := br.ReadBitsLE32(2)
				_ = br.ReadBitsLE32(3) // num_assoc_data
				_ = br.ReadBitsLE32(4) // num_valid_cc
				monoMixdownPresent := br.ReadBool()
				if monoMixdownPresent {
					_ = br.ReadBitsLE32(4) // mono_mixdown_element_number
				}
				stereoMixdownPresent := br.ReadBool()
				if stereoMixdownPresent {
					_ = br.ReadBitsLE32(4) // stereo_mixdown_element_number
				}
				matrixMixdownIdxPresent := br.ReadBool()
				if matrixMixdownIdxPresent {
					_ = br.ReadBitsLE32(2) // matrix_mixdown_idx
					_ = br.ReadBitsLE32(1) // pseudo_surround_enable
				}
				channelCounts := uint16(0)
				readSurroundChannelCount := func(n uint32) uint16 {
					count := uint16(0)
					for i := 0; i < int(n); i++ {
						if ok := br.ReadBool(); ok {
							count += 2
						} else {
							count += 1
						}
						_ = br.ReadBitsLE32(4)
					}
					return count
				}
				channelCounts += readSurroundChannelCount(numFrontChannel)
				channelCounts += readSurroundChannelCount(numSideChannel)
				channelCounts += readSurroundChannelCount(numBackChannel)
				channelCounts += readSurroundChannelCount(numLfeChannel)
				return channelCounts
			} else if channelConfiguration >= 1 && channelConfiguration <= 7 {
				return uint16(int(channelConfiguration))
			} else if channelConfiguration == 11 {
				return 7 // 6.1 Amendment 4 of the AAC standard in 2013
			} else if channelConfiguration == 12 || channelConfiguration == 14 {
				return 8 // 7.1 (a/d) of ITU BS.2159
			} else {
				return 0xFFFF
			}
		}()
		p.AudioObjectType = audioObjectType
		p.ExtendedAudioObjectType = extendedAudioObjectType
		p.SampleRate = sampleFrequencyValue
		p.ChannelCount = channelCount
	}
	return nil
}

/*
Opus in ISO Base Media File Format. refer to : Encapsulation of Opus in ISO Base Media File Format https://www.opus-codec.org/docs/opus_in_isobmff.html#4.3.1

	class ChannelMappingTable (unsigned int(8) OutputChannelCount){
	                unsigned int(8) StreamCount;
	                unsigned int(8) CoupledCount;
	                unsigned int(8 * OutputChannelCount) ChannelMapping;
	            }

	aligned(8) class OpusSpecificBox extends Box('dOps'){
	                unsigned int(8) Version;
	                unsigned int(8) OutputChannelCount;
	                unsigned int(16) PreSkip;
	                unsigned int(32) InputSampleRate;
	                signed int(16) OutputGain;
	                unsigned int(8) ChannelMappingFamily;
	                if (ChannelMappingFamily != 0) {
	                    ChannelMappingTable(OutputChannelCount);
	                }
				}
*/
func (p *OpusDescriptor) parseDescriptor(r *atomReader) error {
	// Raw data
	p.DecoderSpecificInfo = make([]byte, r.a.Size())
	_ = copy(p.DecoderSpecificInfo, "OpusHead") // RFC-7845 add the Opus Magic Header
	opusData := make([]byte, r.a.bodySize)
	_, _ = r.ReadBytes(opusData)
	_ = copy(p.DecoderSpecificInfo[8:], opusData)
	_ = r.Move(int(-r.a.bodySize))
	p.Version = r.ReadUnsignedByte()
	p.OutputChannelCount = r.ReadUnsignedByte()
	p.PreSkip = r.Read2()
	p.InputSampleRate = r.Read4()
	p.OutputGain = r.Read2()
	p.ChannelMappingFamily = r.ReadUnsignedByte()
	p.StreamCount = r.ReadUnsignedByte()
	p.CoupledCount = r.ReadUnsignedByte()
	p.ChannelMapping = make([]byte, p.OutputChannelCount)
	_, _ = r.ReadBytes(p.ChannelMapping)
	return nil
}

// flac descriptor parser
func (p *FlacDescriptor) parseDescriptor(r *atomReader) error {
	length := r.a.bodySize
	if length <= 42 {
		return fmt.Errorf("%w : FlacDescriptor", ErrInvalidAtomSize)
	}
	_ = copy(p.DecoderSpecificInfo, "flaC")
	var err error = nil
	flaCData := make([]byte, length)
	if err = r.Peek(flaCData); err != nil {
		return fmt.Errorf("%w : FlacDescriptor.DecoderSpecificInfo", err)
	}
	_ = copy(p.DecoderSpecificInfo[4:], flaCData)
	version, flags := r.ReadVersionFlags()
	if version != 0 {
		return errors.New("unknown dfLa (FLAC) Version, unsupported")
	}
	if flags != 0 {
		return errors.New("no-zero dfLa (FLAC) flags, unsupported")
	}
	length -= 4
	// refer to https://github.com/xiph/flac/blob/master/doc/isoflac.txt
	metadataFraming := r.Read4()
	blockType := metadataFraming >> 24 & 0x7F
	if blockType != 0 {
		return errors.New("fLaCSpecificBox must have STREAMINFO metadata first")
	}
	blockLength := metadataFraming & 0x00ffffff
	if blockLength != 34 {
		return errors.New("fLaCSpecificBox STREAMINFO block is the wrong size")
	}
	p.StreamInfo = make([]byte, blockLength)
	_, _ = r.ReadBytes(p.StreamInfo)
	_ = r.Move(-34)
	p.SampleRate = 12<<p.StreamInfo[10] + 4<<p.StreamInfo[11] + (4<<p.StreamInfo[12])&0xf
	if p.SampleRate < 0 {
		return errors.New("fLaCSpecificBox STREAMINFO block must have no-zero sample rete")
	}

	p.ChannelCount = 4<<(p.StreamInfo[12])&0x7 + 1
	p.BitPerSample = int(4<<(p.StreamInfo[12]&1) + (p.StreamInfo[13]>>4)&0xf + 1)
	return nil
}

/*
alac bitstream storage in the ISO BMFF

	{
		uint32_t				frameLength;
		uint8_t					compatibleVersion;
		uint8_t					bitDepth;							// max 32
		uint8_t					pb;									// 0 <= pb <= 255
		uint8_t					mb;
		uint8_t					kb;
		uint8_t					numChannels;
		uint16_t				maxRun;
		uint32_t				maxFrameBytes;
		uint32_t				avgBitRate;
		uint32_t				sampleRate;
	} ALACSpecificConfig;
*/
func (p *AlacDescriptor) parseDescriptor(r *atomReader) {
	p.DecoderSpecificInfo = make([]byte, r.a.bodySize)
	_ = r.Peek(p.DecoderSpecificInfo)
	p.FrameLength = r.Read4()
	p.CompatibleVersion = r.ReadUnsignedByte()
	p.BitDepth = r.ReadUnsignedByte()
	p.Pb = r.ReadUnsignedByte()
	p.Mb = r.ReadUnsignedByte()
	p.Kb = r.ReadUnsignedByte()
	p.NumChannels = r.ReadUnsignedByte()
	p.MaxRun = r.Read2()
	p.MaxFrameBytes = r.Read4()
	p.AvgBitRate = r.Read4()
	p.SampleRate = r.Read4()
}

/*
ac-3 bitstream storage in the ISO BMFF :ac3specificBox

	{
		unsigned int(2) Fscod;
		unsigned int(5) Bsid;
		unsigned int(3) bsmod;
		unsigned int(3) Acmod;
		unsigned int(1) lfeon;
		unsigned int(5) bit_rate_code;
		unsigned int(5) reserved = 0;
	}
*/
func (p *Ac3Descriptor) parseDescriptor(r *atomReader) error {
	length := r.a.bodySize
	if length < 2 {
		return errors.New("content is too short for ac3 descriptor")
	}
	tmpByte := r.ReadUnsignedByte()
	p.Fscod = (tmpByte & 0xC0) >> 6
	sampleRateCodes := [3]uint32{48000, 44100, 32000}
	p.SampleRate = sampleRateCodes[p.Fscod]
	p.Bsid = tmpByte >> 1 & 0x1F
	channelCountsByAcmod := [...]uint16{2, 1, 2, 3, 3, 4, 4, 5}
	// only get the channel count form the Acmod , omit the bsmod, lfeon, bit_rate_code.
	p.ChannelCount = channelCountsByAcmod[(r.ReadUnsignedByte()&0x38)>>3]
	return nil
}

/* e-ac-3 bitstream storage in the ISO BMFF :e-ac3specificBox
{
	unsigned int(13) data_rate;
	unsigned int(3) num_ind_sub;
	unsigned int(2) Fscod;
	unsigned int(5) Bsid;
	unsigned int(5) bsmod;
	unsigned int(3) Acmod;
	unsigned int(1) lfeon;
	unsigned int(3) reserved = 0;
	unsigned int(4) num_dep_sub;
	if (num_dep_sub > 0)
		unsigned int(9) chan_loc;
	else
		unsigned int(1) reserved = 0;
}
*/

// parseConfig Dolby AC-4. refer to: ETSI TS 103 190-2 V1.1.1 (2015-09) “Digital Audio Compression (AC‐4) Standard” Annex E
func (p *Ac4Descriptor) parseDescriptor(r *atomReader) error {
	var err error = nil
	p.DecoderSpecificInfo = make([]byte, r.a.bodySize)
	if err = r.Peek(p.DecoderSpecificInfo); err != nil {
		return fmt.Errorf("%w : Ac4Descriptor.DecoderSpecificInfo", err)
	}
	_ = r.Move(1)
	fsIndex := r.ReadUnsignedByte() >> 5 & 0x1
	if fsIndex == 0 { // ETSI TS 103 190-1 [1], clause 4.3.3.2.5
		p.SampleRate = 44100
	} else {
		p.SampleRate = 48000
	}
	return nil
}

// refer to: IMPLEMENTATION OF DTS AUDIO IN MEDIA FILES BASED ON ISO/IEC 14496 Effective Date: February 2014
func (p *DtsDescriptor) parseDescriptor(r *atomReader) error {
	length := r.a.bodySize
	if length < 20 {
		return fmt.Errorf("%w : DtsDescriptor", ErrInvalidAtomSize)
	}
	var err error = nil
	p.DecoderSpecificInfo = make([]byte, length)
	if err = r.Peek(p.DecoderSpecificInfo); err != nil {
		return fmt.Errorf("%w : DtsDescriptor.DecoderSpecificInfo", err)
	}
	p.SamplingRate = r.Read4()
	p.MaxBiterate = r.Read4()
	p.AvgBiterate = r.Read4()
	p.PcmSampleDepth = r.ReadUnsignedByte()
	frameDurationTable := [...]uint16{512, 1024, 2048, 4096}
	tmpByte := r.ReadUnsignedByte()
	p.FrameDuration = frameDurationTable[tmpByte>>6]
	p.StreamConstruction = tmpByte >> 1 & 0x1F
	p.CoreLFEPresent = tmpByte & 0x1
	bitsBuff := make([]byte, 6)
	_, _ = r.ReadBytes(bitsBuff)
	br := newBitReaderFromSlice(bitsBuff)
	p.CoreLayout = br.ReadBitsLE8(6)
	p.CoreSize = br.ReadBitsLE16(14)
	p.StereoDownmix = br.ReadBitsLE8(1)
	p.RepresentationType = br.ReadBitsLE8(3)
	p.ChannelLayout = br.ReadBitsLE16(16)
	p.MultiAssetFlag = br.ReadBitsLE8(1)
	p.LBRDurationMod = br.ReadBitsLE8(1)
	p.ReservedBoxPresent = br.ReadBitsLE8(1)
	// p.ReservedBoxPresent == 1 shows there are more box(es) following.
	return nil
}

/*
	MLPSpecificBox refer to: Dolby TrueHD (MLP) bitstreams within the ISO base media file format

https://developer.dolby.com/globalassets/technology/dolby-truehd/dolbytruehdbitstreamswithintheisobasemediafileformat.pdf

	{
		unsigned int(32) format_info;
		unsigned int(15) peak_data_rate;
		unsigned int(1) reserved = 0;
		unsigned int(32) reserved = 0;
	}
*/
func (p *MlpaDescriptor) parseDescriptor(r *atomReader) {
	p.DecoderSpecificInfo = make([]byte, r.a.bodySize)
	_ = r.Peek(p.DecoderSpecificInfo)
	p.FormatInfo = r.Read4()
	p.PeakDataRate = r.Read2() >> 1
}

/*
 parseConfig AVC file format. refer to: ISO/IEC 14496-15, 5.3.3.1.2
aligned(8) class AVCDecoderConfigurationRecord {
	unsigned int(8) configurationVersion = 1;
	unsigned int(8) AVCProfileIndication;
	unsigned int(8) profile_compatibility;
	unsigned int(8) AVCLevelIndication;
	bit(6) reserved = '111111'b;
	unsigned int(2) LengthSizeMinusOne;
	bit(3) reserved = '111'b;
	unsigned int(5) numOfSequenceParameterSets;
	for (i = 0; i < numOfSequenceParameterSets; i++) {
		unsigned int(16) sequenceParameterSetLength ;
		bit(8*sequenceParameterSetLength) sequenceParameterSetNALUnit;
	}
	unsigned int(8) numOfPictureParameterSets;
	for (i = 0; i < numOfPictureParameterSets; i++) {
		unsigned int(16) pictureParameterSetLength;
		bit(8*pictureParameterSetLength) pictureParameterSetNALUnit;
	}
	if (profile_idc == 100 || profile_idc == 110 ||
	    profile_idc == 122 || profile_idc == 144)
	{
		bit(6) reserved = '111111'b;
		unsigned int(2) chroma_format;
		bit(5) reserved = '11111'b;
		unsigned int(3) bit_depth_luma_minus8;
		bit(5) reserved = '11111'b;
		unsigned int(3) bit_depth_chroma_minus8;
		unsigned int(8) numOfSequenceParameterSetExt;
		for (i = 0; i < numOfSequenceParameterSetExt; i++) {
			unsigned int(16) sequenceParameterSetExtLength;
			bit(8*sequenceParameterSetExtLength) sequenceParameterSetExtNALUnit;
		}
	}
}
*/

func (p *AvcConfig) parseConfig(r *atomReader) error {
	if r.a.bodySize < 0 {
		return fmt.Errorf("%w : AvcConfig", ErrInvalidAtomSize)
	}
	var err error = nil
	p.DecoderSpecificInfo = make([]byte, r.Size())
	if err = r.Peek(p.DecoderSpecificInfo); err != nil {
		return fmt.Errorf("%w : AvcConfig.DecoderSpecificInfo", err)
	}
	p.Version = r.ReadUnsignedByte()
	p.ProfileIndication = r.ReadUnsignedByte()
	p.ProfileCompatibility = r.ReadUnsignedByte()
	p.AvcLevel = r.ReadUnsignedByte()
	lengthSizeMinusOne := r.ReadUnsignedByte() & 0x3
	p.LengthSize = lengthSizeMinusOne + 1
	if p.LengthSize == 3 {
		return errors.New("invalid length size in avc config")
	}
	numOfSequenceParameterSets := r.ReadUnsignedByte() & 0x1F
	for i := uint8(0); i < numOfSequenceParameterSets; i++ {
		sps := make([]byte, r.Read2())
		_, _ = r.ReadBytes(sps)
		p.ListSPS = append(p.ListPPS, sps)
	}
	numOfPictureParameterSets := r.ReadUnsignedByte()
	for i := uint8(0); i < numOfPictureParameterSets; i++ {
		pps := make([]byte, r.Read2())
		_, _ = r.ReadBytes(pps)
		p.ListPPS = append(p.ListPPS, pps)
	}
	//
	return err
}

/*
refer to
ISO/IEC 14496-15:2014 Information technology — Coding of audio-visual objects —
Part 15: Carriage of network abstraction layer (NAL) unit structured video in ISO base media file format

	Ch. 8.3.3.1

online source: https://www.iso.org/obp/ui/#iso:std:iso-iec:14496:-15:ed-3:v1:cor:1:v1:en

	aligned(8) class HEVCDecoderConfigurationRecord {
	   unsigned int(8) configurationVersion = 1;
	   unsigned int(2) general_profile_space;
	   unsigned int(1) general_tier_flag;
	   unsigned int(5) general_profile_idc;
	   unsigned int(32) general_profile_compatibility_flags;
	   unsigned int(48) general_constraint_indicator_flags;
	   unsigned int(8) general_level_idc;
	   bit(4) reserved = ‘1111’b;
	   unsigned int(12) min_spatial_segmentation_idc;
	   bit(6) reserved = ‘111111’b;
	   unsigned int(2) ParallelismType;
	   bit(6) reserved = ‘111111’b;
	   unsigned int(2) chroma_format_idc;
	   bit(5) reserved = ‘11111’b;
	   unsigned int(3) bit_depth_luma_minus8;
	   bit(5) reserved = ‘11111’b;
	   unsigned int(3) bit_depth_chroma_minus8;
	   bit(16) AvgFrameRate;
	   bit(2) ConstantFrameRate;
	   bit(3) NumTemporalLayers;
	   bit(1) TemporalIdNested;
	   unsigned int(2) LengthSizeMinusOne;
	   unsigned int(8) NumOfArrays;
	   for (j=0; j < NumOfArrays; j++) {
	      bit(1) ArrayCompleteness;
	      unsigned int(1) reserved = 0;
	      unsigned int(6) NAL_unit_type;
	      unsigned int(16) NumNalus;
	      for (i=0; i< NumNalus; i++) {
	         unsigned int(16) NalUnitLength;
	         bit(8*NalUnitLength) NalUnit;
	      }
	   }
	}
*/
func (p *HevcConfig) parseConfig(r *atomReader) error {
	if r.Size() < 0 {
		return fmt.Errorf("%w : HevcConfig", ErrInvalidAtomSize)
	}
	var err error = nil
	p.DecoderSpecificInfo = make([]byte, r.Size())
	if err = r.Peek(p.DecoderSpecificInfo); err != nil {
		return fmt.Errorf("%w : HevcConfig.DecoderSpecificInfo", err)
	}
	bitsBuff := make([]byte, 23)
	_, _ = r.ReadBytes(bitsBuff)
	br := newBitReaderFromSlice(bitsBuff)
	_ = br.ReadBitsLE8(8)
	p.GeneralProfileSpace = br.ReadBitsLE8(2)
	p.GeneralTierFlag = br.ReadBitsLE8(1)
	p.GeneralProfileIdc = br.ReadBitsLE8(5)
	p.GeneralProfileCompatibilityFlags = br.ReadBitsLE32(32)
	p.GeneralConstraintIndicatorFlags = br.ReadBitsLE64(48)
	p.GeneralLevelIdc = br.ReadBitsLE8(8)
	_ = br.ReadBitsLE8(4)
	p.MinSpatialSegmentationIdc = br.ReadBitsLE16(12)
	_ = br.ReadBitsLE8(6)
	p.ParallelismType = br.ReadBitsLE8(2)
	_ = br.ReadBitsLE8(6)
	p.ChromaFormatIdc = br.ReadBitsLE8(2)
	_ = br.ReadBitsLE8(5)
	p.BitDepthLumaMinus8 = br.ReadBitsLE8(3)
	_ = br.ReadBitsLE8(5)
	p.BitDepthChromaMinus8 = br.ReadBitsLE8(3)
	p.AvgFrameRate = br.ReadBitsLE16(16)
	p.ConstantFrameRate = br.ReadBitsLE8(2)
	p.NumTemporalLayers = br.ReadBitsLE8(3)
	p.TemporalIdNested = br.ReadBitsLE8(1)
	p.LengthSizeMinusOne = br.ReadBitsLE8(2)
	p.NumOfArrays = br.ReadBitsLE8(8)

	for i := uint8(0); i < p.NumOfArrays; i++ {
		nalUint := new(NalUnitInfo)
		tmpU8 := r.ReadUnsignedByte()
		nalUint.ArrayCompleteness = tmpU8 >> 7
		nalUint.NALUnitType = tmpU8 & 0x3F
		nalUint.NumNalus = r.Read2()
		for j := uint16(0); j < nalUint.NumNalus; j++ {
			nalLen := r.Read2()
			nal := make([]byte, nalLen)
			nalUint.NalUnitLength = append(nalUint.NalUnitLength, nalLen)
			_, _ = r.ReadBytes(nal)
			nalUint.NalUnit = append(nalUint.NalUnit, nal)
		}
	}
	return err
}

/*
refer to: https://aomediacodec.github.io/av1-isobmff/

	aligned (8) class AV1CodecConfigurationRecord {
	  unsigned int (1) marker = 1;
	  unsigned int (7) Version = 1;
	  unsigned int (3) seq_profile;
	  unsigned int (5) seq_level_idx_0;
	  unsigned int (1) seq_tier_0;
	  unsigned int (1) high_bitdepth;
	  unsigned int (1) twelve_bit;
	  unsigned int (1) Monochrome;
	  unsigned int (1) chroma_subsampling_x;
	  unsigned int (1) chroma_subsampling_y;
	  unsigned int (2) chroma_sample_position;
	  unsigned int (3) reserved = 0;

	  unsigned int (1) initial_presentation_delay_present;
	  if (initial_presentation_delay_present) {
	    unsigned int (4) initial_presentation_delay_minus_one;
	  } else {
	    unsigned int (4) reserved = 0;
	  }
	  unsigned int (8)[] configOBUs;
	}
*/
func (p *Av1cConfig) parseConfig(r *atomReader) error {
	if r.Size() < 0 {
		return fmt.Errorf("%w : Av1cConfig", ErrInvalidAtomSize)
	}
	var err error = nil
	p.DecoderSpecificInfo = make([]byte, r.Size())
	if err = r.Peek(p.DecoderSpecificInfo); err != nil {
		return fmt.Errorf("%w : Av1cConfig.DecoderSpecificInfo", err)
	}
	bitsBuff := make([]byte, 4)
	_, _ = r.ReadBytes(bitsBuff)
	br := newBitReaderFromSlice(bitsBuff)
	_ = br.ReadBitsLE8(8)
	p.SeqProfile = br.ReadBitsLE8(3)
	p.SeqLevelIdx0 = br.ReadBitsLE8(5)
	p.SeqTier0 = br.ReadBitsLE8(1)
	p.HighBitdepth = br.ReadBitsLE8(1)
	p.TwelveBit = br.ReadBitsLE8(1)
	p.Monochrome = br.ReadBitsLE8(1)
	p.ChromaSubsamplingX = br.ReadBitsLE8(1)
	p.ChromaSubsamplingY = br.ReadBitsLE8(1)
	p.ChromaSamplePosition = br.ReadBitsLE8(2)
	_ = br.ReadBitsLE8(3)
	p.InitialPresentationDelayPresent = br.ReadBitsLE8(1)
	if p.InitialPresentationDelayPresent == 1 {
		p.InitialPresentationDelayMinusOne = br.ReadBitsLE8(4)
	}
	//  configOBUs not read...
	return err
}

/* refer to: https://www.webmproject.org/vp9/mp4/

class VPCodecConfigurationBox extends FullBox('vpcC', Version = 1, 0){
      VPCodecConfigurationRecord() VpcConfig;
}

aligned (8) class VPCodecConfigurationRecord {
    unsigned int (8)     Profile;
    unsigned int (8)     Level;
    unsigned int (4)     BitDepth;
    unsigned int (3)     ChromaSubsampling;
    unsigned int (1)     VideoFullRangeFlag;
    unsigned int (8)     ColourPrimaries;
    unsigned int (8)     TransferCharacteristics;
    unsigned int (8)     MatrixCoefficients;
    unsigned int (16)    CodecIntializationDataSize;
    unsigned int (8)[]   CodecIntializationData;
}
*/

func (p *VpcConfig) parseConfig(r *atomReader) error {
	if r.Size() < 0 {
		return fmt.Errorf("%w : VpcConfig", ErrInvalidAtomSize)
	}
	var err error = nil
	p.DecoderSpecificInfo = make([]byte, r.Size())
	if err = r.Peek(p.DecoderSpecificInfo); err != nil {
		return fmt.Errorf("%w : VpcConfig.DecoderSpecificInfo", err)
	}
	_, _ = r.ReadVersionFlags()
	p.Profile = r.ReadUnsignedByte()
	p.Level = r.ReadUnsignedByte()
	tmpU8 := r.ReadUnsignedByte()
	p.BitDepth = tmpU8 >> 4
	p.ChromaSubsampling = tmpU8 & 0x0E
	p.VideoFullRangeFlag = tmpU8 & 0x1
	p.ColourPrimaries = r.ReadUnsignedByte()
	p.TransferCharacteristics = r.ReadUnsignedByte()
	p.MatrixCoefficients = r.ReadUnsignedByte()
	p.CodecIntializationDataSize = r.Read2()
	p.CodecIntializationData = make([]byte, r.Read2())
	_, _ = r.ReadBytes(p.CodecIntializationData)
	return err
}

/*
	refer to:

Dolby Vision Streams Within the ISO Base Media File Format (Version 2.1.2)
As part 2.1 of the document defined, the dvcC/dvvC box should follow hevC/avcC box.
class DOVIConfigurationBox extends Box(‘dvcC’ or ‘dvvC’)

	{
	 DOVIDecoderConfigurationRecord() DOVIConfig;
	}

align(8) class DOVIDecoderConfigurationRecord

	{
	 unsigned int (8) dv_version_major;
	 unsigned int (8) dv_version_minor;
	 unsigned int (7) dv_profile;
	 unsigned int (6) DvLevel;
	 bit (1) rpu_present_flag;
	 bit (1) el_present_flag;
	 bit (1) bl_present_flag;
	 unsigned int (4) dv_bl_signal_compatibility_id;
	 const unsigned int (28) reserved = 0;
	 const unsigned int (32)[4] reserved = 0;
	}
*/
func (p *DvcConfig) parseConfig(r *atomReader) error {
	if r.Size() < 0 {
		return fmt.Errorf("%w : DvcConfig", ErrInvalidAtomSize)
	}
	var err error = nil
	p.DecoderSpecificInfo = make([]byte, r.Size())
	if err = r.Peek(p.DecoderSpecificInfo); err != nil {
		return fmt.Errorf("%w : DvcConfig.DecoderSpecificInfo", err)
	}
	p.DvVersionMajor = r.ReadUnsignedByte()
	p.DvVersionMinor = r.ReadUnsignedByte()
	bitsBuff := make([]byte, 3)
	_, _ = r.ReadBytes(bitsBuff)
	br := newBitReaderFromSlice(bitsBuff)
	p.DvProfile = br.ReadBitsLE8(7)
	p.DvLevel = br.ReadBitsLE8(6)
	p.RpuPresentFlag = br.ReadBitsLE8(1)
	p.ElPresentFlag = br.ReadBitsLE8(1)
	p.BlPresentFlag = br.ReadBitsLE8(1)
	p.DvBlSingalCompatibilityId = br.ReadBitsLE8(4)
	return err
}
