package dela

import (
	"fmt"
	"net"
)

type Node struct {
	Id   string
	Name string
	Ip   net.IP
	Port int
}

func (n *Node) Url(path string) string {
	return fmt.Sprintf("http://%s:%d%s", n.Ip, n.Port, path)
}
