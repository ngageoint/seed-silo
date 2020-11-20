package registry

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ngageoint/seed-common/objects"
	"github.com/ngageoint/seed-silo/registry/containeryard"
	"github.com/ngageoint/seed-silo/registry/dockerhub"
	gitlab "github.com/ngageoint/seed-silo/registry/gitlab"
	v2 "github.com/ngageoint/seed-silo/registry/v2"
)

type RepositoryRegistry interface {
	Name() string
	Ping() error
	Repositories() ([]string, error)
	Tags(repository string) ([]string, error)
	Images() ([]string, error)
	ImagesWithManifests() ([]objects.Image, error)
	GetImageManifest(repoName, tag string) (string, error)
}

type RepoRegistryFactory func(url, org, username, password string) (RepositoryRegistry, error)

func NewV2Registry(url, org, username, password string) (RepositoryRegistry, error) {
	v2registry, err := v2.New(url, org, username, password)
	if err != nil {
		if strings.Contains(url, "https://") {
			httpFallback := strings.Replace(url, "https://", "http://", 1)
			v2registry, err = v2.New(httpFallback, org, username, password)
		}
	}

	return v2registry, err
}

func NewDockerHubRegistry(url, org, username, password string) (RepositoryRegistry, error) {
	hub, err := dockerhub.New(url, org, username, password)
	if err != nil {
		if strings.Contains(url, "https://") {
			httpFallback := strings.Replace(url, "https://", "http://", 1)
			hub, err = dockerhub.New(httpFallback, org, username, password)
		}
	}

	return hub, err
}

func NewContainerYardRegistry(url, org, username, password string) (RepositoryRegistry, error) {
	yard, err := containeryard.New(url, org, username, password)
	if err != nil {
		if strings.Contains(url, "https://") {
			httpFallback := strings.Replace(url, "https://", "http://", 1)
			yard, err = containeryard.New(httpFallback, org, username, password)
		}
	}

	return yard, err
}

//NewGitLabRegistry Creates a new GitLab registry
func NewGitLabRegistry(url, org, username, password string) (RepositoryRegistry, error) {
	// Extract group / project information from the org
	group, path, err := gitlab.ExtractOrgPath(url, org, password)

	git, err := gitlab.New(url, group, path, username, password)
	if err != nil {
		if strings.Contains(url, "https://") {
			httpFallback := strings.Replace(url, "https://", "http://", 1)
			git, err = gitlab.New(httpFallback, group, path, username, password)
		}
	}
	return git, err
}

func CreateRegistry(url, org, username, password string) (RepositoryRegistry, error) {
	if !strings.HasPrefix(url, "http") {
		url = "https://" + url
	}
	// check type here! based on URL. can pull URL settings from settings or something
	var err error
	regtype := checkRegistryType(url)
	if regtype == "containeryard" {
		yard, err := NewContainerYardRegistry(url, org, username, password)
		if err == nil {
			if yard != nil && yard.Ping() == nil {
				return yard, nil
			}
			err = yard.Ping()
		}
	}

	if regtype == "v2" {
		v2, err := NewV2Registry(url, org, username, password)
		if err == nil {
			if v2 != nil && v2.Ping() == nil {
				return v2, nil
			}
			err = v2.Ping()
		}
	}

	if regtype == "dockerhub" {
		hub, err := NewDockerHubRegistry(url, org, username, password)
		if err == nil {
			if hub != nil && hub.Ping() == nil {
				return hub, nil
			}
			err = hub.Ping()
		}
	}

	if regtype == "gitlab" {
		git, err := NewGitLabRegistry(url, org, username, password)
		if err == nil {
			if git != nil && git.Ping() == nil {
				return git, nil
			}
			err = git.Ping()
		}
	}

	msg := fmt.Sprintf("ERROR: Could not create registry.  %s: %s", regtype, err.Error())
	errr := errors.New(msg)

	return nil, errr
}

func checkRegistryType(url string) string {
	if strings.Contains(url, "hub.docker.com") {
		return "dockerhub"
	}
	if strings.Contains(url, "containeryard") {
		return "containeryard"
	}
	if strings.Contains(url, "gitlab") {
		return "gitlab"
	}
	return "v2"
}
