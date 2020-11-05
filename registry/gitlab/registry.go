package gitlab

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/ngageoint/seed-common/util"
	"github.com/ngageoint/seed-silo/registry/v2"
)

//GitLabRegistry type representing a GitLab registry
type GitLabRegistry struct {
	URL      string
	Hostname string
	Client   *http.Client
	Org      string
	Project  string
	Username string
	Password string
	v2Base   *v2.V2registry
	Print    util.PrintCallback
}

func (r *GitLabRegistry) Name() string {
	return "GitLabRegistry"
}

//New creates a new gitlab container registry from the given URL
func New(registryUrl, org, project, username, password string) (*GitLabRegistry, error) {
	if util.PrintUtil == nil {
		util.InitPrinter(util.PrintErr, os.Stderr, os.Stdout)
	}

	url := strings.TrimSuffix(registryUrl, "/")

	org_proj := []string{org, project}
	if strings.TrimSpace(org_proj) != "" {
		org = strings.Join(org_proj, "/")
	}

	reg, err := v2.New(url, org, username, password)

	host := strings.Replace(url, "https://", "", 1)
	host = strings.Replace(host, "http://", "", 1)

	client := &http.Client{}

	registry := &GitLabRegistry{
		URL:      url,
		Hostname: host,
		Client:   client,
		Org:      org,
		Project:  project,
		Username: username,
		Password: password,
		v2Base:   reg,
		Print:    util.PrintUtil,
	}

	return registry, err
}

func (r *GitLabRegistry) url(pathTemplate string, args ...interface{}) string {
	pathSuffix := fmt.Sprintf(pathTemplate, args...)
	url := fmt.Sprintf("%s%s", r.URL, pathSuffix)
	return url
}

func (r *GitLabRegistry) Ping() error {
	//query that should quickly return an empty json response
	url := r.url("/search?q=NoImagesWithThisName&t=json")
	var response Response
	err := r.getGitLabJson(url, &response)
	return err
}
