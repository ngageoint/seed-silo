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
	"github.com/ngageoint/seed-cli/util"
	"strings"
)

//Dockerpull pulls specified image from remote repository (default docker.io)
func DockerPull(image, registry, org, username, password string) error {
	if username != "" {
		//set config dir so we don't stomp on other users' logins with sudo
		configDir := constants.DockerConfigDir + time.Now().Format(time.RFC3339)
		os.Setenv(constants.DockerConfigKey, configDir)
		defer util.RemoveAllFiles(configDir)
		defer os.Unsetenv(constants.DockerConfigKey)

		err := util.Login(registry, username, password)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	if registry == "" {
		registry = constants.DefaultRegistry
	}

	registry = strings.Replace(registry, "https://hub.docker.com", "docker.io", 1)

	remoteImage := fmt.Sprintf("%s/%s", registry, image)

	if org != "" {
		remoteImage = fmt.Sprintf("%s/%s/%s", registry, org, image)
	}

	var errs, out bytes.Buffer
	// pull image
	pullArgs := []string{"pull", remoteImage}
	pullCmd := exec.Command("docker", pullArgs...)
	pullCmd.Stderr = io.MultiWriter(os.Stderr, &errs)
	pullCmd.Stdout = &out

	err := pullCmd.Run()
	if err != nil {
		util.PrintUtil( "ERROR: Error executing docker pull.\n%s\n",
			err.Error())
		return err
	}

	if errs.String() != "" {
		util.PrintUtil( "ERROR: Error reading stderr %s\n",
			errs.String())
		return errors.New(errs.String())
	}

	// tag image
	tagArgs := []string{"tag", remoteImage, image}
	tagCmd := exec.Command("docker", tagArgs...)
	tagCmd.Stderr = io.MultiWriter(os.Stderr, &errs)
	tagCmd.Stdout = &out

	err = tagCmd.Run()
	if err != nil {
		util.PrintUtil( "ERROR: Error executing docker tag.\n%s\n",
			err.Error())
		return err
	}

	if errs.String() != "" {
		util.PrintUtil( "ERROR: Error reading stderr %s\n",
			errs.String())
		return errors.New(errs.String())
	}

	return nil
}

//PrintPullUsage prints the seed pull usage information, then exits the program
func PrintPullUsage() {
	util.PrintUtil( "\nUsage:\tseed pull -in IMAGE_NAME [-r REGISTRY_NAME] [-o ORGANIZATION_NAME] [-u Username] [-p password]\n")
	util.PrintUtil( "\nPulls seed image from remote repository.\n")
	util.PrintUtil( "\nOptions:\n")
	util.PrintUtil( "  -%s -%s Docker image name to pull\n",
		constants.ShortImgNameFlag, constants.ImgNameFlag)
	util.PrintUtil( "  -%s -%s\tSpecifies a specific registry (default is index.docker.io).\n",
		constants.ShortRegistryFlag, constants.RegistryFlag)
	util.PrintUtil( "  -%s -%s\tSpecifies a specific organization (default is no organization).\n",
		constants.ShortOrgFlag, constants.OrgFlag)
	util.PrintUtil( "  -%s -%s\tUsername to login to remote registry (default anonymous).\n",
		constants.ShortUserFlag, constants.UserFlag)
	util.PrintUtil( "  -%s -%s\tPassword to login to remote registry (default anonymous).\n",
		constants.ShortPassFlag, constants.PassFlag)
	panic(util.Exit{0})
}
