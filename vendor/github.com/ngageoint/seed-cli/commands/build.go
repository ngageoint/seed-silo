package commands

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/ngageoint/seed-cli/constants"
	"github.com/ngageoint/seed-cli/objects"
	"github.com/ngageoint/seed-cli/util"
)

//DockerBuild Builds the docker image with the given image tag.
func DockerBuild(jobDirectory, username, password string) error {
	if username != "" {
		//set config dir so we don't stomp on other users' logins with sudo
		configDir := constants.DockerConfigDir + time.Now().Format(time.RFC3339)
		os.Setenv(constants.DockerConfigKey, configDir)
		defer util.RemoveAllFiles(configDir)
		defer os.Unsetenv(constants.DockerConfigKey)

		registry, err := util.DockerfileBaseRegistry(jobDirectory)
		if err != nil {
			util.PrintUtil("Error getting registry from dockerfile: %s\n", err.Error())
		}
		err = util.Login(registry, username, password)
		if err != nil {
			util.PrintUtil("Error calling docker login: %s\n", err.Error())
		}
	}

	seedFileName, err := util.SeedFileName(jobDirectory)
	if err != nil && !os.IsNotExist(err) {
		util.PrintUtil( "ERROR: %s\n", err.Error())
		return err
	}

	// Validate seed file
	err = ValidateSeedFile("", seedFileName, constants.SchemaManifest)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: seed file could not be validated. See errors for details.")
		util.PrintUtil( "%s", err.Error())
		util.PrintUtil( "Exiting seed...\n")
		return err
	}

	// retrieve seed from seed manifest
	seed := objects.SeedFromManifestFile(seedFileName)

	// Retrieve docker image name
	imageName := objects.BuildImageName(&seed)

	// Build Docker image
	util.PrintUtil( "INFO: Building %s\n", imageName)
	buildArgs := []string{"build", "-t", imageName, jobDirectory}
	if util.DockerVersionHasLabel() {
		// Set the seed.manifest.json contents as an image label
		label := "com.ngageoint.seed.manifest=" + objects.GetManifestLabel(seedFileName)
		buildArgs = append(buildArgs, "--label", label)
	}
	cmd := exec.Command("docker", buildArgs...)
	var errs bytes.Buffer
	cmd.Stderr = io.MultiWriter(os.Stderr, &errs)
	cmd.Stdout = os.Stderr

	// Run docker build
	if err := cmd.Run(); err != nil {
		util.PrintUtil( "ERROR: Error executing docker build. %s\n",
			err.Error())
		return err
	}

	// check for errors on stderr
	if errs.String() != "" {
		util.PrintUtil( "ERROR: Error building image '%s':\n%s\n",
			imageName, errs.String())
		util.PrintUtil( "Exiting seed...\n")
		return errors.New(errs.String())
	}

	return nil
}

//PrintBuildUsage prints the seed build usage arguments, then exits the program
func PrintBuildUsage() {
	util.PrintUtil( "\nUsage:\tseed build [-d JOB_DIRECTORY]\n")
	util.PrintUtil( "\nOptions:\n")
	util.PrintUtil(
		"  -%s  -%s\tDirectory containing Seed spec and Dockerfile (default is current directory)\n",
		constants.ShortJobDirectoryFlag, constants.JobDirectoryFlag)
	util.PrintUtil( "  -%s -%s\tUsername to login if needed to pull images (default anonymous).\n",
		constants.ShortUserFlag, constants.UserFlag)
	util.PrintUtil( "  -%s -%s\tPassword to login if needed to pull images (default anonymous).\n",
		constants.ShortPassFlag, constants.PassFlag)
	panic(util.Exit{0})
}
