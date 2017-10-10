package commands

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/ngageoint/seed-cli/constants"
	"github.com/ngageoint/seed-cli/util"
)

//SeedInit places a sample seed.manifest.json within given directory (defaults to CWD)
// Should check for file existence in given directory
// If file exists, warn and exit
// If file does not exist, write sample to given directory
func SeedInit(directory string) error {
	seedFileName, exists, err := util.GetSeedFileName(directory)
	if err != nil && exists {
		//an error occurred other than the file not existing, i.e. permission error
		util.PrintUtil( "ERROR: Error occurred writing example Seed manifest to %s.\n%s\n",
			seedFileName, err.Error())
		return errors.New("Error writing example Seed manifest.")
	} else if exists {
		msg := "Pre-existing " + seedFileName + " found. Existing file left unmodified."
		util.PrintUtil( "%s\n", msg)
		return nil
	}


	// TODO: We need to support init of all supported schema versions in the future
	exampleSeedJson, _ := constants.Asset("schema/0.1.0/seed.manifest.example.json")

	err = ioutil.WriteFile(seedFileName, exampleSeedJson, os.ModePerm)
	if err != nil {
		util.PrintUtil( "ERROR: Error occurred writing example Seed manifest to %s.\n%s\n",
			seedFileName, err.Error())
		return errors.New("Error writing example Seed manifest.")
	}

	util.PrintUtil( "Created Seed file: %s\n", seedFileName)

	return nil
}

//PrintBuildUsage prints the seed build usage arguments, then exits the program
func PrintInitUsage() {
	util.PrintUtil( "\nUsage:\tseed init [-d JOB_DIRECTORY]\n")
	util.PrintUtil( "\nOptions:\n")
	util.PrintUtil(
		"  -%s  -%s\tDirectory to place seed.manifest.json example. (default is current directory)\n",
		constants.ShortJobDirectoryFlag, constants.JobDirectoryFlag)
	panic(util.Exit{0})
}
