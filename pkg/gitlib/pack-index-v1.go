package gitlib

func (pi PackIndex) lookupObjectIdV1(indexHead string, indexTail string) (int64, error) {
	fanoutIdx := byteHexToInt(indexHead)
	_, err := pi.file.Seek(int64(fanoutIdx)*4, 0)
	// TODO: fix this error reporting after figuring how we should
	// deal with the error logically.
	if err != nil { return 0, err }
	fanoutValue, err := readBigEndianUInt32(pi.file)
	if err != nil { return 0, err }
	var startValue uint32 = 4 * 256
	if fanoutIdx > 0 {
		startOffset := fanoutIdx - 1
		_, err := pi.file.Seek(int64(startOffset)*4, 0)
		if err != nil { return 0, err }
		startValue, err = readBigEndianUInt32(pi.file)
		if err != nil { return 0, err }
	}
	// from this point forward we check 24-byte items in the
	// range of [startValue, fanoutValue].
	itemCount := (fanoutValue - startValue) / 24
	_, err = pi.file.Seek(int64(startValue), 0)
	if err != nil { return 0, err }
	for range itemCount {
		offset, err := readBigEndianUInt32(pi.file)
		if err != nil { return 0, err }
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
			return int64(offset), nil
		}
	}
	return -1, nil
}

func (pi PackIndex) getObjectCountV1() (int, error) {
	_, err := pi.file.Seek(int64(4*255), 0)
	if err != nil { return 0, err }
	count, err := readBigEndianUInt32(pi.file)
	if err != nil { return 0, err }
	return int(count), nil
}

func (pi PackIndex) getAllObjectIdV1() ([]string, error) {
	count, err := pi.getObjectCountV1()
	if err != nil { return nil, err }
	var res []string
	_, err = pi.file.Seek(int64(4*256), 0)
	if err != nil { return nil, err }
	for range count {
		_, err = pi.file.Seek(4, 1)
		if err != nil { return nil, err }
		s, err := readBytesToHex(pi.file, 20)
		if err != nil { return nil, err }
		res = append(res, s)
	}
	return res, nil
}

