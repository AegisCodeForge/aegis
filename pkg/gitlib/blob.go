package gitlib

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
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

func BlobObjectFromString(s string) *BlobObject {
	h := sha1.New()
	fmt.Fprintf(h, "blob %d\x00%s", len(s), s)
	oid := h.Sum(nil)
	return &BlobObject{
		Id: fmt.Sprintf("%x", oid),
		Data: []byte(s),
	}
}

func (lgr *LocalGitRepository) AddBlobObject(content string) (*BlobObject, error) {
	bobj := BlobObjectFromString(content)
	objPath := path.Join(bobj.Id[:2], bobj.Id[2:])
	objFullPath := path.Join(lgr.GitDirectoryPath, objPath)
	b := new(bytes.Buffer)
	w := zlib.NewWriter(b)
	fmt.Fprintf(w, "blob %d\x00%s", len(content), content)
	w.Close()
	err := os.MkdirAll(path.Join(lgr.GitDirectoryPath, bobj.Id[:2]), fs.ModeDir)
	if err != nil { return nil, err }
	f, err := os.OpenFile(objFullPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil { return nil, err }
	defer f.Close()
	_, err = f.Write(b.Bytes())
	if err != nil { return nil, err }
	return bobj, nil
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

