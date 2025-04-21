package gitlib

import (
	"compress/zlib"
	"errors"
	"io"
)

type OfsDeltaObject struct {
	Id string
	BaseObjectOffset int64
	CommandList []DeltaCommand
	PackIndex *PackIndex
	rawData []byte
}

func (rd OfsDeltaObject) Type() GitObjectType { return OFS_DELTA }
func (rd OfsDeltaObject) ObjectId() string { return rd.Id }
func (rd OfsDeltaObject) RawData() []byte { return rd.rawData }

// This is a re-interpretation of:
// https://github.com/git/git/blob/master/builtin/unpack-objects.c#L463-L481
func readOfsDeltaOffset(f io.Reader) (int64, error) {
	baseOffset := int64(0)
	bytebuf := make([]byte, 1)
	_, err := io.ReadFull(f, bytebuf)
	if err != nil { return 0, err }
	baseOffset = int64(bytebuf[0])&0x7f
	for bytebuf[0]&0x80 > 0 {
		
		baseOffset += 1
		_, err := io.ReadFull(f, bytebuf)
		if err != nil { return 0, err }
		baseOffset = (baseOffset<<7) + (int64(bytebuf[0])&0x7f)
	}
	return baseOffset, nil
}

func (rgo RawGitObject) ReadAsOfsDeltaObject() (*OfsDeltaObject, error) {
	if rgo.objType != OFS_DELTA { return nil, errors.New("Not a OFS_DELTA object") }
	offset, err := readOfsDeltaOffset(rgo.reader)
	if err != nil { return nil, err }
	decompressed, err := zlib.NewReader(rgo.reader)
	if err != nil { return nil, err }
	data := make([]byte, rgo.objSize)
	_, err = io.ReadFull(decompressed, data)
	if err != nil { return nil, err }
	// skip the first two varint; they won't be needed when executing delta.
	i := 0
	for data[i]&0x80 > 0 { i += 1 }
	i += 1
	for data[i]&0x80 > 0 { i += 1 }
	i += 1
	commandList, err := parseDeltaCommandList(data[i:])
	if err != nil {
		return nil, err
	}
	return &OfsDeltaObject{
		Id: rgo.objId,
		BaseObjectOffset: rgo.packOffset-offset,
		CommandList: commandList,
		PackIndex: rgo.packIndex,
		rawData: data,
	}, nil
}
