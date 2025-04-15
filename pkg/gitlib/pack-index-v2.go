package gitlib

func (pi PackIndex) lookupObjectIdV2(indexHead string, indexTail string) (int64, error) {
	// the layout of v2 pack idx file is as follows.
	// 1.  4 byte magic number.
	// 2.  4 byte version number.
	// 3.  256 x 4 byte fanout table. (of which the last item is total
	//     number of object within this pack file)
	// 4.  itemCount x 20 byte name table.
	// 5.  itemCount x 4 byte CRC32.
	// 6.  itemCount x 4 byte level 1 offset.
	// 7.  nCount x 8 byte level 2 offset, where nCount is the number
	//     of level 1 offset that has a MSB of 1.
	// 8.  packfile checksum
	// 9.  idxfile checksum
	fanoutBase := int64(8)
	_, err := pi.file.Seek(int64(8+255*4), 0)
	if err != nil { return 0, err }
	totalItemCount, err := readBigEndianUInt32(pi.file)
	if err != nil { return 0, err }
	objNameTableBase := 8+4*256
	
	fanoutIdx := byteHexToInt(indexHead)
	_, err = pi.file.Seek(fanoutBase+int64(fanoutIdx)*4, 0)
	if err != nil { return 0, err }
	segmentEndIdx, err := readBigEndianUInt32(pi.file)
	if err != nil { return 0, err }
	var segmentStartIdx uint32 = 0
	if fanoutIdx > 0 {
		segmentStartFanoutIdx := fanoutIdx - 1
		_, err := pi.file.Seek(fanoutBase+int64(segmentStartFanoutIdx)*4, 0)
		if err != nil { return 0, err }
		segmentStartIdx, err = readBigEndianUInt32(pi.file)
		if err != nil { return 0, err }
	}
	// from this point forward we check 20-byte items in the
	// range of [startValue, fanoutValue].
	itemCount := segmentEndIdx - segmentStartIdx
	_, err = pi.file.Seek(int64(objNameTableBase)+int64(segmentStartIdx)*20, 0)
	if err != nil { return 0, err }
	inBatchIndex := uint32(0)
	found := false
	for i := range itemCount {
		objidbuf := make([]byte, 20)
		_, err = pi.file.Read(objidbuf)
		if err != nil { return 0, err }
		s := make([]byte, 38)
		j := 0
		for k, b := range objidbuf {
			// we don't need to check first byte; it's already checked.
			if k == 0 { continue }
			s[j] = digitToChar(b>>4)
			s[j+1] = digitToChar(b&0x0f)
			j += 2
		}
		if string(s) == indexTail {
			inBatchIndex = i
			found = true
			break
		}
	}
	if !found { return -1, nil }
	fullTableIndex := inBatchIndex + segmentStartIdx
	_, err = pi.file.Seek(levelOneOffsetBase + int64(fullTableIndex)*4, 0)
	if err != nil { return 0, err }
	levelOneOffset, err := readBigEndianUInt32(pi.file)
	if err != nil { return 0, err }
	if (levelOneOffset & 0x80000000) <= 0 {
		return int64(levelOneOffset), nil
	}
	_, err = pi.file.Seek(levelTwoOffsetBase + int64(fullTableIndex)*8, 0)
	if err != nil { return 0, err }
	levelTwoOffset, err := readBigEndianUInt64(pi.file)
	if err != nil { return 0, err }
	return int64(levelTwoOffset), nil
}

func (pi PackIndex) getObjectCountV2() (int, error) {
	_, err := pi.file.Seek(int64(8+4*255), 0)
	if err != nil { return 0, err }
	count, err := readBigEndianUInt32(pi.file)
	if err != nil { return 0, err }
	return int(count), nil
}

func (pi PackIndex) getAllObjectIdV2() ([]string, error) {
	count, err := pi.getObjectCountV2()
	if err != nil { return nil, err }
	var res []string
	_, err = pi.file.Seek(int64(8+4*256), 0)
	if err != nil { return nil, err }
	for range count {
		s, err := readBytesToHex(pi.file, 20)
		if err != nil { return nil, err }
		res = append(res, s)
	}
	return res, nil
}

