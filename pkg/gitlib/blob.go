package gitlib

import (
	"compress/zlib"
	"errors"
	"fmt"
	"io"
)

// blob does not have a format; it's just the file itself.

type BlobObject struct {
	Id string
	Data []byte
}

func (c BlobObject) Type() GitObjectType { return BLOB }
func (c BlobObject) ObjectId() string { return c.Id }
func (c BlobObject) RawData() []byte { return c.Data }

func IsBlobObject(gobj GitObject) bool {
	_, ok := gobj.(*BlobObject)
	return ok
}

func (c BlobObject) String() string {
	return fmt.Sprintf("Blob{%s,%d}", c.ObjectId(), len(c.RawData()))
}

// NOTE: requires `s` to be pass the common object header.
func (rgo RawGitObject) ReadAsBlobObject() (*BlobObject, error) {
	if rgo.objType != BLOB { return nil, errors.New("Not a blob object") }
	sourceBytes := make([]byte, rgo.objSize)
	decompressedReader, err := zlib.NewReader(rgo.reader)
	if err != nil { return nil, err }
	_, err = io.ReadFull(decompressedReader, sourceBytes)
	if err != nil { return nil, err }
	res := BlobObject{
		Id: rgo.objId,
		Data: sourceBytes,
	}
	return &res, nil
}

// same as `ReadAsBlobObject` but does not use deflate - assumes `rgo.reader`
// returns already-deflated bytes. used when applying delta objects.
func (rgo RawGitObject) readAsBlobObjectNoDeflate() (*BlobObject, error) {
	if rgo.objType != BLOB { return nil, errors.New("Not a blob object") }
	sourceBytes := make([]byte, rgo.objSize)
	_, err := io.ReadFull(rgo.reader, sourceBytes)
	if err != nil { return nil, err }
	res := BlobObject{
		Id: rgo.objId,
		Data: sourceBytes,
	}
	return &res, nil
}

