package commands

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ngageoint/seed-cli/objects"
	"github.com/ngageoint/seed-cli/util"
)

func init() {
	util.InitPrinter(false)
}

func TestDockerBuild(t *testing.T) {
	cases := []struct {
		directory        string
		expected         bool
		expectedErrorMsg string
	}{
		{"../examples/addition-job/", true, ""},
		{"../examples/extractor/", true, ""},
		{"", false, "no such file or directory"},
	}

	for _, c := range cases {
		err := DockerBuild(c.directory, "", "")
		success := err == nil
		if success != c.expected {
			t.Errorf("DockerBuild(%q) == %v, expected %v", c.directory, success, c.expected)
		}
		if err != nil {
			if !strings.Contains(err.Error(), c.expectedErrorMsg) {
				t.Errorf("DockerBuild(%q) == %v, expected %v", c.directory, err.Error(), c.expectedErrorMsg)
			}
		}
	}
}

func TestSeedLabel(t *testing.T) {
	cases := []struct {
		directory        string
		imageName        string
		expected         bool
		expectedErrorMsg string
	}{
		{"../examples/addition-job/", "addition-job-0.0.1-seed:1.0.0", true, ""},
		{"../examples/extractor/", "extractor-0.1.0-seed:0.1.0", true, ""},
	}

	for _, c := range cases {
		DockerBuild(c.directory, "", "")
		seedFileName, exist, _ := util.GetSeedFileName(c.directory)
		if !exist {
			t.Errorf("ERROR: %s cannot be found.\n",
				seedFileName)
			t.Errorf("Make sure you have specified the correct directory.\n")
		}

		// retrieve seed from seed manifest
		seed := objects.SeedFromManifestFile(seedFileName)

		seed2 := objects.SeedFromImageLabel(c.imageName)
		seedStr1 := fmt.Sprintf("%v", seed)
		seedStr2 := fmt.Sprintf("%v", seed2)

		success := seedStr1 != "" && seedStr1 == seedStr2
		if success != c.expected {
			t.Errorf("SeedFromImageLabel(%q) == %v, expected %v", seedFileName, seedStr1, seedStr2)
		}
	}
}

func TestImageName(t *testing.T) {
	cases := []struct {
		filename         string
		expected         string
		expectedErrorMsg string
	}{
		{"../examples/addition-job/seed.manifest.json", "addition-job-0.0.1-seed:1.0.0", ""},
		{"../examples/extractor/seed.manifest.json", "extractor-0.1.0-seed:0.1.0", ""},
	}

	for _, c := range cases {

		seedFileName := util.GetFullPath(c.filename, "")
		// retrieve seed from seed manifest
		seed := objects.SeedFromManifestFile(seedFileName)

		// Retrieve docker image name
		imageName := objects.BuildImageName(&seed)

		if imageName != c.expected {
			t.Errorf("BuildImageName(%q) == %v, expected %v", seedFileName, imageName, c.expected)
		}
	}
}
