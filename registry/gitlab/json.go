package gitlab

import (
	"encoding/json"
	"net/http"
)

//getGitLabJson Returns the JSON response to the GitLab call
func (registry *GitLabRegistry) getGitLabJson(url string, response interface{}) error {

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("PRIVATE-TOKEN", registry.Password)
	client := &http.Client{}
	resp, err := client.Do(req)

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
