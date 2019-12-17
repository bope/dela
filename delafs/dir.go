package delafs

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/bope/dela"
	log "github.com/sirupsen/logrus"
)

type Dir struct {
	fs      *FS
	node    *dela.Node
	path    string
	dirents []Dirent
}

var _ fs.HandleReadDirAller = (*Dir)(nil)
var _ fs.NodeStringLookuper = (*Dir)(nil)
var _ fs.Node = (*Dir)(nil)

func (d *Dir) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Mode = os.ModeDir | 0555
	return nil
}

func (d *Dir) fetch() {
	if len(d.dirents) != 0 {
		return
	}

	url := d.node.Url(d.path)

	log.WithFields(log.Fields{
		"url": url,
	}).Debug("listing files")

	req, err := http.Get(url)
	if err != nil {
		log.WithFields(
			log.Fields{
				"error": err.Error(),
			},
		).Error("list fetch error")
		d.fs.Remove(d.node.Id)
		return
	}
	defer req.Body.Close()
	if err = json.NewDecoder(req.Body).Decode(&d.dirents); err != nil {
		log.WithFields(
			log.Fields{
				"error": err.Error(),
			},
		).Error("json decode error")
		d.fs.Remove(d.node.Id)
		return
	}
}

func (d *Dir) find(name string) (Dirent, bool) {
	d.fetch()
	for _, d := range d.dirents {
		if d.Name == name {
			return d, true
		}
	}

	return Dirent{}, false
}

func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	d.fetch()
	ent, found := d.find(name)

	if !found {
		return nil, fuse.ENOENT
	}

	return ent.toNode(d), nil
}

func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	d.fetch()
	dirents := make([]fuse.Dirent, 0)
	for _, ent := range d.dirents {
		dirents = append(dirents, ent.fuseDirent())
	}
	return dirents, nil
}
