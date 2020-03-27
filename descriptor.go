package fmp4parser

import "errors"

func (p *esdsDescriptors) parseDescriptors(r *bufferHandler) error {
	r.Move(-8)
	anchor := r.Position()
	dscriptorSize := r.Read4()
	_ = r.Read4()
	r.Move(4) // Version(8 bits) + flags(24 bits)
	for i := 0; i < 3; i++ {
		_ = p.findDescriptor(r, anchor+int64(dscriptorSize))
	}
	return nil
}

func (p *esdsDescriptors) findDescriptor(r *bufferHandler, endPos int64) error {
	// defined by 14496-1, section:7.2.2.1
	var esdescrTag uint8 = 0x03
	var decoderConfigTag uint8 = 0x04
	var decoderSpecificTag uint8 = 0x05
	tag := r.ReadByte() // tag's name
	// get the esds' length

	currentLen := r.ReadByte()
	size := currentLen & 0x7f
	for currentLen&0x80 == 0x80 {
		currentLen := r.ReadByte()
		size = size<<7 | currentLen&0x7f
	}
	if uint64(size) > 1<<30 {
		return errors.New("when get esds descriptor, the size is invalid")
	}
	// Start of the ES_Descriptor (defined in 14496-1)
	if uint8(tag) == esdescrTag {
		r.Move(2) // ES_ID
		flags := r.ReadByte()
		if flags&0x80 != 0 { // streamDependenceFlag
			r.Move(2)
		}
		if flags&0x40 != 0 { // uURL_Flag
			len := r.Read2()
			r.Move(int64(len))
		}
		if flags&0x20 != 0 { // OCRstreamFlag
			r.Move(2)
		}
	}
	// Start of the DecoderConfigDescriptor (defined in 14496-1)
	if uint8(tag) == decoderConfigTag {
		objectProfile := r.ReadByte()
		p.audioCodec = getMediaTypeFromObjectType(objectProfile)
		if p.audioCodec == int(0xFFFF) {
			p.audioCodec = audioUNKNOW
		}
		r.Move(12)
	}

	// Start of the DecoderSpecificInfo
	if uint8(tag) == decoderSpecificTag {
		currentPos := r.Position()
		p.decoderSpecificData, _, _ = r.ReadBytes(int(endPos - currentPos))
		_ = r.MoveTo(currentPos)
		frequencyTable := [13]int{96000, 88200, 64000, 48000, 44100, 32000, 24000, 22050, 16000, 12000, 11025, 8000, 7350}

		b, _, _ := r.ReadBytes(int(endPos - r.Position()))
		bitReader := newBitReaderFromString(string(b))
		audioObjectType := bitReader.ReadBits(5)
		if audioObjectType == 31 {
			// https://github.com/FFmpeg/FFmpeg/blob/a0ac49e38ee1d1011c394d7be67d0f08b2281526/libavcodec/mpeg4audio.h#L102
			audioObjectTypeExt := bitReader.ReadBits(6)
			audioObjectType = 32 + audioObjectTypeExt
		}
		sampleIndex := bitReader.ReadBits(4)
		sampleFrequency := func() int {
			if sampleIndex == 0x0F {
				return bitReader.ReadBits(24)
			} else {
				if sampleIndex < len(frequencyTable) {
					return frequencyTable[sampleIndex]
				}
				return -1
			}
		}()
		channelConfiguration := bitReader.ReadBits(4)
		extendedAudioObjectType := 0
		if audioObjectType == 5 || audioObjectType == 29 {
			extendedAudioObjectType = 5
			extendedSampleIndex := bitReader.ReadBits(4)
			sampleFrequency = func() int {
				if extendedSampleIndex == 0x0F {
					return bitReader.ReadBits(24)
				} else {
					if extendedSampleIndex < len(frequencyTable) {
						return frequencyTable[extendedSampleIndex]
					}
					return -1
				}
			}()
			audioObjectType := bitReader.ReadBits(5)
			if audioObjectType == 31 {
				audioObjectTypeExt := bitReader.ReadBits(6)
				audioObjectType = 32 + audioObjectTypeExt
			}
			if audioObjectType == 22 {
				channelConfiguration = bitReader.ReadBits(4)
			}
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
		if _, ok := gASpecificConfig[audioObjectType]; ok {
			if sampleFrequency == -1 {
				return errors.New("unknown frequency")
			}
			sampleFrequencyValue := func() int {
				if sampleFrequency > 92016 {
					return 96000
				} else if sampleFrequency > 75131 {
					return 88200
				} else if sampleFrequency > 55425 {
					return 64000
				} else if sampleFrequency > 46008 {
					return 48000
				} else if sampleFrequency > 37565 {
					return 44100
				} else if sampleFrequency > 27712 {
					return 32000
				} else if sampleFrequency > 23003 {
					return 24000
				} else if sampleFrequency > 18782 {
					return 22050
				} else if sampleFrequency > 13885 {
					return 16000
				} else if sampleFrequency > 11501 {
					return 12000
				} else if sampleFrequency > 9390 {
					return 11025
				} else {
					return 8000
				}
			}()
			_ = bitReader.ReadBits(1) // frameLengthFlag
			dependOnCoreOrder := bitReader.ReadBits(1)
			if dependOnCoreOrder != 0 {
				_ = bitReader.ReadBits(14) // codeCoderDelay
			}
			_ = bitReader.ReadBit() // extensionFlag
			channelCount := func() int {
				if channelConfiguration == 0 {
					_ = bitReader.ReadBits(4) // element_instance_tag
					_ = bitReader.ReadBits(2) // object_type
					_ = bitReader.ReadBits(4) // sampling_frequency_index
					numFrontChannel := bitReader.ReadBits(4)
					numSideChannel := bitReader.ReadBits(4)
					numBackChannel := bitReader.ReadBits(4)
					numLfeChannel := bitReader.ReadBits(2)
					_ = bitReader.ReadBits(3) // num_assoc_data
					_ = bitReader.ReadBits(4) // num_valid_cc
					monoMixdownPresent := bitReader.ReadBit()
					if monoMixdownPresent {
						_ = bitReader.ReadBits(4) // mono_mixdown_element_number
					}
					stereoMixdownPresent := bitReader.ReadBit()
					if stereoMixdownPresent {
						_ = bitReader.ReadBits(4) // stereo_mixdown_element_number
					}
					matrixMixdownIdxPresent := bitReader.ReadBit()
					if matrixMixdownIdxPresent {
						_ = bitReader.ReadBits(2) // matrix_mixdown_idx
						_ = bitReader.ReadBits(1) // pseudo_surround_enable
					}
					channelCounts := 0
					readSurroundChannelCount := func(n int) int {
						count := 0
						for i := 0; i < n; i++ {
							if ok := bitReader.ReadBit(); ok {
								count += 2
							} else {
								count += 1
							}
							_ = bitReader.ReadBits(4)
						}
						return count
					}

					channelCounts += readSurroundChannelCount(numFrontChannel)
					channelCounts += readSurroundChannelCount(numSideChannel)
					channelCounts += readSurroundChannelCount(numBackChannel)
					channelCounts += readSurroundChannelCount(numLfeChannel)
					return channelCounts
				} else if channelConfiguration >= 1 && channelConfiguration <= 7 {
					return channelConfiguration
				} else if channelConfiguration == 11 {
					return 7
				} else if channelConfiguration == 12 || channelConfiguration == 14 {
					return 8
				} else {
					logs.err.Println("invalid channel configuration")
					return -1
				}
			}()
			p.audioObjectType = audioObjectType
			p.extendedAudioObjectType = extendedAudioObjectType
			p.audioSampleRate = sampleFrequencyValue
			p.audioChannelCount = channelCount
		}
	}
	return nil
}
