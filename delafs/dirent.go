package delafs

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type Dirent struct {
	Name string          `json:"name"`
	Type fuse.DirentType `json:"type"`
	Size int64           `json:"size,omitempty"`
}

func (d Dirent) fuseDirent() fuse.Dirent {
	return fuse.Dirent{
		Name: d.Name,
		Type: d.Type,
	}
}

func (d Dirent) toNode(parent *Dir) fs.Node {
	if d.Type == fuse.DT_File {
		return &File{
			fs:   parent.fs,
			node: parent.node,
			path: parent.path + "/" + d.Name,
			size: d.Size,
		}
	}

	return &Dir{
		fs:      parent.fs,
		node:    parent.node,
		path:    parent.path + "/" + d.Name,
		dirents: make([]Dirent, 0),
	}
}
