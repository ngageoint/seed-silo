package dockerhub

import (
	"encoding/json"
	"errors"
	"net/http"
)

var (
	//ErrNoMorePages error representing no more pages
	ErrNoMorePages = errors.New("No more pages")
)

// getDockerHubPaginatedJson works with the list of repositories for a user
// returned by docker hub. accepts a string and a pointer, and returns the
// next page URL while updating pointed-to variable with a parsed JSON
// value. When there are no more pages it returns `ErrNoMorePages`.
func (registry *DockerHubRegistry) getDockerHubPaginatedJson(url string, response interface{}) (string, error) {

	req, err := http.NewRequest("GET", url, nil)
	resp, err := registry.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(response)
	r := response.(*repositoriesResponse)
	if err != nil {
		registry.Print("Error retrieving url %s: %s\n", url, err.Error())
		return "", err
	}
	if r.Next == "" {
		err = ErrNoMorePages
	}
	return r.Next, err
}
