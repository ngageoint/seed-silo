package v2

import (
	"errors"
	"fmt"
	"github.com/ngageoint/seed-common/objects"
	"github.com/ngageoint/seed-common/util"
	"net/http"
	"os"
	"strings"
)

type V2registry struct {
	Hostname string
	Org      string
	Username string
	Password string
	Print    util.PrintCallback
	Client   *http.Client
}

func New(url, org, username, password string) (*V2registry, error) {
	if util.PrintUtil == nil {
		util.InitPrinter(util.PrintErr, os.Stderr, os.Stdout)
	}

	// reg, err := registry.New(url, username, password)
	// if reg != nil {
	// 	host := strings.Replace(url, "https://", "", 1)
	// 	host = strings.Replace(host, "http://", "", 1)
	// return &V2registry{r: reg, Hostname: host, Org: org, Username: username, Password: password, Print: util.PrintUtil}, err

	return &V2registry{Hostname: url, Org: org, Username: username, Password: password, Print: util.PrintUtil, Client: &http.Client{}}, nil
	// }
	// return nil, err
}

func (v2 *V2registry) Name() string {
	return "V2"
}

// func (v2 *V2registry) GetAuthToken() authToken {
// 	return authtokens[v2.Org]
// }

func (v2 *V2registry) Ping() error {
	_, err := v2.Repositories()
	return err
}

// func (v2 *V2registry) Repositories() ([]string, error) {
// 	return v2.Repositories()
// }

func (v2 *V2registry) Tags(repository string) ([]string, error) {
	return v2.Tags(repository)
}

func (v2 *V2registry) url(pathTemplate string, args ...interface{}) string {
	pathSuffix := fmt.Sprintf(pathTemplate, args...)
	url := fmt.Sprintf("%s%s", v2.Hostname, pathSuffix)
	return url
}

func (v2 *V2registry) Images() ([]string, error) {
	url := v2.Hostname + "/v2/_catalog"
	v2.Print("Searching %s for Seed images...\n", url)
	repositories, err := v2.Repositories()

	var images []string
	for _, repo := range repositories {
		if !strings.HasSuffix(repo, "-seed") {
			continue
		}
		tags, err := v2.Tags(repo)
		if err != nil {
			v2.Print(err.Error())
			continue
		}
		for _, tag := range tags {
			images = append(images, repo+":"+tag)
		}
	}

	return images, err
}

func (v2 *V2registry) ImagesWithManifests() ([]objects.Image, error) {
	imageNames, err := v2.Images()
	v2.Print("Images found in V2 Registry %s with Org %s: \n %v", v2.Hostname, v2.Org, imageNames)
	v2.Print("Getting Manifests for %d images in V2 Registry %s with Org %s", len(imageNames), v2.Hostname, v2.Org)

	if err != nil {
		return nil, err
	}

	images := []objects.Image{}

	for _, imgstr := range imageNames {
		v2.Print("Getting manifest for %s", imgstr)
		temp := strings.Split(imgstr, ":")
		if len(temp) != 2 {
			v2.Print("ERROR: Invalid seed name: %s. Unable to split into name/tag pair\n", imgstr)
			continue
		}
		manifest, err := v2.GetImageManifest(temp[0], temp[1])
		if err != nil {
			//skip images with empty manifests
			v2.Print("ERROR: Error reading v2 manifest for %s: %s\n Skipping.\n", imgstr, err.Error())
			continue
		}

		imageStruct := objects.Image{Name: imgstr, Registry: v2.Hostname, Org: v2.Org, Manifest: manifest}
		images = append(images, imageStruct)
	}

	return images, err
}

func (v2 *V2registry) GetImageManifest(repoName, tag string) (string, error) {
	manifest := ""
	mv2, err := v2.ManifestV2(repoName, tag)
	if err == nil {
		resp, err := v2.DownloadLayer(repoName, mv2.Config.Digest)
		if err == nil {
			manifest, err = objects.GetSeedManifestFromBlob(resp)
		}
	}

	if err == nil && manifest == "" {
		err = errors.New("Empty seed manifest!")
	}

	return manifest, err
}
