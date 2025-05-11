package gitlib

// full tree.

import "errors"

type TreeNodeType byte

const (
	FILE = 1
	DIRECTORY = 2
)

type TreeFileNode struct {
	Type TreeNodeType
	Name string
	Path string
	Id string
}

type TreeDirNode struct {
	Type TreeNodeType
	Name string
	Path string
	Id string
	Children map[string]TreeNode
}

type TreeNode interface {
	GetType() TreeNodeType
	GetName() string
	GetPath() string
}

func (f *TreeFileNode) GetName() string { return f.Name }
func (f *TreeFileNode) GetPath() string { return f.Path }
func (f *TreeFileNode) GetType() TreeNodeType { return f.Type }
func (f *TreeDirNode) GetName() string { return f.Name }
func (f *TreeDirNode) GetPath() string { return f.Path }
func (f *TreeDirNode) GetType() TreeNodeType { return f.Type }

func NewFileNode(name string, path string) *TreeFileNode {
	return &TreeFileNode{ Type: FILE, Name: name, Path: path }
}

func NewDirNode(name string, path string, children map[string]TreeNode) *TreeDirNode {
	ch := children
	if ch != nil { ch = make(map[string]TreeNode, 0) }
	return &TreeDirNode{
		Type: DIRECTORY,
		Name: name,
		Path: path,
		Children: ch,
	}
}

var ErrTreeNodeAlreadyExist = errors.New("Tree node with the same name already exists.")

func (dn *TreeDirNode) AddNode(name string, node TreeNode) error {
	// add a node to a dir node.
	// the caller doesn't need to worry about managing path; since
	// this function would manage that for you.
	if dn.Children == nil {
		dn.Children = make(map[string]TreeNode, 0)
	}
	_, ok := dn.Children[name]
	if ok { return ErrTreeNodeAlreadyExist }
	dn.Children[name] = node
	return nil
}

func (lgr *LocalGitRepository) buildTree(stp string, name string, prefix string) (TreeNode, error) {
	obj, err := lgr.ReadObject(stp)
	if err != nil { return nil, err }
	p := prefix + "/" + name
	switch obj.Type() {
	case TREE:
		res := NewDirNode(name, p, nil)
		for _, item := range obj.(*TreeObject).ObjectList {
			tn, err := lgr.buildTree(item.Hash, item.Name, p)
			if err != nil { return nil, err }
			res.AddNode(item.Name, tn)
		}
		return res, nil
	case BLOB:
		res := NewFileNode(name, p)
		return res, nil
	default:
		// we ignore tag & commit objs. they shouldn't appear
		// in tree objects anyway.
		return nil, nil
	}
}

func (lgr *LocalGitRepository) BuildTree(startingPointId string, name string) (TreeNode, error) {
	// build a tree when given a repository and a "root" object id.
	// NOTE THAT blob object does not contain the file name itself; in
	// the case that `startingPointId` resolves to a blob object, the
	// resulting `TreeNode` would have the objectId as its name.  for
	// this reason it is advised to only use this method on tree
	// objects.
	
	// NOTE: i wish i could read a git object's metadata
	// without reading the whole object, but this is impossible due to
	// the existence of delta objects which git uses for both tree
	// objects and blob objects.
	// i'm genuinely worried about this being a pain point - i do not
	// wish to call this function for every request, but if we keep
	// a pre-calculated version in memory we'll have to deal with
	// syncing problem since updates to repositories are done through
	// ssh from another process that's transient. i'm planning to set
	// something up using zeromq and i shall truly fix this problem
	// after.
	r, err := lgr.ReadObject(startingPointId)
	if err != nil { return nil, err }
	if IsBlobObject(r) {
		return NewFileNode(name, name), nil
	}
	res := NewDirNode(name, name, nil)
	for _, item := range r.(*TreeObject).ObjectList {
		tn, err := lgr.buildTree(item.Hash, item.Name, name)
		if err != nil { return nil, err }
		res.AddNode(item.Name, tn)
	}
	return res, nil
}

