package fmp4parser

import (
	"errors"
	"math"
)

// parse AudioSampleEntry
func (p *boxStsd) parseAudioSampleEntry(r *deMuxReader, a *atom, isQuickTimeFormat bool) error {
	logD.Print("parsing moov.trak.mdia.minf.stbl.stsd,  box is ", a, r.Position())
	var err error = nil
	entryStartPosition := r.Position()
	r.Move(8) // 6-bytes reserved + 2-bytes data_reference_index ISOBMFF 8.5.2.2
	entrySize := a.atomSize
	entryType := a.atomType
	audioEntry := new(audioSampleEntry)
	if isQuickTimeFormat {
		audioEntry.quickTimeVersion = int(r.Read2())
		r.Move(6)
	} else {
		r.Move(8)
	}
	audioEntry.originalFormat = entryType
	audioEntry.format = entryType
	if audioEntry.quickTimeVersion == 0 || audioEntry.quickTimeVersion == 1 {
		audioEntry.channelCount = r.Read2() // 2 bytes
		audioEntry.sampleSize = r.Read2()   // 2bytes
		r.Move(4)                           // 2 bytes + 2 bytes (compressionID + packetsize)
		if audioEntry.sampleRate = uint32(r.Read2()); audioEntry.sampleRate == 0 {
			audioEntry.sampleRate = uint32(r.Read2())
		} else {
			r.Move(2)
		}
		if audioEntry.quickTimeVersion == 1 {
			audioEntry.qttfSamplesPerPacket = r.Read4()
			audioEntry.qttfBytesPerPacket = r.Read4()
			audioEntry.qttfBytesPerFrame = r.Read4()
			audioEntry.qttfBytesPerSample = r.Read4()
			logD.Print(audioEntry.qttfBytesPerFrame, audioEntry.qttfBytesPerPacket, audioEntry.qttfSamplesPerPacket, audioEntry.qttfBytesPerSample)
		}
	} else if audioEntry.quickTimeVersion == 2 {
		r.Move(16) // always[3,16,Minus2,0,65536], sizeOfStructOnly
		tmpSampleRate := r.Read8()
		audioEntry.sampleRate = uint32(math.Round(float64(tmpSampleRate)))
		audioEntry.channelCount = r.Read2()   // 2 bytes
		r.Move(4)                             // always 0x7F000000
		constBitsPerChannel := int(r.Read4()) //	constBitsPerChannel 4 bytes
		flags := int(r.Read4())
		r.Move(8) //	constBytesPerAudioPacket(32-bit) + constLPCMFramesPerAudioPacket(32-bit)
		if entryType == lpcmSampleEntry {
			// The way to deal with "lpcm" comes from ffmpeg. Very thanks
			bitsPerSample := processAudioEntryLPCM(constBitsPerChannel, flags)
			if bitsPerSample != 0 {
				audioEntry.qttfBytesPerSample = bitsPerSample
			}
		}
	} else {
		return ErrUnsupportedSampleEntry
	}
	newPosition := r.Position()
	// logD.Print(newPosition)
	// get information of Track Encryption Box
	if entryType == encaSampleEntry {
		logD.Print("parsing moov.trak.mdia.minf.stbl.stsd, current parsing audio sample entry: ", a)
		p.protectedInfo = new(ProtectedInformation)
		if err = processEncryptedSampleEntry(p.protectedInfo, r, entryStartPosition+entrySize-newPosition); err != nil {
			logD.Print("error, processEncryptedSampleEntry return ", err)
			return err
		}
		audioEntry.format = p.protectedInfo.DataFormat
	}
	audioEntry.descriptorsRawData = make(map[CodecType][]byte)
	audioEntry.decoderDescriptors = make(map[CodecType]interface{})
	_ = r.MoveTo(newPosition)
	for ; r.Position() < entryStartPosition+entrySize; _ = r.MoveTo(newPosition) {
		box := r.ReadAtomHeader()
		logD.Print("parsing moov.trak.mdia.stbl.stsd.audioSampleEntries, current box is ", box)
		newPosition = r.Position() + box.atomSize
		switch box.atomType {
		case fourCCwave:
			{
				if !isQuickTimeFormat {
					break
				}
				esdsSize, err := r.FindAtomWithinScope(fourCCesds, box.atomSize)
				if err != nil || esdsSize <= 0 {
					err = errors.New("not found esds box")
					logD.Print(err)
					break // not found an "esds" box
				}
				esds := new(EsDescriptor)
				err = esds.parseDescriptor(r)
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
				err = esds.parseDescriptor(r)
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
				err = opus.parseDescriptor(r, box.atomSize)
				audioEntry.codec = AudioCodecOPUS
				audioEntry.descriptorsRawData[audioEntry.codec] = opus.DecoderSpecificInfo
				audioEntry.decoderDescriptors[audioEntry.codec] = opus
				break
			}
		case fourCCdfla:
			{
				flac := new(FlacDescriptor)
				err = flac.parseDescriptor(r, box.atomSize)
				audioEntry.codec = AudioCodecFLAC
				audioEntry.descriptorsRawData[audioEntry.codec] = flac.DecoderSpecificInfo
				audioEntry.decoderDescriptors[audioEntry.codec] = flac
				break
			}
		case fourCCalac:
			{
				// https://github.com/macosforge/alac/blob/c38887c5c5e64a4b31108733bd79ca9b2496d987/codec/ALACAudioTypes.h#L162
				audioEntry.codec = AudioCodecALAC
				audioEntry.descriptorsRawData[audioEntry.codec], _, _ = r.ReadBytes(int(box.atomSize))
				audioEntry.decoderDescriptors = nil
				break
			}
		case fourCCdac3:
			{
				ac3 := new(Ac3Descriptor)
				err = ac3.parseAc3Descriptor(r, box.atomSize)
				audioEntry.sampleRate = ac3.SampleRate
				audioEntry.channelCount = ac3.ChannelCount
				audioEntry.codec = AudioCodecAC3
				audioEntry.decoderDescriptors[audioEntry.codec] = ac3
				break
			}
		case fourCCdec3:
			{
				eac3 := new(Ac3Descriptor)
				err = eac3.parseEac3Descriptor(r, box.atomSize)
				audioEntry.sampleRate = eac3.SampleRate
				audioEntry.channelCount = eac3.ChannelCount
				audioEntry.codec = AudioCodecEAC3
				audioEntry.decoderDescriptors[audioEntry.codec] = eac3
				break
			}
		case fourCCddts:
			{
				dts := new(DtsDescriptor)
				err = dts.parseDtsDescriptor(r, box.atomSize)
				audioEntry.codec = AudioCodecDTS
				audioEntry.descriptorsRawData[audioEntry.codec] = dts.DecoderSpecificInfo
				audioEntry.decoderDescriptors[audioEntry.codec] = dts
				break
			}
		case fourCCdac4:
			{
				ac4 := new(Ac4Descriptor)
				err = ac4.parseAC4Descriptor(r, box.atomSize)
				audioEntry.codec = AudioCodecAC4
				audioEntry.descriptorsRawData[audioEntry.codec] = ac4.DecoderSpecificInfo
				audioEntry.decoderDescriptors[audioEntry.codec] = ac4
				break
			}
		case fourCCdmlp:
			{
				mlpa := new(MlpaDescriptor)
				err = mlpa.parseMlaDescriptor(r, box.atomSize)
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
	p.audioSampleEntry = audioEntry
	return err
}

// parse VideoSampleEntry
func (p *boxStsd) parseVideoSampleEntry(r *deMuxReader, a *atom) error {
	var err error = nil
	entryStartPosition := r.Position()
	logD.Print("video entry pos = ", r.Position())
	entrySize := a.atomSize
	entryType := a.atomType
	videoEntry := new(videoSampleEntry)
	r.Move(6) // reserved
	videoEntry.dataReferenceIndex = r.Read2()
	r.Move(16) // reserved
	videoEntry.width = r.Read2()
	videoEntry.height = r.Read2()
	r.Move(46) // unused + reserved 14 bytes, compressorname_size + p_compressorname 32 bytes
	videoEntry.depth = r.Read2()
	r.Move(2) // pre-defined

	videoEntry.originalFormat = entryType
	videoEntry.format = entryType

	newPosition := r.Position()
	// get information of Track Encryption Box
	if entryType == encvSampleEntry {
		logD.Print("parsing moov.trak.mdia.minf.stbl.stsd, track is encrypted.")
		p.protectedInfo = new(ProtectedInformation)
		if err = processEncryptedSampleEntry(p.protectedInfo, r, entryStartPosition+entrySize-newPosition); err != nil {
			return err
		}
		videoEntry.format = p.protectedInfo.DataFormat
	}
	videoEntry.configurationRecordsRawData = make(map[CodecType][]byte)
	videoEntry.decoderConfigurationRecords = make(map[CodecType]interface{})
	_ = r.MoveTo(newPosition)
	logD.Print("1234  ", r.Position())
	for ; r.Position() < entryStartPosition+entrySize; _ = r.MoveTo(newPosition) {
		box := r.ReadAtomHeader()
		logD.Print(box)
		newPosition = r.Position() + box.atomSize
		switch box.atomType {
		case fourCCavcC:
			if entryType != avc1SampleEntry && entryType != avc3SampleEntry && entryType != encvSampleEntry {
				return errors.New("invalid video sample entry")
			}
			avc := new(AvcConfig)
			err = avc.parse(r, box.atomSize)
			videoEntry.codec = VideoCodecH264
			videoEntry.configurationRecordsRawData[videoEntry.codec] = avc.DecoderSpecificInfo
			videoEntry.decoderConfigurationRecords[videoEntry.codec] = avc
			break
		case fourCChvcC:
			if entryType != hev1SampleEntry && entryType != hvc1SampleEntry && entryType != hVC1SampleEntry && entryType != encvSampleEntry {
				return errors.New("invalid video sample entry")
			}
			hevc := new(HevcConfig)
			err = hevc.parse(r, box.atomSize)
			videoEntry.codec = VideoCodecHEVC
			videoEntry.configurationRecordsRawData[videoEntry.codec] = hevc.DecoderSpecificInfo
			videoEntry.decoderConfigurationRecords[videoEntry.codec] = hevc
			break
		case fourCCav1c:
			if entryType != av01SampleEntry {
				return errors.New("invalid video sample entry")
			}
			av1c := new(Av1cConfig)
			err = av1c.parse(r, box.atomSize)
			videoEntry.codec = VideoCodecAV1
			videoEntry.configurationRecordsRawData[videoEntry.codec] = av1c.DecoderSpecificInfo
			videoEntry.decoderConfigurationRecords[videoEntry.codec] = av1c
			break
		case fourCCvpcC:
			if entryType != vp08SampleEntry && entryType != vp09SampleEntry && entryType != encvSampleEntry {
				return errors.New("invalid video sample entry")
			}
			vpc := new(VpcConfig)
			err = vpc.parse(r, box.atomSize)
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
			err = dvc.parse(r, box.atomSize)
			videoEntry.codec = VideoCodecDolbyVision
			videoEntry.configurationRecordsRawData[videoEntry.codec] = dvc.DecoderSpecificInfo
			videoEntry.decoderConfigurationRecords[videoEntry.codec] = dvc
			break
		case fourCCcolr:
			parseColr(videoEntry, r, box.atomSize)
			break
		case fourCCpasp:
			parsePasp(videoEntry, r)
			break
		case fourCCclap:
			parseClap(videoEntry, r)
			break
		default:
			logD.Print("atom type in sample descriptor is not parsed yet, ", box)
			break
		}
	}
	p.videoSampleEntry = videoEntry
	logD.Print(videoEntry)
	return err
}

func processAudioEntryLPCM(constBitsPerChannel, flags int) uint32 {
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

func processEncryptedSampleEntry(p *ProtectedInformation, r *deMuxReader, leftSize int64) error {
	sinfSize, err := r.FindAtomWithinScope(fourCCsinf, leftSize)
	if err != nil || sinfSize < 8 {
		logD.Print("can not find the atom sinf")
		return ErrIncompleteCryptoBox
	}
	logD.Print("parsing moov.trak.mdia.stbl.stsd.enca.sinf, sinf size is ", sinfSize)
	leftSize = sinfSize
	stopPosition := r.Position() + leftSize
	for r.Position() < stopPosition {
		a := r.ReadAtomHeader()
		switch a.atomType {
		case fourCCfrma:
			p.DataFormat = r.Read4()
			logD.Print("parsing moov.trak.mdia.stbl.stsd.enca.sinf, current trak is encrypted, the format is: ", int2String(p.DataFormat))
			break
		case fourCCschm:
			r.Move(4)
			p.SchemeType = r.Read4()
			p.SchemeVersion = r.Read4()
			r.Move(a.atomSize - 12)
			break
		case fourCCschi:
			if p.SchemeType != encryptionShemeTypeCenc && p.SchemeType != encryptionShemeTypeCens &&
				p.SchemeType != encryptionShemeTypeCbcs && p.SchemeType != encryptionShemeTypeCbc1 {
				return errors.New("unsupported encrypted scheme type")
			}
			logD.Print("parsing moov.trak.mdia.stbl.stsd.enca.sinf, current track's encrypted shceme is: ", int2String(p.SchemeType))
			parseSchiFormat(p, r)
			break
		default:
			r.Move(a.atomSize)
			break
		}
	}
	return err
}

// parse 'tenc' box
func parseSchiFormat(p *ProtectedInformation, r *deMuxReader) {
	_ = r.ReadAtomHeader() // omit the header of "tenc" box
	p.TencVersion, _ = r.ReadVersionFlags()
	r.Move(1) // reversed
	if p.TencVersion == 0 {
		r.Move(1) // reversed
	} else {
		defaultByteBlock := r.ReadUnsignedByte()
		p.DefaultCryptByteBlock = (defaultByteBlock & 0xF0) >> 4
		p.DefaultSkipByteBlock = defaultByteBlock & 0x0F
	}
	p.DefaultIsProtected = r.ReadUnsignedByte()
	logD.Print("parsing moov.trak.mdia.stbl.stsd.enca.sinf, current track is protected? : ", p.DefaultIsProtected != 0)
	p.DefaultPerSampleIVSize = r.ReadUnsignedByte()
	p.DefaultKID, _, _ = r.ReadBytes(16)
	if p.DefaultIsProtected == 1 && p.DefaultPerSampleIVSize == 0 {
		p.DefaultConstantIVSize = r.ReadUnsignedByte()
		p.DefaultConstantIV, _, _ = r.ReadBytes(int(p.DefaultConstantIVSize))
	}
}

func parseColr(v *videoSampleEntry, r *deMuxReader, n int64) {
	colourType := r.Read4()
	v.colourType = colourType
	if colourType == 0x6e636c78 { // "nclx"
		v.colorPrimaries = r.Read2()
		v.transferCharacteristics = r.Read2()
		v.matrixCoefficients = r.Read2()
		v.fullRangeFlag = r.ReadUnsignedByte() != 0
	} else { // "rICC"
		v.iCCProfile, _, _ = r.ReadBytes(int(n - 4))
	}
}

func parsePasp(v *videoSampleEntry, r *deMuxReader) {
	v.hSpacing = r.Read4()
	v.vSpacing = r.Read4()
}

func parseClap(v *videoSampleEntry, r *deMuxReader) {
	v.cleanApertureWidthN = r.Read4()
	v.cleanApertureHeightD = r.Read4()
	v.cleanApertureHeightN = r.Read4()
	v.cleanApertureHeightD = r.Read4()
	v.horizOffN = r.Read4()
	v.horizOffD = r.Read4()
	v.vertOffN = r.Read4()
	v.vertOffD = r.Read4()
}
