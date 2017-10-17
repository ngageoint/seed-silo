package v2

import (
	"strings"

	"github.com/heroku/docker-registry-client/registry"
	"github.com/ngageoint/seed-cli/objects"
	"github.com/ngageoint/seed-cli/util"
)

type v2registry struct {
	r        *registry.Registry
	Username string
	Password string
	Print    util.PrintCallback
}

func New(url, username, password string) (*v2registry, error) {
	if util.PrintUtil == nil {
		util.InitPrinter(false)
	}

	reg, err := registry.New(url, username, password)
	if reg != nil {
		return &v2registry{r: reg, Username: username, Password: password, Print: util.PrintUtil}, err
	}
	return nil, err
}

func (r *v2registry) Name() string {
	return "V2"
}

func (r *v2registry) Ping() error {
	_, err := r.r.Repositories()
	return err
}

func (r *v2registry) Repositories(org string) ([]string, error) {
	return r.r.Repositories()
}

func (r *v2registry) Tags(repository, org string) ([]string, error) {
	return r.r.Tags(repository)
}

func (r *v2registry) Images(org string) ([]string, error) {
	url := r.r.URL + "/v2/_catalog"
	r.Print("Searching %s for Seed images...\n", url)
	repositories, err := r.r.Repositories()

	var images []string
	for _, repo := range repositories {
		if !strings.HasSuffix(repo, "-seed") {
			continue
		}
		tags, err := r.Tags(repo, org)
		if err != nil {
			r.Print(err.Error())
			continue
		}
		for _, tag := range tags {
			images = append(images, repo+":"+tag)
		}
	}

	return images, err
}

func (r *v2registry) ImagesWithManifests(org string) ([]objects.Image, error) {
	imageNames, err := r.Images(org)

	if err != nil {
		return nil, err
	}

	images := []objects.Image{}

	url := strings.Replace(r.r.URL, "http://", "", 1)
	url = strings.Replace(url, "https://", "", 1)
	username := r.Username
	password := r.Password

	for _, imgstr := range imageNames {
		manifest := ""
		//TODO: find better, lightweight way to get manifest on low side
		imageName, err := util.DockerPull(imgstr, url, org, username, password)
		if err == nil {
			manifest, err = util.GetSeedManifestFromImage(imageName)
		}
		if err != nil {
			r.Print("ERROR: Could not get manifest: %s\n", err.Error())
		}

		imageStruct := objects.Image{Name: imgstr, Registry: url, Org: org, Manifest: manifest}
		images = append(images, imageStruct)
	}

	return images, err
}
