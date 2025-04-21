package gitlib

import (
	"errors"
	"io"
	"log"
	"os"
	"path"
	"strings"
)

type PackIndex struct {
	// 1 or 2.
	Version int
	// the one in the filename.
	PackId string
	// the file handler for the pack index file.
	// we retain this handler since the file could be hella big
	// (indices are 31-bit wide) and i'm wary of the threat of
	// the possible huge .idx files to the memory.
	// we ignore the crc32 for now. in the future we must take that
	// into consideration.
	// TODO: fix this (according to above)
	file *os.File
	parent *LocalGitRepository
}

func (gr *LocalGitRepository) makePackIndex(packId string) (*PackIndex, error) {
	p := path.Join(gr.GitDirectoryPath, "objects", "pack", "pack-"+packId+".idx")
	f, err := os.Open(p)
	if err != nil { return nil, err }
	// NOTE: we don't defer the closing of `f` because we need it to
	// read out data from .pack files down the line.
	magicNumber, err := readBigEndianUInt32(f)
	if err != nil { return nil, err }
	version := 1
	if magicNumber == 0xff744f63 {
		versionNumber, err := readBigEndianUInt32(f)
		if err != nil { return nil, err }
		version = int(versionNumber)
	}
	pi := PackIndex{
		Version: version,
		PackId: packId,
		file: f,
		parent: gr,
	}
	return &pi, nil
}


func (pi *PackIndex) Dispose() {
	pi.file.Close()
	pi.file = nil
}

func (gr LocalGitRepository) readAllPackIndex() (map[string]*PackIndex, error) {
	var res map[string]*PackIndex = make(map[string]*PackIndex)
	p := path.Join(gr.GitDirectoryPath, "objects", "pack")
	f, err := os.ReadDir(p)
	if err != nil { return nil, err }
	if len(f) <= 0 { return nil, nil }
	for _, item := range f {
		if item.IsDir() { continue }
		name := item.Name()
		if path.Ext(name) != ".idx" { continue }
		if !strings.HasPrefix(name, "pack-") { continue }
		name = name[len("pack-"):len(name)-len(".idx")]
		pi, err := gr.makePackIndex(name)
		if err != nil { continue }
		res[name] = pi
	}
	return res, nil
}

// s need to be lowercase.
func (pi PackIndex) lookupObjectId(s string) (int64, error) {
	indexHead := s[:2]
	indexTail := s[2:]
	switch pi.Version {
	case 1: return pi.lookupObjectIdV1(indexHead, indexTail)
	case 2: return pi.lookupObjectIdV2(indexHead, indexTail)
	default: return 0, errors.New("Invalid pack index version")
	}
}

func (pi PackIndex) GetAllObjectId() ([]string, error) {
	switch pi.Version {
	case 1: return pi.getAllObjectIdV1()
	case 2: return pi.getAllObjectIdV2()
	default: return nil, errors.New("Invalid version for pack index")
	}
}


func (pi PackIndex) openPackFile() (*os.File, error) {
	p := path.Join(pi.parent.GitDirectoryPath, "objects", "pack", "pack-"+pi.PackId+".pack")
	return os.Open(p)
}

// there are three types of packed object header:
// 1.  normal object, which is 1+n byte type&size in varint
// 2.  REF_DELTA, which is:
//     1.  1+n byte type&size in varint
//     2.  base object id (20-byte)
// 3.  OFS_DELTA, which is:
//     1.  1+n byte type&size in varint
//     2.  n byte offset 7-bit varint
// in all cases, the beginning 1+n byte size in varint always refers
// to the size *before* zlib compression.
// `parseObjectAtOffset` does not go beyond the type&size in varint
// part; the caller is expected to dispatch on the type of the resulting
// GitObjectHeader/RawGitObject.  according to the document[1]: The
// base object could also be deltified if it’s in the same
// pack. Ref-delta can also refer to an object outside the pack
// (i.e. the so-called "thin pack").

func parsePackedObjectHeaderAtOffset(pf *os.File, offset int64) (GitObjectHeader, error) {
	_, err := pf.Seek(offset, 0)
	if err != nil { log.Panic(err) }
	bytebuf := make([]byte, 1)
	_, err = io.ReadFull(pf, bytebuf)
	if err != nil { return GitObjectHeader{}, err }
	typenum := GitObjectType((bytebuf[0]>>4)&0x7)
	shiftCount := 4
	size := int64(bytebuf[0]&0xf)
	for (bytebuf[0]&0x80) != 0 {
		_, err = io.ReadFull(pf, bytebuf)
		if err != nil { return GitObjectHeader{}, err }
		size += int64(bytebuf[0]&0x7f)<<int64(shiftCount)
		shiftCount += 7
	}
	return GitObjectHeader{
		Type: typenum,
		Size: int(size),
	}, nil
}

func (pi PackIndex) openPackedObject(objid string) (RawGitObject, error) {
	offset, err := pi.lookupObjectId(objid)
	if err != nil { return RawGitObject{}, err }
	if offset == -1 { return RawGitObject{}, errors.New("No such object in this pack") }
	pf, err := pi.openPackFile()
	if err != nil { return RawGitObject{}, err }
	header, err := parsePackedObjectHeaderAtOffset(pf, offset)
	if err != nil { pf.Close(); return RawGitObject{}, err }
	return RawGitObject{
		objId: objid,
		objType: header.Type,
		objSize: header.Size,
		packIndex: &pi,
		reader: pf,
		readerIsUncompressed: false,
		packOffset: offset,
	}, nil
}

