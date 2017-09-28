package v2

import (
	"strings"

	"github.com/heroku/docker-registry-client/registry"
	"github.com/ngageoint/seed-cli/util"
)

type v2registry struct {
	r *registry.Registry
	Print  util.PrintCallback
}

func New(url, username, password string) (*v2registry, error) {
	reg, err := registry.New(url, username, password)
	if reg != nil {
		return &v2registry{r: reg, Print: util.PrintUtil}, err
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
	r.Print( "Searching %s for Seed images...\n", url)
	repositories, err := r.r.Repositories()

	var images []string
	for _, repo := range repositories {
		if !strings.HasSuffix(repo, "-seed") {
			continue
		}
		tags, err := r.Tags(repo, org)
		if err != nil {
			print( err.Error())
			continue
		}
		for _, tag := range tags {
			images = append(images, repo+":"+tag)
		}
	}

	return images, err
}
