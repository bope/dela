package delafs

import (
	"context"
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type RootDir struct {
	sfs *FS
}

var _ fs.NodeStringLookuper = (*RootDir)(nil)
var _ fs.HandleReadDirAller = (*RootDir)(nil)

func (d *RootDir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = os.ModeDir | 0555
	return nil
}

func (d *RootDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	dirs := make([]fuse.Dirent, 0)

	for _, node := range d.sfs.nodes {
		dirs = append(dirs, fuse.Dirent{
			Name: node.Id,
			Type: fuse.DT_Dir,
		})
	}
	return dirs, nil
}

func (d *RootDir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	node, ok := d.sfs.nodes[name]
	if !ok {
		return nil, fuse.ENOENT
	}
	return &Dir{
		fs:      d.sfs,
		node:    &node,
		path:    "/",
		dirents: make([]Dirent, 0),
	}, nil
}
