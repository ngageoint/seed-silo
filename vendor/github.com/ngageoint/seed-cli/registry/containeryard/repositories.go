package containeryard

import "github.com/ngageoint/seed-cli/objects"

type Response struct {
	Results Results
}

type Results struct {
	Community map[string]*Image
	Imports   map[string]*Image
}

type Image struct {
	Author    string
	Compliant bool
	Error     bool
	Labels    map[string]string
	Obsolete  bool
	Pulls     string
	Stars     int
	Tags      map[string]Tag
}

type Tag struct {
	Age     int
	Created string
	Digest  string
	Size    string
}

//Result struct representing JSON result
type Result struct {
	Name string
}

func (registry *ContainerYardRegistry) Repositories(org string) ([]string, error) {
	url := registry.url("/search?q=%s&t=json", "-seed")
	repos := make([]string, 0, 10)
	var err error //We create this here, otherwise url will be rescoped with :=
	var response Response

	err = registry.getContainerYardJson(url, &response)
	if err == nil {
		for repoName, _ := range response.Results.Community {
			repos = append(repos, repoName)
		}
		for repoName, _ := range response.Results.Imports {
			repos = append(repos, repoName)
		}
	}
	return repos, err

}

func (registry *ContainerYardRegistry) Tags(repository, org string) ([]string, error) {
	url := registry.url("/search?q=%s&t=json", repository)
	registry.Print( "Searching %s for Seed images...\n", url)
	tags := make([]string, 0, 10)
	var err error //We create this here, otherwise url will be rescoped with :=
	var response Response

	err = registry.getContainerYardJson(url, &response)
	if err == nil {
		for _, image := range response.Results.Community {
			for tagName, _ := range image.Tags {
				tags = append(tags, tagName)
			}
		}
		for _, image := range response.Results.Imports {
			for tagName, _ := range image.Tags {
				tags = append(tags, tagName)
			}
		}
	}
	return tags, err
}

//Images returns all seed images on the registry
func (registry *ContainerYardRegistry) Images(org string) ([]string, error) {
	images, err := registry.ImagesWithManifests(org)
	imageStrs := []string{}
	for _, img := range images {
		imageStrs = append(imageStrs, img.Name)
	}
	return imageStrs, err
}

//Images returns all seed images on the registry along with their manifests, if available
func (registry *ContainerYardRegistry) ImagesWithManifests(org string) ([]objects.Image, error) {
	//TODO: Update after container yard generates unique manifests for each tag
	url := registry.url("/search?q=%s&t=json", "-seed")
	repos := make([]objects.Image, 0, 10)
	var err error //We create this here, otherwise url will be rescoped with :=
	var response Response

	err = registry.getContainerYardJson(url, &response)
	if err == nil {
		for repoName, image := range response.Results.Community {
			manifestLabel := ""
			for name, value := range image.Labels {
				if name == "com.ngageoint.seed.manifest" {
					manifestLabel = value
				}
			}
			for tagName, _ := range image.Tags {
				imageStr := repoName + ":" + tagName
				img := objects.Image{Name: imageStr, Registry: registry.URL, Org: org, Manifest: manifestLabel}
				repos = append(repos, img)
			}
		}
		for repoName, image := range response.Results.Imports {
			manifestLabel := ""
			for name, value := range image.Labels {
				if name == "com.ngageoint.seed.manifest" {
					manifestLabel = value
				}
			}
			for tagName, _ := range image.Tags {
				imageStr := repoName + ":" + tagName
				img := objects.Image{Name: imageStr, Registry: registry.URL, Org: org, Manifest: manifestLabel}
				repos = append(repos, img)
			}
		}
	}
	return repos, nil
}
