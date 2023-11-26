package main

func (track *boxTrak) constructPacketList() {
	movie := track.movie
	if track.edts != nil && track.edts.entryCount > 0 {
		elst := track.edts
		emptyDuration := uint64(0)
		editStartIndex := uint32(0)
		startTime := int64(0)
		for i := uint32(0); i < elst.entryCount; i++ {
			if i == 0 && elst.mediaTime[0] == -1 {
				// the first edit list is empty
				emptyDuration = elst.editDuration[0]
				editStartIndex = 1
			} else if i == editStartIndex && elst.mediaTime[i] >= 0 {
				startTime = elst.mediaTime[i]
			} else {
				// do nothing
			}
		}
		if movie.timeScale > 0 && (emptyDuration != 0 || startTime != 0) {
			if emptyDuration != 0 {
				emptyDuration = emptyDuration * uint64(track.timeScale) / uint64(movie.timeScale)
			}
			track.timeOffset = startTime - int64(emptyDuration)
		}
	}

	track.packets = make([]Packet, track.sampleNumber)

	// set DTS and PTS
	accuSample := 0
	accuDur := uint64(0)
	for i := 0; i < int(track.stts.entryCount); i++ {
		for j := 0; j < int(track.stts.sampleCount[i]); j++ {
			track.packets[accuSample].DTS = accuDur
			track.packets[accuSample].Duration = track.stts.sampleDelta[i]
			accuDur += uint64(track.stts.sampleDelta[i])
			accuSample++
		}
	}
	accuSample = 0
	for i := 0; i < int(track.ctts.entryCount); i++ {
		for j := 0; j < int(track.ctts.sampleCount[i]); j++ {
			track.packets[accuSample].PTS = uint64(int32(track.packets[accuSample].DTS) + int32(track.timeOffset) + track.ctts.sampleOffset[i])
			accuSample++
		}
	}

	// set sample size
	for i := uint32(0); i < track.stsz.sampleCount; i++ {
		if track.stsz.sampleSize == 0 {
			track.packets[i].Size = track.stsz.entrySize[i]
		} else {
			track.packets[i].Size = track.stsz.sampleSize
		}
	}

	// set sample offset
	chunkOffset := track.stco.chunkOffset
	accuSampleCount := 0
	lastChunkCount := 0
	lastFirstChunk := 0
	lastSamplePerCount := 0
	lastSampleDescriptionIndex := 0

	// get samples' offset and description index
	for i := 0; i <= int(track.stsc.entryCount); i++ {
		if lastChunkCount != 0 {
			totalSampleCount := lastChunkCount * lastSamplePerCount
			lastOffset := chunkOffset[lastFirstChunk]
			for j := 0; j < totalSampleCount; j++ {
				nextPacket := track.packets[accuSampleCount]
				nextPacket.offset = lastOffset
				nextPacket.DescriptorIndex = lastSampleDescriptionIndex
				lastOffset += uint64(nextPacket.Size) // move offset
				accuSampleCount++                     // increase 1
			}
		} else {
			if i == int(track.stsc.entryCount) {
				lastChunkCount = len(chunkOffset) - lastFirstChunk
			} else {
				lastChunkCount = int(track.stsc.firstChunk[i]-1) - lastFirstChunk
			}
			lastFirstChunk = int(track.stsc.firstChunk[i] - 1)
			lastSamplePerCount = int(track.stsc.samplePerChunk[i])
			lastSampleDescriptionIndex = int(track.stsc.sampleDescriptionIndex[i])
		}
	}
}
