package v2

import (
	"encoding/json"

	"net/http"
	"net/url"
	// "strconv"
	"time"

	// manifestV1 "github.com/docker/distribution/manifest/schema1"
	manifestV2 "github.com/docker/distribution/manifest/schema2"
)

type APIToken struct {
	// Token   authToken
	Created               time.Time
	Expires               time.Time
	realm, service, scope string
	Token                 string  `json:"token"`
	ExpiresIn             float64 `json:"expires_in"`
}

func (registry *V2registry) GetOrCreateToken(org string, urlstring string) (APIToken, error) {

	apiToken := APIToken{}
	if !Authtokenmap().Check(org) {
		//token doesent exist yet
		req, err := http.NewRequest("GET", urlstring, nil)

		if err != nil {
			return apiToken, err
		}

		req.Header.Set("Accept", manifestV2.MediaTypeManifest)
		resp, err := registry.Client.Do(req)

		if err != nil {
			return apiToken, err
		}

		if resp.StatusCode == http.StatusUnauthorized {
			// var service, scope, realm string

			challenges := parseAuthHeader(resp.Header)
			for _, challenge := range challenges {
				if challenge.Scheme == "bearer" {
					apiToken.realm = challenge.Parameters["realm"]
					apiToken.service = challenge.Parameters["service"]
					apiToken.scope = challenge.Parameters["scope"]
					break
				}
			}

			paresedURL, err := url.Parse(apiToken.realm)

			if err != nil {
				return apiToken, err
			}

			q := paresedURL.Query()
			q.Set("service", apiToken.service)
			if apiToken.scope != "" {
				q.Set("scope", apiToken.scope)
			}

			paresedURL.RawQuery = q.Encode()

			authRequest, err := http.NewRequest("GET", paresedURL.String(), nil)

			if registry.Username != "" || registry.Password != "" {
				authRequest.SetBasicAuth(registry.Username, registry.Password)
			}

			client := http.Client{
				Transport: registry.Client.Transport,
			}

			response, err := client.Do(authRequest)
			if err != nil {
				return apiToken, err
			}

			if response.StatusCode != http.StatusOK {
				return apiToken, err
			}
			defer response.Body.Close()

			decoder := json.NewDecoder(response.Body)
			err = decoder.Decode(&apiToken)
			apiToken.Created = time.Now()
			// expstr, err := strconv.Atoi(apiToken.expiresIn)
			apiToken.Expires = time.Now().Add(time.Second * time.Duration(apiToken.ExpiresIn))

			Authtokenmap().Set(org, apiToken)

		}
	} else { //token exists, check expiration
		apiToken, err := Authtokenmap().Get(org)

		if err != nil {
			return apiToken, err
		}
		if apiToken.IsExpired() {
			registry.RenewToken(apiToken)
		}
		return apiToken, nil
	}
	// err := errors.New("token is fine")
	return apiToken, nil
}

func (token *APIToken) IsExpired() bool {
	if token.Expires.Sub(token.Created).Seconds() < 0 {
		return true
	}
	return false
}

func (registry *V2registry) RenewToken(token APIToken) error {
	paresedURL, err := url.Parse(token.realm)

	if err != nil {
		return err
	}

	q := paresedURL.Query()
	q.Set("service", token.service)
	if token.scope != "" {
		q.Set("scope", token.scope)
	}

	paresedURL.RawQuery = q.Encode()

	authRequest, err := http.NewRequest("GET", paresedURL.String(), nil)

	if registry.Username != "" || registry.Password != "" {
		authRequest.SetBasicAuth(registry.Username, registry.Password)
	}

	client := http.Client{
		Transport: registry.Client.Transport,
	}

	response, err := client.Do(authRequest)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return err
	}
	defer response.Body.Close()

	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&token)
	token.Created = time.Now()
	// expstr, err := strconv.Atoi(token.expiresIn)
	token.Expires = time.Now().Add(time.Second * time.Duration(token.ExpiresIn))
	return nil
}
