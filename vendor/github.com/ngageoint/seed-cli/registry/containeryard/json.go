package containeryard

import (
	"encoding/json"
	"net/http"
)

// getContainerYardJson works with the list of repositories returned by container yard
func (registry *ContainerYardRegistry) getContainerYardJson(url string, response interface{}) error {
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
