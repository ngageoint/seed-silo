package v2

import (
	"bytes"
	// "encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	// "net/url"

	"github.com/docker/distribution/digest"

	manifestV1 "github.com/docker/distribution/manifest/schema1"
	manifestV2 "github.com/docker/distribution/manifest/schema2"
)

func (registry *V2registry) Manifest(repository, reference string) (*manifestV1.SignedManifest, error) {
	url := registry.url("/v2/%s/manifests/%s", repository, reference)
	// registry.Logf("registry.manifest.get url=%s repository=%s reference=%s", url, repository, reference)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", manifestV1.MediaTypeManifest)
	resp, err := registry.Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	signedManifest := &manifestV1.SignedManifest{}
	err = signedManifest.UnmarshalJSON(body)
	if err != nil {
		return nil, err
	}

	return signedManifest, nil
}

func (registry *V2registry) ManifestV2(repository, reference string) (*manifestV2.DeserializedManifest, error) {
	registryURL := registry.url("/v2/%s/manifests/%s", repository, reference)
	log.Printf("registry.manifest.get url=%s repository=%s reference=%s", registryURL, repository, reference)

	req, err := http.NewRequest("GET", registryURL, nil)
	if err != nil {
		return nil, err
	}

	// look for token here, set it here
	token, err := registry.GetOrCreateToken(repository, registryURL)
	if err != nil {
		return nil, err
	}
	log.Printf("Setting auth token to: %s", token.Token)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.Token))
	// req.SetBasicAuth(registry.Username, registry.Password)

	req.Header.Set("Accept", manifestV2.MediaTypeManifest)
	resp, err := registry.Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.Printf("Manifest response returned error code: %v", resp.StatusCode)
		log.Printf("Manifest response body: \n %s", body)
	}
	if err != nil {
		return nil, err
	}

	deserialized := &manifestV2.DeserializedManifest{}
	err = deserialized.UnmarshalJSON(body)
	if err != nil {
		return nil, err
	}
	return deserialized, nil
}

func (registry *V2registry) ManifestDigest(repository, reference string) (digest.Digest, error) {
	url := registry.url("/v2/%s/manifests/%s", repository, reference)
	// registry.Logf("registry.manifest.head url=%s repository=%s reference=%s", url, repository, reference)

	resp, err := registry.Client.Head(url)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return "", err
	}
	return digest.ParseDigest(resp.Header.Get("Docker-Content-Digest"))
}

func (registry *V2registry) DeleteManifest(repository string, digest digest.Digest) error {
	url := registry.url("/v2/%s/manifests/%s", repository, digest)
	// registry.Logf("registry.manifest.delete url=%s repository=%s reference=%s", url, repository, digest)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	resp, err := registry.Client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}
	return nil
}

func (registry *V2registry) PutManifest(repository, reference string, signedManifest *manifestV1.SignedManifest) error {
	url := registry.url("/v2/%s/manifests/%s", repository, reference)
	// registry.Logf("registry.manifest.put url=%s repository=%s reference=%s", url, repository, reference)

	body, err := signedManifest.MarshalJSON()
	if err != nil {
		return err
	}

	buffer := bytes.NewBuffer(body)
	req, err := http.NewRequest("PUT", url, buffer)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", manifestV1.MediaTypeManifest)
	resp, err := registry.Client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	return err
}
