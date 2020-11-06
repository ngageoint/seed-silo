package gitlab

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ngageoint/seed-common/objects"
)

//Result struct representing JSON result
type repositoriesResponse struct {
	Results []Repository
}

type Repository struct {
	Id        int
	Name      string
	Path      string
	ProjectId int
	Location  string
	CreatedAt string
	Tags      []Tag
}

type tagsResponse struct {
	Results []Tag
}

type Tag struct {
	Name     string
	Path     string
	Location string
}

//Repositories Returns the seed repositories for the given group/org/project
func (registry *GitLabRegistry) Repositories() ([]string, error) {
	org := registry.Org
	project := registry.Project

	var repo string
	if strings.TrimSpace(org) != "" && strings.TrimSpace(project) != "" {

		repo := fmt.Sprintf("projects/%s%%2F%s", registry.Org, registry.Project)
	} else if strings.TrimSpace(org) != "" {
		repo := fmt.Sprintf("groups/%s", registry.Org)
	} else {
		repo := fmt.Sprintf("projects/%s", strings.ReplaceAll(registry.Project, "/", "%2F"))
	}

	url := registry.url("/api/v4/%s/registry/repositories", repo)
	repos := make([]string, 0, 10)
	var err error //We create this here, otherwise url will be rescoped with :=
	var response repositoriesResponse

	err = registry.getGitLabJson(url, &response)
	if err == nil {
		for _, r := range response.Results {
			if !strings.HasSuffix(r.Name, "-seed") {
				continue
			}
			repos = append(repos, r.Name)
		}
	}
	return repos, err
}

//Tags returns the tags for a specific gitlab registry
func (registry *GitLabRegistry) Tags(repository string) ([]string, error) {

	org := registry.Org
	project := registry.Project

	var reg string
	if strings.TrimSpace(org) != "" && strings.TrimSpace(project) != "" {
		reg := fmt.Sprintf("projects/%s%%2F%s", registry.Org, registry.Project)
	} else if strings.TrimSpace(org) != "" {
		reg := fmt.Sprintf("groups/%s", registry.Org)
	} else {
		reg := fmt.Sprintf("projects/%s", strings.ReplaceAll(registry.Project, "/", "%2F"))
	}

	registry.Print("Searching %s for Seed images...\n", url)
	// Need to find the id of the specific repository
	var respository string

	url := registry.url("/api/v4/%s/registry/repositories/%s/tags", reg, repository)
	tags := make([]string, 0, 10)
	var err error //We create this here, otherwise url will be rescoped with :=
	var response tagsResponse

	err = registry.getGitLabJson(url, &response)
	if err == nil {
		for _, r := range response.Results {
			tags = append(tags, r.Name)
		}
	}
	return tags, err
}

//Images returns all seed images on the registry
func (registry *GitLabRegistry) Images() ([]string, error) {

	org := registry.Org
	project := registry.Project

	var reg string
	if strings.TrimSpace(org) != "" && strings.TrimSpace(project) != "" {
		reg := fmt.Sprintf("projects/%s%%2F%s", registry.Org, registry.Project)
	} else if strings.TrimSpace(org) != "" {
		reg := fmt.Sprintf("groups/%s", registry.Org)
	} else {
		reg := fmt.Sprintf("projects/%s", strings.ReplaceAll(registry.Project, "/", "%2F"))
	}

	url := registry.url("/api/v4/%s/registry/repositories/?tags=true", reg)
	var response repositoriesResponse
	err := registry.getGitLabJson(url, &response)
	repos := []string{}

	for _, r := range response.Results {
		if !strings.HasSuffix(r.Name, "-seed") {
			continue
		}
		if len(r.Tags) > 0 {
			for _, t := range r.Tags {
				img := r.Name + ":" + t.Name
				repos = append(repos, img)
			}
		} else {
			repos = append(repos, r.Name)
		}

	}

	if err != nil {
		return nil, err
	}

	return repos, err
}

//ImagesWithManifests returns all seed images on the registry along with their manifests, if available
func (registry *GitLabRegistry) ImagesWithManifests() ([]objects.Image, error) {
	imageNames, err := registry.Images()

	images := []objects.Image{}

	url := registry.URL
	var org string

	if registry.Org != "" && registry.Project != "" {
		org := registry.Org + "/" + registry.Project
	} else if registry.Org != "" {
		org := registry.Org
	} else {
		org := registry.Project
	}

	manifest := ""
	for _, imgstr := range imageNames {
		temp := strings.Split(imgstr, ":")
		if len(temp) != 2 {
			registry.Print("ERROR: Invalid seed name: %s. Unable to split into name/tag pair\n", imgstr)
			continue
		}
		manifest, err = registry.GetImageManifest(temp[0], temp[1])
		imageStruct := objects.Image{Name: imgstr, Registry: url, Org: org, Manifest: manifest}
		images = append(images, imageStruct)
	}

	return images, err
}

//GetImageManifest returns the image manifest from a gitlab repo
func (registry *GitLabRegistry) GetImageManifest(repoName, tag string) (string, error) {
	manifest := ""
	mv2, err := registry.v2Base.ManifestV2(repoName, tag)
	if err == nil {
		resp, err := registry.v2Base.DownloadLayer(repoName, mv2.Config.Digest)
		if err == nil {
			manifest, err = objects.GetSeedManifestFromBlob(resp)
		}
	}

	if err == nil && manifest == "" {
		err = errors.New("Empty seed manifest!")
	}

	return manifest, err
}
