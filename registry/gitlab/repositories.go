package gitlab

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ngageoint/seed-common/objects"
)

//Result struct representing JSON result
type repositoriesResponse struct {
	Results []Repository
}

//Repository struct representing a GitLab repository
type Repository struct {
	ID        int
	Name      string
	Path      string
	ProjectID int
	Location  string
	CreatedAt string
	Tags      []Tag
}

//Tag struct representing the GitLab tag structure
type Tag struct {
	Name     string
	Path     string
	Location string
}

//Repositories Returns the seed repositories for the given group/org/project
func (registry *GitLabRegistry) Repositories() ([]string, error) {
	var repo string
	if strings.TrimSpace(registry.Org) != "" && strings.TrimSpace(registry.Path) != "" {
		repo = fmt.Sprintf("projects/%s%%2F%s", registry.Org, registry.Path)
	} else if strings.TrimSpace(registry.Org) != "" {
		repo = fmt.Sprintf("groups/%s", registry.Org)
	} else {
		repo = fmt.Sprintf("projects/%s", strings.Replace(registry.Path, "/", "%2F", -1))
	}

	url := registry.url("/api/v4/%s/registry/repositories", repo)
	repos := make([]string, 0, 10)
	var err error //We create this here, otherwise url will be rescoped with :=
	var response []Repository

	err = registry.getGitLabJson(url, &response)
	if err == nil {
		for _, r := range response {
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

	var reg string
	if strings.TrimSpace(registry.Org) != "" && strings.TrimSpace(registry.Path) != "" {
		reg = fmt.Sprintf("projects/%s%%2F%s", registry.Org, registry.Path)
	} else if strings.TrimSpace(registry.Org) != "" {
		reg = fmt.Sprintf("groups/%s", registry.Org)
	} else {
		reg = fmt.Sprintf("projects/%s", strings.Replace(registry.Path, "/", "%2F", -1))
	}

	// Need to find the id of the specific repository
	// var repository string
	repo, err := registry.GetRepositoryInfo(repository)
	if err != nil {
		return nil, err
	}

	url := registry.url("/api/v4/%s/registry/repositories/%d/tags", reg, repo.ID)
	tags := make([]string, 0, 10)
	var response []Tag

	err = registry.getGitLabJson(url, &response)
	if err == nil {
		for _, r := range response {
			tags = append(tags, r.Name)
		}
	}

	return tags, err
}

//Images returns all seed images on the registry
func (registry *GitLabRegistry) Images() ([]string, error) {
	var reg string
	if strings.TrimSpace(registry.Org) != "" && strings.TrimSpace(registry.Path) != "" {
		reg = fmt.Sprintf("projects/%s%%2F%s", registry.Org, registry.Path)
	} else if strings.TrimSpace(registry.Org) != "" {
		reg = fmt.Sprintf("groups/%s", registry.Org)
	} else {
		reg = fmt.Sprintf("projects/%s", strings.Replace(registry.Path, "/", "%2F", -1))
	}

	url := registry.url("/api/v4/%s/registry/repositories/?tags=true", reg)
	var response repositoriesResponse
	err := registry.getGitLabJson(url, &response.Results)
	repos := []string{}
	for _, r := range response.Results {
		if !strings.HasSuffix(r.Name, "-seed") {
			continue
		}
		if len(r.Tags) > 0 {
			for _, t := range r.Tags {

				registry.Tags(r.Name)

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

	if registry.Org != "" && registry.Path != "" {
		org = registry.Org + "/" + registry.Path
	} else if registry.Org != "" {
		org = registry.Org
	} else {
		org = registry.Path
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
	fullRepo := repoName
	if strings.TrimSpace(registry.Org) != "" && strings.TrimSpace(registry.Path) != "" {
		fullRepo = fmt.Sprintf("%s/%s/%s", registry.Org, registry.Path, repoName)
	} else if strings.TrimSpace(registry.Org) != "" {
		fullRepo = fmt.Sprintf("%s/%s", registry.Org, repoName)
	} else if strings.TrimSpace(registry.Path) != "" {
		fullRepo = fmt.Sprintf("%s/%s", registry.Path, repoName)
	}

	manifest := ""

	mv2, err := registry.v2Base.ManifestV2(fullRepo, tag)
	if err == nil {
		resp, err := registry.v2Base.DownloadLayer(fullRepo, mv2.Config.Digest)
		if err == nil {
			manifest, err = objects.GetSeedManifestFromBlob(resp)
		}
	}

	if err == nil && manifest == "" {
		err = errors.New("Empty seed manifest!")
	}

	return manifest, err
}

//GetRepositoryInfo returns the id for a given repository located in the GitLab registry
func (registry *GitLabRegistry) GetRepositoryInfo(repository string) (*Repository, error) {

	var repo string
	if strings.TrimSpace(registry.Org) != "" && strings.TrimSpace(registry.Path) != "" {
		repo = fmt.Sprintf("projects/%s%%2F%s", registry.Org, registry.Path)
	} else if strings.TrimSpace(registry.Org) != "" {
		repo = fmt.Sprintf("groups/%s", registry.Org)
	} else {
		repo = fmt.Sprintf("projects/%s", strings.Replace(registry.Path, "/", "%2F", -1))
	}

	url := registry.url("/api/v4/%s/registry/repositories", repo)
	var err error //We create this here, otherwise url will be rescoped with :=
	var response []Repository

	err = registry.getGitLabJson(url, &response)
	if err == nil {
		for _, r := range response {
			if r.Name == repository {
				return &r, err
			}
		}
	}

	return nil, err
}

//ExtractOrgPath extracts the group and path portions of a GitLab registry Org field
func ExtractOrgPath(url, org, token string) (group, path string, err error) {
	orgParts := strings.Split(org, "/")
	if len(orgParts) >= 1 {
		group = orgParts[0]

		//Try and see if the first part is an organization you have access to
		fullURL := fmt.Sprintf("%s/api/v4/groups/%s", url, group)
		req, err := http.NewRequest("GET", fullURL, nil)
		req.Header.Add("PRIVATE-TOKEN", token)
		client := &http.Client{}
		resp, err := client.Do(req)

		if err != nil {
			return "", org, err
		}
		defer resp.Body.Close()

		// Not a group - or just don't have access
		if resp.StatusCode == 404 {
			group = ""
			path = org
		} else {
			path = strings.TrimPrefix(org, group)
			path = strings.TrimPrefix(path, "/")
		}
	}
	return group, path, nil
}
