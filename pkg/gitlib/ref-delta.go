package gitlib

import (
	"compress/zlib"
	"errors"
	"io"
)

type RefDeltaObject struct {
	Id string
	BaseObjectId string
	CommandList []DeltaCommand
	PackIndex *PackIndex
	rawData []byte
}

func (rd RefDeltaObject) Type() GitObjectType { return REF_DELTA }
func (rd RefDeltaObject) ObjectId() string { return rd.Id }
func (rd RefDeltaObject) RawData() []byte { return rd.rawData }

func (rgo RawGitObject) ReadAsRefDeltaObject() (*RefDeltaObject, error) {
	if rgo.objType != REF_DELTA { return nil, errors.New("Not a REF_DELTA object") }
	base, err := readBytesToHex(rgo.reader, 20)
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
	if err != nil { return nil, err }
	return &RefDeltaObject{
		Id: rgo.objId,
		BaseObjectId: base,
		CommandList: commandList,
		rawData: data,
	}, nil
}

