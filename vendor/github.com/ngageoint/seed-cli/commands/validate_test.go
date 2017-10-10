package commands

import (
	"strings"
	"testing"

	"github.com/ngageoint/seed-cli/constants"
	"github.com/ngageoint/seed-cli/util"
)

func init() {
	util.InitPrinter(false)
}

func TestValidate(t *testing.T) {
	cases := []struct {
		seedFileName     string
		expected         bool
		expectedErrorMsg string
	}{
		{"../examples/addition-job/seed.manifest.json", true, ""},
		{"../examples/extractor/seed.manifest.json", true, ""},
		{"../testdata/invalid-missing-job/seed.manifest.json",
			false, "job is required"},
		{"../testdata/invalid-missing-job-interface-inputs-files-name/seed.manifest.json",
			false, "name is required"},
		{"../testdata/invalid-reserved-name/seed.manifest.json",
			false, "Multiple Name values are assigned the same INPUT Name value. Each Name value must be unique."},
	}

	for _, c := range cases {
		name := util.GetFullPath(c.seedFileName, "")
		err := ValidateSeedFile("", name, constants.SchemaManifest)
		success := err == nil
		if success != c.expected {
			t.Errorf("ValidateSeedFile(%q, %q, %q) == %v, expected %v", "", name, constants.SchemaManifest, success, c.expected)
		}
		if err != nil {
			if !strings.Contains(err.Error(), c.expectedErrorMsg) {
				t.Errorf("ValidateSeedFile(%q, %q, %q) == %v, expected %v", "", name, constants.SchemaManifest, err.Error(), c.expectedErrorMsg)
			}
		}
	}
}
