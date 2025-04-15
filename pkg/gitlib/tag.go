package gitlib

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"strings"
)

type TagObject struct {
	Id string
	TaggedObjId string
	TaggedObjType string
	TagName string
	TaggerInfo AuthorTime
	TagMessage string
	Signature string
	rawData []byte
}

func (c TagObject) Type() int { return TAG }
func (c TagObject) ObjectId() string { return c.Id }
func (c TagObject) RawData() []byte { return c.rawData }

func (c TagObject) String() string {
	return fmt.Sprintf("Tag{%s,%s}", c.Id, c.TaggedObjId)
}

func (rgo RawGitObject) ReadAsTagObject() (*TagObject, error) {
	if rgo.objType != TAG { return nil, errors.New("Not a tag object") }
	sourceBytes := make([]byte, rgo.objSize)
	decompressedReader, err := zlib.NewReader(rgo.reader)
	if err != nil { return nil, err }
	_, err = io.ReadFull(decompressedReader, sourceBytes)
	if err != nil { return nil, err }
	resobj, err := parseTagObject(rgo.objId, bytes.NewReader(sourceBytes))
	if err != nil { return nil, err }
	resobj.rawData = sourceBytes
	resobj.Id = rgo.objId
	return resobj, nil
}

func (rgo RawGitObject) readAsTagObjectNoDeflate() (*TagObject, error) {
	if rgo.objType != TAG { return nil, errors.New("Not a tag object") }
	sourceBytes := make([]byte, rgo.objSize)
	_, err := io.ReadFull(rgo.reader, sourceBytes)
	if err != nil { return nil, err }
	resobj, err := parseTagObject(rgo.objId, bytes.NewReader(sourceBytes))
	if err != nil { return nil, err }
	resobj.rawData = sourceBytes
	resobj.Id = rgo.objId
	return resobj, nil
}

func parseTagObject(objid string, f io.Reader) (*TagObject, error) {
	sourceBytes, err := io.ReadAll(f)
	if err != nil { return nil, err }
	source := string(sourceBytes)
	splittedSource := strings.SplitN(source, "\n\n", 2)
	header := splittedSource[0]
	message := splittedSource[1]
	res := TagObject{}
	for line := range strings.SplitSeq(header, "\n"){
		lineType, lineContent, _ := strings.Cut(line, " ")
		switch lineType {
		case "object":
			res.TaggedObjId = lineContent
		case "type":
			res.TaggedObjType = lineContent
		case "tag":
			res.TagName = lineContent
		case "tagger":
			res.TaggerInfo = parseAuthorTime(lineContent)
		default:
		}
	}
	readingSignature := false
	sig := make([]string, 0)
	tagMessage := make([]string, 0)
	for line := range strings.SplitSeq(message, "\n") {
		// NOTE: unlike commit object, signed tags have their signatures
		// *after** the message with no field header (instead of a field
		// header of "gpgsig" like commit objects).
		trimmed := strings.TrimSpace(line)
		if readingSignature && trimmed == "-----END PGP SIGNATURE-----" {
			sig = append(sig, trimmed)
			readingSignature = false
		} else if readingSignature {
			sig = append(sig, line)
		} else if trimmed == "-----BEGIN PGP SIGNATURE-----" {
			sig = append(sig, trimmed)
			readingSignature = true
		} else {
			tagMessage = append(tagMessage, line)
		}
	}
	res.Id = objid
	res.TagMessage = strings.Join(tagMessage, "\n")
	res.Signature = strings.Join(sig, "\n")
	res.rawData = sourceBytes
	return &res, nil
}

