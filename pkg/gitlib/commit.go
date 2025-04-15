package gitlib

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

type CommitObject struct {
	Id string
	TreeObjId string
	ParentId string
	AuthorInfo AuthorTime
	CommitterInfo AuthorTime
	CommitTime time.Time
	CommitMessage string
	Signature string
	rawData []byte
}

func (c CommitObject) Type() int { return COMMIT }
func (c CommitObject) ObjectId() string { return c.Id }
func (c CommitObject) RawData() []byte { return c.rawData }

func IsCommitObject(gobj GitObject) bool {
	_, ok := gobj.(*CommitObject)
	return ok
}

func (c CommitObject) String() string {
	return fmt.Sprintf("Commit{%s,%s,%s}", c.ObjectId(), c.ParentId, c.TreeObjId)
}

// parse a commit object.
// NOTE: requires `s` to be pass the common object header.
func (rgo RawGitObject) ReadAsCommitObject() (*CommitObject, error) {
	if rgo.objType != COMMIT { return nil, errors.New("Not a commit object") }
	sourceBytes := make([]byte, rgo.objSize)
	decompressedReader, err := zlib.NewReader(rgo.reader)
	if err != nil { return nil, err }
	_, err = io.ReadFull(decompressedReader, sourceBytes)
	if err != nil { return nil, err }
	resobj, err := parseCommitObject(rgo.objId, bytes.NewReader(sourceBytes))
	if err != nil { return nil, err }
	resobj.rawData = sourceBytes
	return resobj, nil
}

// parse a commit object.
// NOTE: requires `s` to be pass the common object header.
func (rgo RawGitObject) readAsCommitObjectNoDeflate() (*CommitObject, error) {
	if rgo.objType != COMMIT { return nil, errors.New("Not a commit object") }
	sourceBytes := make([]byte, rgo.objSize)
	_, err := io.ReadFull(rgo.reader, sourceBytes)
	if err != nil { return nil, err }
	resobj, err := parseCommitObject(rgo.objId, bytes.NewReader(sourceBytes))
	if err != nil { return nil, err }
	resobj.rawData = sourceBytes
	resobj.Id = rgo.objId
	return resobj, nil
}

func parseCommitObject(objid string, f io.Reader) (*CommitObject, error) {
	sourceBytes, err := io.ReadAll(f)
	if err != nil { return nil, err }
	source := string(sourceBytes)
	splittedSource := strings.SplitN(source, "\n\n", 2)
	header := splittedSource[0]
	message := splittedSource[1]
	res := CommitObject{}
	sig := make([]string, 0)
	receivingSig := false
	for line := range strings.SplitSeq(header, "\n")  {
		if receivingSig {
			sig = append(sig, line)
			if strings.TrimSpace(line) == "-----END PGP SIGNATURE-----" {
				receivingSig = false
			}
			continue
		}
		lineType, lineContent, _ := strings.Cut(line, " ")
		switch lineType {
		case "tree":
			res.TreeObjId = lineContent
		case "parent":
			res.ParentId = lineContent
		case "author":
			res.AuthorInfo = parseAuthorTime(lineContent)
		case "committer":
			res.CommitterInfo = parseAuthorTime(lineContent)
		case "gpgsig":
			sig = append(sig, lineContent)
			receivingSig = true
		default:
		}
	}
	res.Id = objid
	res.CommitMessage = message
	res.rawData = sourceBytes
	res.Signature = strings.Join(sig, "\n")
	return &res, nil
}

