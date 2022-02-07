package gitlab

import (
	"errors"
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
	Path     string
	Username string
	Password string
	v2Base   *v2.V2registry
	Print    util.PrintCallback
}

func (r *GitLabRegistry) Name() string {
	return "GitLabRegistry"
}

//New creates a new gitlab container registry from the given URL
func New(registryUrl, org, path, username, password string) (*GitLabRegistry, error) {
	if util.PrintUtil == nil {
		util.InitPrinter(util.PrintErr, os.Stderr, os.Stdout)
	}

	url := strings.TrimSuffix(registryUrl, "/")
	path = strings.TrimSuffix(path, "/")

	v2org := org
	if strings.TrimSpace(org) != "" && strings.TrimSpace(path) != "" {
		v2org = fmt.Sprintf("%s/%s", org, path)
	} else if strings.TrimSpace(path) != "" {
		v2org = path
	}

	host := strings.Replace(url, "https://", "", 1)
	host = strings.Replace(host, "http://", "", 1)

	client := &http.Client{}

	registry := &GitLabRegistry{
		URL:      url,
		Hostname: host,
		Client:   client,
		Org:      org,
		Path:     path,
		Username: username,
		Password: password,
		Print:    util.PrintUtil,
	}

	// Need to set the v2 base url as the location url - not just the registry url
	// due to differences in accessing the gitlab API vs the docker v2 API
	location, err := registry.GetRegistryLocation()
	reg, err := v2.New(location, v2org, username, password)
	registry.v2Base = reg

	return registry, err
}

//GetRegistryLocation returns the location for a given repository located in the GitLab registry
func (r *GitLabRegistry) GetRegistryLocation() (string, error) {

	var repo string
	if strings.TrimSpace(r.Org) != "" && strings.TrimSpace(r.Path) != "" {
		repo = fmt.Sprintf("projects/%s%%2F%s", r.Org, r.Path)
	} else if strings.TrimSpace(r.Org) != "" {
		repo = fmt.Sprintf("groups/%s", r.Org)
	} else {
		repo = fmt.Sprintf("projects/%s", strings.Replace(r.Path, "/", "%2F", -1))
	}

	// List the repositories for the registry so we can grab the location field
	url := r.url("/api/v4/%s/registry/repositories?per_page=200", repo)
	var response []Repository

	err := r.getGitLabJson(url, &response)
	if err == nil {
		// Get the first result and grab the location field
		if len(response) > 0 {
			location := strings.TrimSuffix(response[0].Location, response[0].Path)
			location = strings.TrimSuffix(location, "/")
			scheme := strings.TrimSuffix(r.URL, r.Hostname)
			location = fmt.Sprintf("%s%s", scheme, location)
			return location, nil
		}
	}

	// Default to returning the url, but it's likely there's an error
	return r.URL, err
}

//url Returns the full URL for the given path template
func (r *GitLabRegistry) url(pathTemplate string, args ...interface{}) string {
	pathSuffix := fmt.Sprintf(pathTemplate, args...)
	url := fmt.Sprintf("%s%s", r.URL, pathSuffix)
	return url
}

//Ping Verifies the registry is alive and callable
func (r *GitLabRegistry) Ping() error {

	//query that should quickly return not an error
	url := r.url("/api/v4/groups/")

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("PRIVATE-TOKEN", r.Password)
	resp, err := r.Client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return errors.New(resp.Status)
		}
	}

	return err
}
