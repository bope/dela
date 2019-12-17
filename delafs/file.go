package delafs

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/bope/dela"
	log "github.com/sirupsen/logrus"
)

type File struct {
	fs   *FS
	node *dela.Node
	path string
	size int64
}

var _ fs.NodeOpener = (*File)(nil)
var _ fs.Node = (*File)(nil)

func (f *File) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Mode = 0444
	attr.Size = uint64(f.size)
	return nil
}

func (f *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	return &FileHandle{
		fs:   f.fs,
		node: f.node,
		path: f.path,
		size: f.size,
	}, nil
}

type FileHandle struct {
	fs   *FS
	node *dela.Node
	path string
	size int64
}

var _ fs.Handle = (*FileHandle)(nil)

func (fh *FileHandle) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	url := fh.node.Url(fh.path)

	log.WithFields(log.Fields{
		"url":    url,
		"offset": req.Offset,
		"size":   req.Size,
	}).Debug("reading file")

	buf := make([]byte, req.Size)

	fsreq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.WithFields(
			log.Fields{
				"error": err.Error(),
			},
		).Error("request error")
		fh.fs.Remove(fh.node.Id)
		return err
	}
	var rangeHeader string
	if req.Offset+int64(req.Size) > fh.size {
		rangeHeader = fmt.Sprintf("bytes=%d-", req.Offset)
	} else {
		rangeHeader = fmt.Sprintf("bytes=%d-%d", req.Offset, req.Offset+int64(req.Size))
	}

	fsreq.Header.Add("Range", rangeHeader)
	fsresp, err := http.DefaultClient.Do(fsreq)

	if err != nil {
		log.WithFields(
			log.Fields{
				"error": err.Error(),
			},
		).Error("request error")
		fh.fs.Remove(fh.node.Id)
		return err
	}

	defer fsresp.Body.Close()
	n, err := io.ReadFull(fsresp.Body, buf)

	if err == io.ErrUnexpectedEOF || err == io.EOF {
		err = nil
	}

	if err != nil {
		log.WithFields(
			log.Fields{
				"error": err.Error(),
			},
		).Error("request read error")
		return err
	}

	resp.Data = buf[:n]
	return err
}
