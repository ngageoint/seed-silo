package dockerhub

import (
	"errors"
	"strings"

	"github.com/ngageoint/seed-common/objects"
)

type repositoriesResponse struct {
	Count    int
	Next     string
	Previous string
	Results  []Result
}

//Result struct representing JSON result
type Result struct {
	Name string
}

//Repositories Returns seed repositories for the given user/organization
func (registry *DockerHubRegistry) Repositories() ([]string, error) {
	user := registry.Org
	url := registry.url("/v2/repositories/%s/", user)
	repos := make([]string, 0, 10)
	var err error //We create this here, otherwise url will be rescoped with :=
	var response repositoriesResponse
	for err == nil {
		response.Next = ""
		url, err = registry.getDockerHubPaginatedJson(url, &response)
		for _, r := range response.Results {
			if !strings.HasSuffix(r.Name, "-seed") {
				continue
			}
			repos = append(repos, r.Name)
		}
	}
	if err != ErrNoMorePages {
		return nil, err
	}
	return repos, nil
}

//Tags Returns tags for a given user/organization and repository
func (registry *DockerHubRegistry) Tags(repository string) ([]string, error) {
	user := registry.Org
	url := registry.url("/v2/repositories/%s/%s/tags", user, repository)
	tags := make([]string, 0, 10)
	var err error //We create this here, otherwise url will be rescoped with :=
	var response repositoriesResponse
	for err == nil {
		response.Next = ""
		url, err = registry.getDockerHubPaginatedJson(url, &response)
		for _, r := range response.Results {
			tags = append(tags, r.Name)
		}
	}
	if err != ErrNoMorePages {
		return nil, err
	}
	return tags, nil
}

//Images returns seed images for a given user/repository.  It will grab all of the seed repositories and combine them
//with any tags it can find to build a list of images.
func (registry *DockerHubRegistry) Images() ([]string, error) {
	url := registry.url("/v2/repositories/%s/", registry.Org)
	registry.Print("Searching %s for Seed images...\n", url)
	repos := make([]string, 0, 10)
	var err error //We create this here, otherwise url will be rescoped with :=
	var response repositoriesResponse
	for err == nil {
		response.Next = ""
		url, err = registry.getDockerHubPaginatedJson(url, &response)
		for _, r := range response.Results {
			if !strings.HasSuffix(r.Name, "-seed") {
				continue
			}
			// Add all tags if found
			if rs, _ := registry.Tags(r.Name); len(rs) > 0 {
				for _, tag := range rs {
					img := r.Name + ":" + tag
					repos = append(repos, img)
				}
				// No tags found - so just add the repo name
			} else {
				repos = append(repos, r.Name)
			}
		}
	}
	if err != ErrNoMorePages {
		return nil, err
	}
	return repos, nil
}

func (registry *DockerHubRegistry) ImagesWithManifests() ([]objects.Image, error) {
	imageNames, err := registry.Images()

	if err != nil {
		return nil, err
	}

	images := []objects.Image{}

	url := "docker.io"

	manifest := ""
	for _, imgstr := range imageNames {
		temp := strings.Split(imgstr, ":")
		if len(temp) != 2 {
			registry.Print("ERROR: Invalid seed name: %s. Unable to split into name/tag pair\n", imgstr)
			continue
		}
		manifest, err = registry.GetImageManifest(temp[0], temp[1])

		imageStruct := objects.Image{Name: imgstr, Registry: url, Org: registry.Org, Manifest: manifest}
		images = append(images, imageStruct)
	}

	return images, err
}

func (registry *DockerHubRegistry) GetImageManifest(repoName, tag string) (string, error) {
	manifest := ""
	orgRepoName := registry.Org + "/" + repoName
	mv2, err := registry.v2Base.ManifestV2(orgRepoName, tag)
	if err == nil {
		resp, err := registry.v2Base.DownloadLayer(orgRepoName, mv2.Config.Digest)
		if err == nil {
			manifest, err = objects.GetSeedManifestFromBlob(resp)
		}
	}

	if err == nil && manifest == "" {
		err = errors.New("Empty seed manifest!")
	}

	return manifest, err
}
