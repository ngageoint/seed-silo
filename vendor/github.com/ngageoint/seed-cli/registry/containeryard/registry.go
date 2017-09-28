package containeryard

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ngageoint/seed-cli/util"
)

//ContainerYardRegistry type representing a Container Yard registry
type ContainerYardRegistry struct {
	URL    string
	Client *http.Client
	Print  util.PrintCallback
}

func (r *ContainerYardRegistry) Name() string {
	return "ContainerYardRegistry"
}

//New creates a new docker hub registry from the given URL
func New(registryUrl string) (*ContainerYardRegistry, error) {
	url := strings.TrimSuffix(registryUrl, "/")
	registry := &ContainerYardRegistry{
		URL:    url,
		Client: &http.Client{},
		Print: util.PrintUtil,
	}

	return registry, nil
}

func (r *ContainerYardRegistry) url(pathTemplate string, args ...interface{}) string {
	pathSuffix := fmt.Sprintf(pathTemplate, args...)
	url := fmt.Sprintf("%s%s", r.URL, pathSuffix)
	return url
}

func (r *ContainerYardRegistry) Ping() error {
	//query that should quickly return an empty json response
	url := r.url("/search?q=NoImagesWithThisName&t=json")
	var response Response
	err := r.getContainerYardJson(url, &response)
	return err
}
