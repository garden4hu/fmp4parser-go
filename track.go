package main

func (track *boxTrak) constructPacketList(n int) {
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
}
