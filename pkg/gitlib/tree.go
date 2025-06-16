package gitlib

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
)

// the actual current binary format of the tree object seems to
// not be documented anywhere. anyway, the format goes as follows:
// 1.  6 byte "object mode", digits in ascii
// 2.  1 byte space character 0x20
// 3.  variable length zero-terminated string, which is the file name;
// 4.  20 byte SHA-1 hash, which to us is the object id.
// tree objects represent folders within the repo.
// possible valid mode numbers are listed as follows:
// 100644  -  normal file.
// 100755  -  executable file.
// 120000  -  symbolic link.
// 040000  -  tree objects.
// 160000  -  submodules.
// according to git's source code, any file is either 100755
// or 100644, depending on the file permission they have when they're
// registered with git. also, any non-file non-link non-tree mode
// would be considered as a submodule mode.
// info source:
//    https://stackoverflow.com/questions/54596206/what-are-the-possible-modes-for-entries-in-a-git-tree-object
//    https://github.com/git/git/blob/ab336e8f1c8009c8b1aab8deb592148e69217085/cache.h#L285-L294

const (
	TREE_NORMAL_FILE = 100644
	TREE_EXECUTABLE_FILE = 100755
	TREE_SYMBOLIC_LINK = 120000
	TREE_TREE_OBJECT = 40000
	TREE_SUBMODULE = 160000
)

type TreeObjectItem struct {
	Mode int
	Name string
	Hash string
}

func (c TreeObjectItem) String() string {
	return fmt.Sprintf("<<[%d]%s,%s>>", c.Mode, c.Name, c.Hash)
}

type TreeObject struct {
	Id string
	ObjectList []TreeObjectItem
	rawData []byte
}

func (c TreeObject) String() string {
	return fmt.Sprintf("Tree{%s,%s}", c.Id, c.ObjectList)
}

func (c TreeObject) Type() GitObjectType { return TREE }
func (c TreeObject) ObjectId() string { return c.Id }
func (c TreeObject) RawData() []byte { return c.rawData }

func IsTreeObj(c GitObject) bool {
	_, ok := c.(*TreeObject)
	return ok
}

func resolveCanonicalMode(m int64) int64 {
	switch m {
	case TREE_NORMAL_FILE: return TREE_NORMAL_FILE
	case TREE_EXECUTABLE_FILE: return TREE_EXECUTABLE_FILE
	case TREE_SYMBOLIC_LINK: return TREE_SYMBOLIC_LINK
	case TREE_TREE_OBJECT: return TREE_TREE_OBJECT
	default: return TREE_SUBMODULE
	}
}

// NOTE: requires `s` to be pass the common object header.
func (rgo RawGitObject) ReadAsTreeObject() (*TreeObject, error) {
	if rgo.objType != TREE { return nil, errors.New("Not a tree object") }
	decompressedReader, err := zlib.NewReader(rgo.reader)
	if err != nil { return nil, err }
	sourceBytes := make([]byte, rgo.objSize)
	_, err = io.ReadFull(decompressedReader, sourceBytes)
	if err != nil { return nil, err }
	newReader := bytes.NewReader(sourceBytes)
	resobj, err := parseTreeObject(rgo.objId, newReader)
	if err != nil { return nil, err }
	resobj.rawData = sourceBytes
	return resobj, nil
}

// NOTE: requires `s` to be pass the common object header.
func (rgo RawGitObject) readAsTreeObjectNoDeflate() (*TreeObject, error) {
	if rgo.objType != TREE { return nil, errors.New("Not a tree object") }
	sourceBytes := make([]byte, rgo.objSize)
	_, err := io.ReadFull(rgo.reader, sourceBytes)
	if err != nil { return nil, err }
	newReader := bytes.NewReader(sourceBytes)
	resobj, err := parseTreeObject(rgo.objId, newReader)
	if err != nil { return nil, err }
	resobj.rawData = sourceBytes
	resobj.Id = rgo.objId
	return resobj, nil
}

func parseTreeObject(objid string, f io.Reader) (*TreeObject, error) {
	submoduleList := make([]TreeObjectItem, 0)
	dirList := make([]TreeObjectItem, 0)
	fileList := make([]TreeObjectItem, 0)
	for {
		modeAndName, err := readZeroTerminatedString(f)
		if err != nil { break }
		modeAndNameList := strings.Split(modeAndName, " ")
		mode, err := strconv.ParseInt(modeAndNameList[0], 10, 64)
		if err != nil { return nil, err }
		mode = int64(resolveCanonicalMode(mode))
		objid, err := readBytesToHex(f, 20)
		if err != nil { return nil, err }
		treeItem := TreeObjectItem{
			Mode: int(mode),
			Name: modeAndNameList[1],
			Hash: objid,
		}
		switch mode {
		case TREE_SUBMODULE:
			submoduleList = append(submoduleList, treeItem)
		case TREE_TREE_OBJECT:
			dirList = append(dirList, treeItem)
		default:
			fileList = append(fileList, treeItem)
		}
	}
	objlist := slices.Concat(submoduleList, dirList, fileList)
	tree := TreeObject{
		Id: objid,
		ObjectList: objlist,
	}
	return &tree, nil
}

func (gr LocalGitRepository) ResolveTreePath(t *TreeObject, p string) (GitObject, error) {
	var gobj GitObject = t
	var err error = nil
	var tobj *TreeObject = t
	for item := range strings.SplitSeq(p, "/") {
		tobj = gobj.(*TreeObject)
		if len(item) <= 0 { continue }
		found := false
		if item == "." { continue }
		for _, sub := range tobj.ObjectList {
			if sub.Name == item {
				found = true;
				gobj, err = gr.ReadObject(sub.Hash)
				if err != nil { return nil, err }
				break
			}
		}
		if !found {
			return nil, errors.New(fmt.Sprintf("Cannot find object named %s in tree", item))
		}
	}
	return gobj, nil
}

