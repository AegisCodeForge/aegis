package gitlib

import (
	"io"
	"os"
	"path"
	"strings"
)

type PackedRefIndexItem struct {
	Id string
	Name string
}

// this function reads the "packed-refs" file, not any of the "pack-xyz" files.
func (gr LocalGitRepository) readPackedRefIndex() ([]PackedRefIndexItem, error) {
	var res []PackedRefIndexItem
	p := path.Join(gr.GitDirectoryPath, "packed-refs")
	f, err := os.Open(p)
	if os.IsNotExist(err) { return nil, nil }
	if err != nil { return nil, err }
	defer f.Close()
	a, err := io.ReadAll(f)
	if err != nil { return nil, err }
	for item := range strings.SplitSeq(string(a), "\n") {
		if len(item) <= 0 { continue }
		if strings.HasPrefix(item, "#") { continue }
		splitted := strings.Split(item, " ")
		res = append(res, PackedRefIndexItem{
			Id: strings.TrimSpace(splitted[0]),
		    Name: strings.TrimSpace(splitted[1]),
		})
    }
	return res, nil
}

