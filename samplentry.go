package main

import (
	"errors"
	"math"
)

// parseConfig AudioSampleEntry
func (p *boxTrak) parseAudioSampleEntry(r *atomReader) error {
	var err error = nil
	_ = r.Move(8) // 6-bytes reserved + 2-bytes data_reference_index ISOBMFF 8.5.2.2
	entryType := r.a.atomType
	audioEntry := new(audioSampleEntry)
	if p.quickTimeFormat {
		audioEntry.quickTimeVersion = int(r.Read2())
		_ = r.Move(6)
	} else {
		_ = r.Move(8)
	}
	audioEntry.originalFormat = entryType
	audioEntry.format = entryType
	// Compatible with quicktime. In fact, AudioSampleEntry in ISOBMFF has the same layout with
	// the version 0 of quicktime.
	if audioEntry.quickTimeVersion == 0 || audioEntry.quickTimeVersion == 1 {
		audioEntry.channelCount = r.Read2() // 2 bytes
		audioEntry.sampleSize = r.Read2()   // 2bytes
		_ = r.Move(4)                       // 2 bytes + 2 bytes (compressionID + packetSize)
		if audioEntry.sampleRate = uint32(r.Read2()); audioEntry.sampleRate == 0 {
			audioEntry.sampleRate = uint32(r.Read2())
		} else {
			_ = r.Move(2)
		}
		if audioEntry.quickTimeVersion == 1 {
			audioEntry.qttfSamplesPerPacket = r.Read4()
			audioEntry.qttfBytesPerPacket = r.Read4()
			audioEntry.qttfBytesPerFrame = r.Read4()
			audioEntry.qttfBytesPerSample = r.Read4()
			logD.Print(audioEntry.qttfBytesPerFrame, audioEntry.qttfBytesPerPacket, audioEntry.qttfSamplesPerPacket, audioEntry.qttfBytesPerSample)
		}
	} else if audioEntry.quickTimeVersion == 2 {
		_ = r.Move(16) // it always [3,16,Minus2,0,65536], sizeOfStructOnly
		tmpSampleRate := r.Read8()
		audioEntry.sampleRate = uint32(math.Round(float64(tmpSampleRate)))
		audioEntry.channelCount = r.Read2()   // 2 bytes
		_ = r.Move(4)                         // always 0x7F000000
		constBitsPerChannel := int(r.Read4()) //	constBitsPerChannel 4 bytes
		flags := int(r.Read4())
		_ = r.Move(8) //	constBytesPerAudioPacket(32-bit) + constLPCMFramesPerAudioPacket(32-bit)
		if entryType == lpcmSampleEntry {
			// The way to deal with "lpcm" comes from ffmpeg. Very thanks
			bitsPerSample := p.processAudioEntryLPCM(constBitsPerChannel, flags)
			if bitsPerSample != 0 {
				audioEntry.qttfBytesPerSample = bitsPerSample
			}
		}
	} else {
		return ErrUnsupportedSampleEntry
	}

	// get information of Track Encryption Box
	if entryType == encaSampleEntry {
		sinf, err := r.FindSubAtom(fourCCsinf)
		if err != nil {
			return errors.New("not find valid protection box in encrypted track")
		}
		p.processEncryptedSampleEntry(sinf)
	}
	audioEntry.descriptorsRawData = make(map[CodecType][]byte)
	audioEntry.decoderDescriptors = make(map[CodecType]interface{})

	for {
		ar, err := r.GetNextAtom()
		if err != nil {
			if err == ErrNoMoreAtom {
				break
			} else {
				return err
			}
		}
		switch ar.a.atomType {
		case fourCCwave:
			{
				if !p.quickTimeFormat {
					break
				}
				esdsR, err := ar.FindSubAtom(fourCCesds)
				if err != nil {
					break
				}
				esds := new(EsDescriptor)
				err = esds.parseDescriptor(esdsR)
				audioEntry.channelCount = esds.ChannelCount
				audioEntry.sampleRate = esds.SampleRate
				audioEntry.codec = esds.AudioCodec
				audioEntry.descriptorsRawData[audioEntry.codec] = esds.DecoderSpecificInfo
				audioEntry.decoderDescriptors[audioEntry.codec] = esds
				logD.Printf("parsing moov.trak.mdia.stbl.stsd.audioSampleEntries, sample descriptor: wave/esds channel_count is %d sampleRate is %d", audioEntry.channelCount, audioEntry.sampleRate)
				break
			}
		case fourCCesds:
			{
				esds := new(EsDescriptor)
				err = esds.parseDescriptor(ar)
				audioEntry.channelCount = esds.ChannelCount
				audioEntry.sampleRate = esds.SampleRate
				audioEntry.codec = esds.AudioCodec
				audioEntry.descriptorsRawData[audioEntry.codec] = esds.DecoderSpecificInfo
				audioEntry.decoderDescriptors[audioEntry.codec] = esds
				logD.Printf("parsing moov.trak.mdia.stbl.stsd.audioSampleEntries, sample descriptor: esds channel_count is %d sampleRate is %d", audioEntry.channelCount, audioEntry.sampleRate)
				break
			}
		case fourCCdops:
			{
				opus := new(OpusDescriptor)
				err = opus.parseDescriptor(ar)
				audioEntry.codec = AudioCodecOPUS
				audioEntry.descriptorsRawData[audioEntry.codec] = opus.DecoderSpecificInfo
				audioEntry.decoderDescriptors[audioEntry.codec] = opus
				break
			}
		case fourCCdfla:
			{
				flac := new(FlacDescriptor)
				err = flac.parseDescriptor(ar)
				audioEntry.codec = AudioCodecFLAC
				audioEntry.descriptorsRawData[audioEntry.codec] = flac.DecoderSpecificInfo
				audioEntry.decoderDescriptors[audioEntry.codec] = flac
				break
			}
		case fourCCalac:
			{
				// https://github.com/macosforge/alac/blob/c38887c5c5e64a4b31108733bd79ca9b2496d987/codec/ALACAudioTypes.h#L162
				alac := new(AlacDescriptor)
				alac.parseDescriptor(ar)
				audioEntry.channelCount = uint16(alac.NumChannels)
				audioEntry.sampleRate = alac.SampleRate
				audioEntry.codec = AudioCodecALAC
				audioEntry.descriptorsRawData[audioEntry.codec] = alac.DecoderSpecificInfo
				audioEntry.decoderDescriptors[audioEntry.codec] = alac
				break
			}
		case fourCCdac3:
			{
				ac3 := new(Ac3Descriptor)
				err = ac3.parseDescriptor(ar)
				audioEntry.sampleRate = ac3.SampleRate
				audioEntry.channelCount = ac3.ChannelCount
				audioEntry.codec = AudioCodecAC3
				audioEntry.decoderDescriptors[audioEntry.codec] = ac3
				break
			}
		case fourCCdec3:
			{
				eac3 := new(Ac3Descriptor)
				err = eac3.parseDescriptor(ar)
				audioEntry.sampleRate = eac3.SampleRate
				audioEntry.channelCount = eac3.ChannelCount
				audioEntry.codec = AudioCodecEAC3
				audioEntry.decoderDescriptors[audioEntry.codec] = eac3
				break
			}
		case fourCCddts:
			{
				dts := new(DtsDescriptor)
				err = dts.parseDescriptor(ar)
				audioEntry.channelCount = dts.ChannelLayout
				audioEntry.codec = AudioCodecDTS
				audioEntry.descriptorsRawData[audioEntry.codec] = dts.DecoderSpecificInfo
				audioEntry.decoderDescriptors[audioEntry.codec] = dts
				break
			}
		case fourCCdac4:
			{
				ac4 := new(Ac4Descriptor)
				err = ac4.parseDescriptor(ar)
				audioEntry.codec = AudioCodecAC4
				audioEntry.sampleRate = ac4.SampleRate
				audioEntry.descriptorsRawData[audioEntry.codec] = ac4.DecoderSpecificInfo
				audioEntry.decoderDescriptors[audioEntry.codec] = ac4
				break
			}
		case fourCCdmlp:
			{
				mlpa := new(MlpaDescriptor)
				mlpa.parseDescriptor(ar)
				audioEntry.codec = AudioCodecMLP
				audioEntry.decoderDescriptors[audioEntry.codec] = mlpa
				break
			}
		default:
			{
				if entryType == alawSampleEntry {
					audioEntry.codec = AudioCodecALAW
				} else if entryType == ulawSampleEntry {
					audioEntry.codec = AudioCodecMULAW
				} else if entryType == dtshSampleEntry || entryType == dtslSampleEntry {
					audioEntry.codec = AudioCodecDTSHD
				} else if entryType == dtseSampleEntry {
					audioEntry.codec = AudioCodecDTSEXPRESS
				} else if entryType == lpcmSampleEntry || entryType == sowtSampleEntry || entryType == twosSampleEntry {
					audioEntry.codec = AudioCodecRAW
				} else if entryType == samrSampleEntry {
					audioEntry.codec = AudioCodecAMRNB
				} else if entryType == sawbSampleEntry {
					audioEntry.codec = AudioCodecAMRWB
				}
			}
			break
		}
	}
	p.audioEntry = audioEntry
	return err
}

// parseConfig VideoSampleEntry
func (p *boxTrak) parseVideoSampleEntry(r *atomReader) error {
	var err error = nil
	entryType := r.a.atomType
	videoEntry := new(videoSampleEntry)
	_ = r.Move(6) // reserved
	videoEntry.dataReferenceIndex = r.Read2()
	_ = r.Move(16) // reserved
	videoEntry.width = r.Read2()
	videoEntry.height = r.Read2()
	_ = r.Move(46) // unused + reserved 14 bytes, compressorname_size + p_compressorname 32 bytes
	videoEntry.depth = r.Read2()
	_ = r.Move(2) // pre-defined

	videoEntry.originalFormat = entryType
	videoEntry.format = entryType

	// get information of Track Encryption Box
	if entryType == encvSampleEntry {
		sinf, err := r.FindSubAtom(fourCCsinf)
		if err != nil {
			return errors.New("not find valid protection box in encrypted track")
		}
		p.processEncryptedSampleEntry(sinf)
	}
	videoEntry.configurationRecordsRawData = make(map[CodecType][]byte)
	videoEntry.decoderConfigurationRecords = make(map[CodecType]interface{})

	for {
		ar, err := r.GetNextAtom()
		if err != nil {
			if err == ErrNoMoreAtom {
				break
			}
		}
		switch ar.a.atomType {
		case fourCCavcC:
			if entryType != avc1SampleEntry && entryType != avc3SampleEntry && entryType != encvSampleEntry {
				return errors.New("invalid video sample entry")
			}
			avc := new(AvcConfig)
			err = avc.parseConfig(ar)
			videoEntry.codec = VideoCodecH264
			videoEntry.configurationRecordsRawData[videoEntry.codec] = avc.DecoderSpecificInfo
			videoEntry.decoderConfigurationRecords[videoEntry.codec] = avc
			break
		case fourCChvcC:
			if entryType != hev1SampleEntry && entryType != hvc1SampleEntry && entryType != hVC1SampleEntry && entryType != encvSampleEntry {
				return errors.New("invalid video sample entry")
			}
			hevc := new(HevcConfig)
			err = hevc.parseConfig(ar)
			videoEntry.codec = VideoCodecHEVC
			videoEntry.configurationRecordsRawData[videoEntry.codec] = hevc.DecoderSpecificInfo
			videoEntry.decoderConfigurationRecords[videoEntry.codec] = hevc
			break
		case fourCCav1c:
			if entryType != av01SampleEntry {
				return errors.New("invalid video sample entry")
			}
			av1c := new(Av1cConfig)
			err = av1c.parseConfig(ar)
			videoEntry.codec = VideoCodecAV1
			videoEntry.configurationRecordsRawData[videoEntry.codec] = av1c.DecoderSpecificInfo
			videoEntry.decoderConfigurationRecords[videoEntry.codec] = av1c
			break
		case fourCCvpcC:
			if entryType != vp08SampleEntry && entryType != vp09SampleEntry && entryType != encvSampleEntry {
				return errors.New("invalid video sample entry")
			}
			vpc := new(VpcConfig)
			err = vpc.parseConfig(ar)
			if entryType == vp08SampleEntry {
				videoEntry.codec = VideoCodecVP8
			} else {
				videoEntry.codec = VideoCodecVP9
			}
			videoEntry.configurationRecordsRawData[videoEntry.codec] = vpc.DecoderSpecificInfo
			videoEntry.decoderConfigurationRecords[videoEntry.codec] = vpc
			break
			// Dolby Vision configuration box should be parsed after by avcC/hvcC box
		case fourCCdvcC:
			fallthrough
		case fourCCdvvC:
			dvc := new(DvcConfig)
			err = dvc.parseConfig(ar)
			videoEntry.codec = VideoCodecDolbyVision
			videoEntry.configurationRecordsRawData[videoEntry.codec] = dvc.DecoderSpecificInfo
			videoEntry.decoderConfigurationRecords[videoEntry.codec] = dvc
			break
		case fourCCcolr:
			p.parseColr(videoEntry, ar)
			break
		case fourCCpasp:
			p.parsePasp(videoEntry, ar)
			break
		case fourCCclap:
			p.parseClap(videoEntry, ar)
			break
		default:
			logD.Print("atom type in sample descriptor is not parsed yet, ", ar.a)
			break
		}
	}
	p.videoEntry = videoEntry
	return err
}

func (p *boxTrak) processAudioEntryLPCM(constBitsPerChannel, flags int) uint32 {
	codec := func(bps int, flags int) lpcmCodecId {
		flt := flags & 1
		be := flags & 2
		sflags := 0
		if (flags & 4) != 0 {
			sflags = -1
		}
		if bps <= 0 || bps > 64 {
			return None
		}
		if flt != 0 {
			switch bps {
			case 32:
				if be == 0 {
					return pcmF32LE
				}
				return pcmF32BE
			case 64:
				if be == 0 {
					return pcmF64LE
				}
				return pcmF64BE
			default:
				return None
			}
		} else {
			bps += 7
			bps >>= 3
			if sflags&(1<<(bps-1)) != 0 {
				switch bps {
				case 1:
					return pcmS8
				case 2:
					if be == 0 {
						return pcmS16LE
					}
					return pcmS16BE
				case 3:
					if be == 0 {
						return pcmS24LE
					}
					return pcmS24BE
				case 4:
					if be == 0 {
						return pcmS32LE
					}
					return pcmS32BE
				case 8:
					if be == 0 {
						return pcmS64LE
					}
					return pcmS64BE
				default:
					return None
				}
			} else {
				switch bps {
				case 1:
					return pcmU8
				case 2:
					if be == 0 {
						return pcmU16LE
					}
					return pcmU16BE
				case 3:
					if be == 0 {
						return pcmU24LE
					}
					return pcmU24BE
				case 4:
					if be == 0 {
						return pcmU32LE
					}
					return pcmU32BE
				default:
					return None
				}
			}
		}
	}(constBitsPerChannel, flags)
	switch codec {
	case pcmS8:
		fallthrough
	case pcmU8:
		if constBitsPerChannel == 16 {
			codec = pcmS16BE
		}
		break
	case pcmS16LE:
		fallthrough
	case pcmS16BE:
		if constBitsPerChannel == 8 {
			codec = pcmS8
		} else if constBitsPerChannel == 24 {
			if codec == pcmS16BE {
				codec = pcmS24BE
			} else {
				codec = pcmS24LE
			}
		} else if constBitsPerChannel == 32 {
			if codec == pcmS16BE {
				codec = pcmS32BE
			} else {
				codec = pcmS32LE
			}
		}
	default:
	}
	return func(codec lpcmCodecId) uint32 {
		switch codec {
		case pcmALaw:
			fallthrough
		case pcmMULaw:
			fallthrough
		case pcmVIDC:
			fallthrough
		case pcmS8:
			fallthrough
		case pcmS8Planar:
			fallthrough
		case pcmU8:
			fallthrough
		case pcmZORK:
			return 8

		case pcmS16BE:
			fallthrough
		case pcmS16BEPlanar:
			fallthrough
		case pcmS16LE:
			fallthrough
		case pcmS16LEPlanar:
			fallthrough
		case pcmU16BE:
			fallthrough
		case pcmU16LE:
			return 16
		case pcmS24DAUD:
			fallthrough
		case pcmS24BE:
			fallthrough
		case pcmS24LE:
			fallthrough
		case pcmS24LEPlanar:
			fallthrough
		case pcmU24BE:
			fallthrough
		case pcmU24LE:
			return 24
		case pcmS32BE:
			fallthrough
		case pcmS32LE:
			fallthrough
		case pcmS32LEPlanar:
			fallthrough
		case pcmU32BE:
			fallthrough
		case pcmU32LE:
			fallthrough
		case pcmF32BE:
			fallthrough
		case pcmF32LE:
			fallthrough
		case pcmF24LE:
			fallthrough
		case pcmF16LE:
			return 32
		case pcmF64BE:
			fallthrough
		case pcmF64LE:
			fallthrough
		case pcmS64BE:
			fallthrough
		case pcmS64LE:
			return 64
		default:
			return 0
		}
	}(codec)
}

func (p *boxTrak) processEncryptedSampleEntry(r *atomReader) {
	protection := new(ProtectedInformation)
	for {
		a, err := r.GetNextAtom()
		if err == ErrNoMoreAtom {
			break
		}
		switch a.a.atomType {
		case fourCCfrma: // Original Format
			p.format = a.Read4() // data_format , coding name
			break
		case fourCCschm: // Scheme type
			_ = r.Move(4) // version + flags
			protection.SchemeType = r.Read4()
			protection.SchemeVersion = r.Read4()
			break
		case fourCCschi: // Scheme Information
			_ = r.ReadAtomHeader() // "tenc" header
			v, _ := r.ReadVersionFlags()
			_ = r.Move(1)
			if v == 0 {
				_ = r.Move(1)
			} else {
				defaultByteBlock := r.ReadUnsignedByte()
				protection.DefaultCryptByteBlock = (defaultByteBlock & 0xF0) >> 4
				protection.DefaultSkipByteBlock = defaultByteBlock & 0x0F
			}
			protection.DefaultIsProtected = r.ReadUnsignedByte()
			protection.DefaultPerSampleIVSize = r.ReadUnsignedByte()
			protection.DefaultKID = make([]byte, 16)
			_, _ = r.ReadBytes(protection.DefaultKID)
			if protection.DefaultIsProtected == 1 && protection.DefaultPerSampleIVSize == 0 {
				protection.DefaultConstantIVSize = r.ReadUnsignedByte()
				protection.DefaultConstantIV = make([]byte, protection.DefaultConstantIVSize)
				_, _ = r.ReadBytes(protection.DefaultConstantIV)
			}
			break
		}
	}
	p.protection = append(p.protection, protection)
}

func (p *boxTrak) parseColr(v *videoSampleEntry, r *atomReader) {
	colourType := r.Read4()
	v.colourType = colourType
	if colourType == 0x6e636c78 { // "nclx"
		v.colorPrimaries = r.Read2()
		v.transferCharacteristics = r.Read2()
		v.matrixCoefficients = r.Read2()
		v.fullRangeFlag = r.ReadUnsignedByte() != 0
	} else { // "rICC"
		v.iCCProfile = make([]byte, r.Size()-4)
		_, _ = r.ReadBytes(v.iCCProfile)
	}
}

func (p *boxTrak) parsePasp(v *videoSampleEntry, r *atomReader) {
	v.hSpacing = r.Read4()
	v.vSpacing = r.Read4()
}

func (p *boxTrak) parseClap(v *videoSampleEntry, r *atomReader) {
	v.cleanApertureWidthN = r.Read4()
	v.cleanApertureHeightD = r.Read4()
	v.cleanApertureHeightN = r.Read4()
	v.cleanApertureHeightD = r.Read4()
	v.horizOffN = r.Read4()
	v.horizOffD = r.Read4()
	v.vertOffN = r.Read4()
	v.vertOffD = r.Read4()
}
