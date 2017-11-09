package ksync

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	log "github.com/sirupsen/logrus"

	"github.com/vapor-ware/ksync/pkg/docker"
	"github.com/vapor-ware/ksync/pkg/service"
)

// BackgroundWatch starts up watch in the background via. docker container.
func BackgroundWatch(cfgPath string, upgrade bool) error {
	name := "ksync-watch"

	status, err := service.GetStatus(name)
	if err != nil {
		return err
	}

	if status.Running {
		if !upgrade {
			return fmt.Errorf("already running")
		}

		if stopErr := service.Stop(name); stopErr != nil {
			return stopErr
		}
	}

	cntr, err := docker.Client.ContainerCreate(
		context.Background(),
		&container.Config{
			Cmd: []string{
				"/ksync",
				// TODO: pull from config
				"--log-level=debug",
				"watch",
			},
			// TODO: make configurable
			Image: "gcr.io/elated-embassy-152022/ksync/ksync:canary",
			Labels: map[string]string{
				"heritage": "ksync",
			},
		},
		&container.HostConfig{
			// TODO: needs to be more configurable
			Binds: []string{
				fmt.Sprintf("%s:/root/.kube/config", KubeCfgPath),
				fmt.Sprintf("%s:/root/.ksync.yaml", cfgPath),
				// TODO: configurable?
				"/var/run/docker.sock:/var/run/docker.sock",
				"/:/host",
			},
			RestartPolicy: container.RestartPolicy{Name: "on-failure"},
		},
		&network.NetworkingConfig{},
		"ksync-watch")

	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"id": cntr.ID,
	}).Debug("container created")

	if err := service.Start(&cntr); err != nil { // nolint: megacheck
		return err
	}

	return nil
}