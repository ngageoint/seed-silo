package commands

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/ngageoint/seed-cli/util"
)

func init() {
	util.InitPrinter(false)
}

func TestSeedInit(t *testing.T) {
	cases := []struct {
		directory   string
		expectedErr error
	}{
		{"../testdata/dummy-scratch/", nil},
		{"../testdata/complete/", errors.New("Existing file left unmodified.")},
	}

	for _, c := range cases {
		err := SeedInit(c.directory)

		if c.expectedErr == nil && err != nil {
			t.Errorf("SeedInit(%q) == %v, expected %v", c.directory, err.Error(), c.expectedErr)
		}
		if err != nil {
			if !strings.Contains(err.Error(), c.expectedErr.Error()) {
				t.Errorf("SeedInit(%q) == %v, expected %v", c.directory, err.Error(), c.expectedErr.Error())
			}
		}
	}

	// Cleanup test file
	os.Remove("../testdata/dummy-scratch/seed.manifest.json")
}
