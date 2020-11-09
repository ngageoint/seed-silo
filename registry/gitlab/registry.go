package gitlab

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/ngageoint/seed-common/constants"
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
	Token    string
	v2Base   *v2.V2registry
	Print    util.PrintCallback
}

func (r *GitLabRegistry) Name() string {
	return "GitLabRegistry"
}

//New creates a new gitlab container registry from the given URL
func New(registryUrl, org, project, token string) (*GitLabRegistry, error) {
	if util.PrintUtil == nil {
		util.InitPrinter(util.PrintErr, os.Stderr, os.Stdout)
	}

	url := strings.TrimSuffix(registryUrl, "/")
	project = strings.TrimSuffix(project, "/")

	if strings.TrimSpace(org) != "" && strings.TrimSpace(project) != "" {
		org = fmt.Sprintf("%s/%s", org, project)
	} else if strings.TrimSpace(project) != "" {
		org = project
	}

	reg, err := v2.New(url, org, "", token)

	host := strings.Replace(url, "https://", "", 1)
	host = strings.Replace(host, "http://", "", 1)

	client := &http.Client{}

	registry := &GitLabRegistry{
		URL:      url,
		Hostname: host,
		Client:   client,
		Org:      org,
		Project:  project,
		Token:    token,
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

//Ping Verifies the registry is alive
func (r *GitLabRegistry) Ping() error {
	//query that should quickly return an empty json response

	url := r.url("/v4/groups/%s/registry/repositories", constants.DefaultOrg)
	resp, err := r.Client.Get(url)
	if resp != nil {
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return errors.New(resp.Status)
		}
	}
	return err
}
