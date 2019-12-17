package discover

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/bope/dela"
	"github.com/google/uuid"
	"github.com/micro/mdns"
)

type Service struct {
	queryInterval time.Duration
	service       string

	nodeChannel chan<- dela.Node
	nodeId      string
	nodeName    string
	mdnsService *mdns.MDNSService
	server      *mdns.Server

	ignore map[string]struct{}
}

func New(nodeName, nodeId string, nodeChannel chan<- dela.Node) *Service {
	if nodeId == "" {
		nodeId = uuid.New().String()
	}

	if nodeName == "" {
		var err error

		nodeName, err = os.Hostname()
		if err != nil {
			nodeName = nodeId
		}
	}

	return &Service{
		queryInterval: 1 * time.Second,
		service:       "_share._tcp",
		nodeId:        nodeId,
		nodeName:      nodeName,
		nodeChannel:   nodeChannel,
		ignore:        map[string]struct{}{nodeId: struct{}{}},
	}
}

func (s *Service) Discover(ctx context.Context) {
	ticker := time.NewTicker(s.queryInterval)
	entries := make(chan *mdns.ServiceEntry)

	go func() {
		for {
			select {
			case entry := <-entries:
				info, err := decode(entry.InfoFields)

				if err != nil {
					continue
				}

				if info.Service != s.service {
					continue
				}

				if _, found := s.ignore[info.Id]; found {
					continue
				}

				node := dela.Node{
					Id:   info.Id,
					Name: info.Name,
					Ip:   entry.AddrV4,
					Port: entry.Port,
				}
				s.nodeChannel <- node
			case <-ctx.Done():
				return
			}
		}
	}()

	var err error

	for {
		select {
		case <-ticker.C:
			query := &mdns.QueryParam{
				Service:             s.service,
				Domain:              "",
				Timeout:             time.Second,
				Entries:             entries,
				WantUnicastResponse: false,
			}

			if err = mdns.Query(query); err != nil {
				log.Printf("mdns query error: %s\n", err.Error())
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) Register(port int) error {
	info := &mdnsInfo{
		Service: s.service,
		Id:      s.nodeId,
		Name:    s.nodeName,
	}

	txt, err := encode(info)

	if err != nil {
		return err
	}

	s.mdnsService, err = mdns.NewMDNSService(s.nodeName, s.service, "", "", port, nil, txt)
	if err != nil {
		return err
	}

	s.server, err = mdns.NewServer(&mdns.Config{Zone: s.mdnsService})
	return err
}

func (s *Service) Deregister() error {
	return s.server.Shutdown()
}
