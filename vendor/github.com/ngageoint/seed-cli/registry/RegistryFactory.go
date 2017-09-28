package registry

import (
	"strings"

	"github.com/ngageoint/seed-cli/registry/containeryard"
	"github.com/ngageoint/seed-cli/registry/dockerhub"
	"github.com/ngageoint/seed-cli/registry/v2"
)

type RepositoryRegistry interface {
	Name() string
	Ping() error
	Repositories(org string) ([]string, error)
	Tags(repository, org string) ([]string, error)
	Images(org string) ([]string, error)
}

type RepoRegistryFactory func(url, username, password string) (RepositoryRegistry, error)

func NewV2Registry(url, username, password string) (RepositoryRegistry, error) {
	v2registry, err := v2.New(url, username, password)
	if err != nil {
		if strings.Contains(url, "https://") {
			httpFallback := strings.Replace(url, "https://", "http://", 1)
			v2registry, err = v2.New(httpFallback, username, password)
		}
	}

	return v2registry, err
}

func NewDockerHubRegistry(url, username, password string) (RepositoryRegistry, error) {
	hub, err := dockerhub.New(url)
	if err != nil {
		if strings.Contains(url, "https://") {
			httpFallback := strings.Replace(url, "https://", "http://", 1)
			hub, err = dockerhub.New(httpFallback)
		}
	}

	return hub, err
}

func NewContainerYardRegistry(url, username, password string) (RepositoryRegistry, error) {
	yard, err := containeryard.New(url)
	if err != nil {
		if strings.Contains(url, "https://") {
			httpFallback := strings.Replace(url, "https://", "http://", 1)
			yard, err = containeryard.New(httpFallback)
		}
	}

	return yard, err
}

func CreateRegistry(url, username, password string) (RepositoryRegistry, error) {
	if !strings.HasPrefix(url, "http") {
		url = "https://" + url
	}

	v2, err1 := NewV2Registry(url, username, password)
	if err1 == nil && v2 != nil && v2.Ping() == nil {
		return v2, nil
	}

	hub, err2 := NewDockerHubRegistry(url, username, password)
	if err2 == nil && hub != nil && hub.Ping() == nil {
		return hub, nil
	}

	yard, err3 := NewContainerYardRegistry(url, username, password)
	if err3 == nil && yard != nil && yard.Ping() == nil {
		return yard, nil
	}

	return nil, err1
}
