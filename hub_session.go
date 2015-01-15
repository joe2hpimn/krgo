package main

import (
	"fmt"

	"github.com/docker/docker/dockerversion"
	"github.com/docker/docker/registry"
)

func init() {
	dockerversion.VERSION = "1.4.1" //needed otherwise error 500 on push
}

type hubSession struct {
	registry.Session
}

//return a HubSession associated with the repository contained in imageName
func newHubSession(imageName, userName, password string) (*hubSession, error) {
	hostname, _, err := registry.ResolveRepositoryName(imageName)
	if err != nil {
		return nil, fmt.Errorf("failed to find repository for image %v: %v", imageName, err)
	}
	endpoint, err := registry.NewEndpoint(hostname, []string{})
	if err != nil {
		return nil, err
	}

	authConfig := &registry.AuthConfig{}
	if userName != "" && password != "" {
		authConfig.Username = userName
		authConfig.Password = password
	}

	var metaHeaders map[string][]string

	session, err := registry.NewSession(authConfig, registry.HTTPRequestFactory(metaHeaders), endpoint, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker hub session: %v", err)
	}

	return &hubSession{*session}, nil
}
