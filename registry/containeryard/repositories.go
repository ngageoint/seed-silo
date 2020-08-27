package containeryard

import (
	"errors"
	"strings"

	"github.com/ngageoint/seed-common/objects"
	"github.com/ngageoint/seed-common/util"
)

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

func (registry *ContainerYardRegistry) Repositories() ([]string, error) {
	url := registry.url("/search?q=%s&t=json", "-seed")
	repos := make([]string, 0, 10)
	var err error //We create this here, otherwise url will be rescoped with :=
	var response Response

	err = registry.getContainerYardJson(url, &response)
	if err == nil {
		for repoName := range response.Results.Community {
			repos = append(repos, repoName)
		}
		for repoName := range response.Results.Imports {
			repos = append(repos, repoName)
		}
	}
	return repos, err

}

func (registry *ContainerYardRegistry) Tags(repository string) ([]string, error) {
	url := registry.url("/search?q=%s&t=json", repository)
	registry.Print("Searching %s for Seed images...\n", url)
	tags := make([]string, 0, 10)
	var err error //We create this here, otherwise url will be rescoped with :=
	var response Response

	err = registry.getContainerYardJson(url, &response)
	if err == nil {
		for _, image := range response.Results.Community {
			for tagName := range image.Tags {
				tags = append(tags, tagName)
			}
		}
		for _, image := range response.Results.Imports {
			for tagName := range image.Tags {
				tags = append(tags, tagName)
			}
		}
	}
	return tags, err
}

//Images returns all seed images on the registry
func (registry *ContainerYardRegistry) Images() ([]string, error) {
	images, err := registry.ImagesWithManifests()
	imageStrs := []string{}
	for _, img := range images {
		imageStrs = append(imageStrs, img.Name)
	}
	return imageStrs, err
}

//Images returns all seed images on the registry along with their manifests, if available
func (registry *ContainerYardRegistry) ImagesWithManifests() ([]objects.Image, error) {
	//TODO: Update after container yard generates unique manifests for each tag
	url := registry.url("/search?q=%s&t=json", "-seed")
	repos := make([]objects.Image, 0, 10)
	var err error //We create this here, otherwise url will be rescoped with :=
	var response Response

	err = registry.getContainerYardJson(url, &response)
	if err == nil {
		for repoName, image := range response.Results.Community {
			if !strings.HasPrefix(repoName, registry.Org) {
				registry.Print("Skipping image %s because it does not belong to org %s", repoName, registry.Org)
				continue
			}
			manifestLabel := ""
			for name, value := range image.Labels {
				if name == "com.ngageoint.seed.manifest" {
					manifestLabel = util.UnescapeManifestLabel(value)
				}
			}
			if manifestLabel == "" {
				registry.Print("Skipping image %s due to missing manifest label", repoName)
				continue
			}
			for tagName := range image.Tags {
				manifestLabel, err = registry.GetImageManifest(repoName, tagName)
				if err != nil {
					//skip images with empty manifests
					registry.Print("ERROR: Error reading v2 manifest for %s: %s\n Skipping.\n", repoName, err.Error())
					continue
				}
				imageStr := repoName + ":" + tagName
				org := registry.Org
				parts := strings.SplitN(imageStr, "/", 2)
				if len(parts) == 2 {
					org = parts[0]
					imageStr = parts[1]
				} else {
					registry.Print("Error parsing org out of repo name: %s \n", repoName)
				}
				img := objects.Image{Name: imageStr, Registry: registry.Hostname, Org: org, Manifest: manifestLabel}
				repos = append(repos, img)
			}
		}
		for repoName, image := range response.Results.Imports {
			if !strings.HasPrefix(repoName, registry.Org) {
				registry.Print("Skipping image %s because it does not belong to org %s", repoName, registry.Org)
				continue
			}
			manifestLabel := ""
			for name, value := range image.Labels {
				if name == "com.ngageoint.seed.manifest" {
					manifestLabel = util.UnescapeManifestLabel(value)
				}
			}
			if manifestLabel == "" {
				registry.Print("Skipping image %s due to missing manifest label", repoName)
				continue
			}
			for tagName := range image.Tags {
				manifestLabel, err = registry.GetImageManifest(repoName, tagName)
				if err != nil {
					//skip images with empty manifests
					registry.Print("ERROR: Error reading v2 manifest for %s: %s\n Skipping.\n", repoName, err.Error())
					continue
				}
				imageStr := repoName + ":" + tagName
				org := registry.Org
				parts := strings.SplitN(imageStr, "/", 2)
				if len(parts) == 2 {
					org = parts[0]
					imageStr = parts[1]
				} else {
					registry.Print("Error parsing org out of repo name: %s \n", repoName)
				}
				img := objects.Image{Name: imageStr, Registry: registry.Hostname, Org: org, Manifest: manifestLabel}
				repos = append(repos, img)
			}
		}
	}
	return repos, nil
}

func (registry *ContainerYardRegistry) GetImageManifest(repoName, tag string) (string, error) {
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
