package gitlab

import (
	"encoding/json"
	"net/http"
)

//getGitLabJson returns
func (registry *GitLabRegistry) getGitLabJson(url string, response interface{}) error {

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(response)
	if err != nil {
		return err
	}
	return err
}
