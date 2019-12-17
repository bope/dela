package delafs

import (
	"context"

	"bazil.org/fuse/fs"
	"github.com/bope/dela"
	log "github.com/sirupsen/logrus"
)

type FS struct {
	nodes  map[string]dela.Node
	nodeCh <-chan dela.Node
}

func NewFS(nodeCh <-chan dela.Node) *FS {
	return &FS{
		nodes:  make(map[string]dela.Node),
		nodeCh: nodeCh,
	}
}

func (dfs *FS) Root() (fs.Node, error) {
	return &RootDir{
		sfs: dfs,
	}, nil
}

func (dfs *FS) Collect(ctx context.Context) {
	for {
		select {
		case node := <-dfs.nodeCh:
			if _, found := dfs.nodes[node.Id]; found {
				continue
			}
			log.WithFields(log.Fields{
				"id":   node.Id,
				"name": node.Name,
				"url":  node.Url("/"),
			}).Info("node found")
			dfs.nodes[node.Id] = node
		case <-ctx.Done():
			return
		}
	}
}

func (dfs *FS) Remove(id string) {
	delete(dfs.nodes, id)
}
