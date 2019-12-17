package cmd

import (
	"context"
	"os"
	"os/signal"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/bope/dela"
	"github.com/bope/dela/delafs"
	"github.com/bope/dela/discover"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func recv(signalCh <-chan os.Signal, doneCh chan struct{}) {
	for {
		select {
		case sig := <-signalCh:
			log.WithFields(log.Fields{
				"signal": sig,
			}).Debug("signal received")
			doneCh <- struct{}{}
		}
	}
}

func Run(mountDir, shareDir string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalCh := make(chan os.Signal)
	done := make(chan struct{})
	signal.Notify(signalCh, os.Interrupt, os.Kill)

	go recv(signalCh, done)

	id := uuid.New().String()
	name, err := os.Hostname()

	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("error getting hostname")
		return err
	}

	nodes := make(chan dela.Node)
	disc := discover.New(name, id, nodes)

	if shareDir != "" {
		server, err := delafs.NewServer(shareDir)

		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("error creating listener")
			return err
		}

		log.WithFields(log.Fields{
			"name": name,
			"id":   id,
			"port": server.Port,
		}).Debug("registering mdns service")
		err = disc.Register(server.Port)
		defer func() {
			log.Debug("stopping mdns service")
			disc.Deregister()
		}()

		go server.Start()

		if mountDir == "" {
			<-done
		}
	}

	if mountDir != "" {
		clientfs := delafs.NewFS(nodes)
		log.Debug("starting mdns discovery")
		go disc.Discover(ctx)
		log.Debug("starting filesystem node collection")
		go clientfs.Collect(ctx)

		log.WithFields(log.Fields{
			"dir": mountDir,
		}).Info("mounting filesystem")
		fsc, err := fuse.Mount(
			mountDir,
			fuse.FSName("dela"),
			fuse.Subtype("dela"),
			fuse.LocalVolume(),
			fuse.VolumeName("dela"),
		)

		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("error mounting filesystem")
			return err
		}
		defer fsc.Close()
		go func() {
			if err = fs.Serve(fsc, clientfs); err != nil {
				log.WithFields(log.Fields{
					"error": err.Error(),
				}).Error("filesystem error")
			}
		}()

		<-done

		log.WithFields(log.Fields{
			"dir": mountDir,
		}).Info("unmounting filesystem")
		if err = fuse.Unmount(mountDir); err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("error unmounting filesystem")
			return err
		}

	}

	return nil
}
