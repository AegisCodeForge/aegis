package gitlib

import (
	"bytes"
	"io"
)

// things for handling "delta commands", the thing git uses to describe
// how one object should be built basing on another object.

type DeltaCopyCommand struct {
	offset int64
	size int64
}

type DeltaDataCommand struct {
	data []byte
}

type DeltaCommand interface {
	Execute(base []byte) []byte
}

/**
  There are two kinds of delta command. We call them copy and data here.
  Their types are determined by the most significant bit of their fist
  byte. if it's 1, then it's a copy command; if it's 0, then it's a data
  command.
  Data is simple. The rest 7 bits of the first byte determines the length
  (in bytes) of the data that follows.
  Copy is slightly complicated.
    + A copy command contains:
      + An offset (maximum 4 byte)
      + A size (maximum 3 bytes)
      Both of them are in little-endian order.
    + Whether a certain byte is present is determined by the rest 7 bits
      of the first byte, with the first least significant bit representing
      the first remaining byte (1 - present; 0 - not present). The layout
      is as follows:
      +-----+------+------+------+------+-----+-----+-----+
      | fsb | off1 | off2 | off3 | off4 | sz1 | sz2 | sz3 |
      +-----+------+------+------+------+-----+-----+-----+
      This means that, for example, if FSB is 0b10010001, only off1 and
      sz1 would be present, thus the command is like this:
      +----------+------+-----+
      | 10010001 | off1 | sz1 |
      +----------+------+-----+
      if FSB is 0b10000101 instead, the command would be like this:
      (only off1 and off3 present)
      +----------+------+------+
      | 10000101 | off1 | off3 |
      +----------+------+------+
      note that in this case the 3rd byte is off3 instead of off2. the
      missing bytes are considered as zero. For example, this:
      +----------+------+------+
      | 10000011 | 0xcd | 0xab |
      +----------+------+------+
      has an offset of 0xabcd, since the FSB is 0b10000011 which means
      the next two bytes are off1 and off2, but this:
      +----------+------+------+
      | 10000101 | 0xcd | 0xab |
      +----------+------+------+
      would have an offset of 0xab00cd, since the next two bytes are
      off1 and off3 (off2 considered 0).
    According to the document, any size field 0 is not considered as size 0
    but size 0x10000 instead.
 */


func (dcc DeltaCopyCommand) Execute(v []byte) []byte {
	return v[dcc.offset:dcc.offset+dcc.size]
}

func (ddc DeltaDataCommand) Execute(_ []byte) []byte {
	return ddc.data
}

// returns the number of bytes read.
func parseDeltaCommandList(b []byte) ([]DeltaCommand, error) {
	f := bytes.NewReader(b)
	res := make([]DeltaCommand, 0)
	bytebuf := make([]byte, 1)
	for {
		_, err := io.ReadFull(f, bytebuf)
		if err != nil { break }
		if bytebuf[0]&0x80 <= 0 {
			data := make([]byte, bytebuf[0])
			_, err := io.ReadFull(f, data)
			if err != nil { return nil, err }
			res = append(res, DeltaDataCommand{
				data: data,
			})
		} else {
			// don't bother with loop here; the simpler the better.
			h := bytebuf[0]
			byte1 := 0
			if h & 0x1 > 0 {
				_, err := io.ReadFull(f, bytebuf)
				if err != nil { return nil, err }
				byte1 = int(bytebuf[0])
			}
			byte2 := 0
			if h & 0x2 > 0 {
				_, err := io.ReadFull(f, bytebuf)
				if err != nil { return nil, err }
				byte2 = int(bytebuf[0])
			}
			byte3 := 0
			if h & 0x4 > 0 {
				_, err := io.ReadFull(f, bytebuf)
				if err != nil { return nil, err }
				byte3 = int(bytebuf[0])
			}
			byte4 := 0
			if h & 0x8 > 0 {
				_, err := io.ReadFull(f, bytebuf)
				if err != nil { return nil, err }
				byte4 = int(bytebuf[0])
			}
			byte5 := 0
			if h & 0x10 > 0 {
				_, err := io.ReadFull(f, bytebuf)
				if err != nil { return nil, err }
				byte5 = int(bytebuf[0])
			}
			byte6 := 0
			if h & 0x20 > 0 {
				_, err := io.ReadFull(f, bytebuf)
				if err != nil { return nil, err }
				byte6 = int(bytebuf[0])
			}
			byte7 := 0
			if h & 0x40 > 0 {
				_, err := io.ReadFull(f, bytebuf)
				if err != nil { return nil, err }
				byte7 = int(bytebuf[0])
			}
			offset := (byte4<<24)|(byte3<<16)|(byte2<<8)|byte1
			copysize := (byte7<<16)|(byte6<<8)|byte5
			if copysize == 0 { copysize = 0x10000 }
			res = append(res, DeltaCopyCommand{
				offset: int64(offset),
				size: int64(copysize),
			})
		}
	}
	return res, nil
}

