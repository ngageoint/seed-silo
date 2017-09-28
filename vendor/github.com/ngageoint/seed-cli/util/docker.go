package util

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

//CheckSudo Checks error for telltale sign seed command should be run as sudo
func CheckSudo() {
	cmd := exec.Command("docker", "info")

	// attach stderr pipe
	errPipe, err := cmd.StderrPipe()
	if err != nil {
		PrintUtil("ERROR: Error attaching to version command stderr. %s\n", err.Error())
	}

	// Run docker build
	if err := cmd.Start(); err != nil {
		PrintUtil( "ERROR: Error executing docker version. %s\n",
			err.Error())
	}

	slurperr, _ := ioutil.ReadAll(errPipe)
	er := string(slurperr)
	if er != "" {
		if strings.Contains(er, "Cannot connect to the Docker daemon. Is the docker daemon running on this host?") ||
			strings.Contains(er, "dial unix /var/run/docker.sock: connect: permission denied") {
			PrintUtil( "Elevated permissions are required by seed to run Docker. Try running the seed command again as sudo.\n")
			panic(Exit{1})
		}
	}
}

//DockerVersionHasLabel returns if the docker version is greater than 1.11.1
func DockerVersionHasLabel() bool {
	return DockerVersionGreaterThan(1, 11, 1)
}

//DockerVersionHasLabel returns if the docker version is greater than 1.13.0
func DockerVersionHasReferenceFilter() bool {
	return DockerVersionGreaterThan(1, 13, 0)
}

//DockerVersionGreaterThan returns if the docker version is greater than the specified version
func DockerVersionGreaterThan(major, minor, patch int) bool {
	cmd := exec.Command("docker", "version", "-f", "{{.Client.Version}}")

	// Attach stdout pipe
	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		PrintUtil("ERROR: Error attaching to version command stdout. %s\n", err.Error())
	}

	// Run docker version
	if err := cmd.Start(); err != nil {
		PrintUtil("ERROR: Error executing docker version. %s\n", err.Error())
	}

	// Print out any std out
	slurp, _ := ioutil.ReadAll(outPipe)
	if string(slurp) != "" {
		version := strings.Split(string(slurp), ".")

		// check each part of version. Return false if 1st < 1, 2nd < 11, 3rd < 1
		if len(version) > 1 {
			v1, _ := strconv.Atoi(version[0])
			v2, _ := strconv.Atoi(version[1])

			// check for minimum of 1.11.1
			if v1 == major {
				if v2 > minor {
					return true
				} else if v2 == minor && len(version) == 3 {
					v3, _ := strconv.Atoi(version[2])
					if v3 >= patch {
						return true
					}
				}
			} else if v1 > major {
				return true
			}

			return false
		}
	}

	return false
}

//ImageExists returns true if a local image already exists, false otherwise
func ImageExists(imageName string) (bool, error) {
	// Test if image has been built; Rebuild if not
	imgsArgs := []string{"images", "-q", imageName}
	imgOut, err := exec.Command("docker", imgsArgs...).Output()
	if err != nil {
		PrintUtil( "ERROR: Error executing docker %v\n", imgsArgs)
		PrintUtil( "%s\n", err.Error())
		return false, err
	} else if string(imgOut) == "" {
		PrintUtil( "INFO: No docker image found locally for image name %s.\n",
			imageName)
		return false, nil
	}
	return true, nil
}

//ImageCpuUsage displays CPU usage of image
func ImageCpuUsage(imageName string) {

}

//ImageMemoryUsage displays memory usage of image
func ImageMemoryUsage(imageName string) {

}

func Login(registry, username, password string) error {
	var errs, out bytes.Buffer
	args := []string{"login", "-u", username, "-p", password, registry}
	cmd := exec.Command("docker", args...)
	cmd.Stderr = io.MultiWriter(os.Stderr, &errs)
	cmd.Stdout = &out

	err := cmd.Run()

	if errs.String() != "" {
		PrintUtil( "ERROR: Error reading stderr %s\n",
			errs.String())
		return errors.New(errs.String())
	}

	if err != nil {
		errMsg := fmt.Sprintf("ERROR: Error executing docker login.\n%s\n", err.Error())
		errors.New(errMsg)
		return err
	}

	PrintUtil( "%s", out.String())
	return nil
}

func Tag(origImg, img string) error {
	var errs bytes.Buffer

	// Run docker tag
	if img != origImg {
		PrintUtil( "INFO: Tagging image %s as %s\n", origImg, img)
		tagCmd := exec.Command("docker", "tag", origImg, img)
		tagCmd.Stderr = io.MultiWriter(os.Stderr, &errs)
		tagCmd.Stdout = os.Stderr

		if err := tagCmd.Run(); err != nil {
			PrintUtil( "ERROR: Error executing docker tag. %s\n",
				err.Error())
		}
		if errs.String() != "" {
			PrintUtil( "ERROR: Error tagging image '%s':\n%s\n", origImg, errs.String())
			PrintUtil( "Exiting seed...\n")
			return errors.New(errs.String())
		}
	}

	return nil
}

func Push(img string) error {
	var errs bytes.Buffer

	// docker push
	PrintUtil( "INFO: Performing docker push %s\n", img)
	errs.Reset()
	pushCmd := exec.Command("docker", "push", img)
	pushCmd.Stderr = io.MultiWriter(os.Stderr, &errs)
	pushCmd.Stdout = os.Stdout

	// Run docker push
	if err := pushCmd.Run(); err != nil {
		PrintUtil( "ERROR: Error executing docker push. %s\n",
			err.Error())
		return err
	}

	// Check for errors. Exit if error occurs
	if errs.String() != "" {
		PrintUtil( "ERROR: Error pushing image '%s':\n%s\n", img,
			errs.String())
		PrintUtil( "Exiting seed...\n")
		return errors.New(errs.String())
	}

	return nil
}

func RemoveImage(img string) error {
	var errs bytes.Buffer

	PrintUtil( "INFO: Removing local image %s\n", img)
	rmiCmd := exec.Command("docker", "rmi", img)
	rmiCmd.Stderr = io.MultiWriter(os.Stderr, &errs)
	rmiCmd.Stdout = os.Stdout

	if err := rmiCmd.Run(); err != nil {
		PrintUtil( "ERROR: Error executing docker rmi. %s\n",
			err.Error())
		return err
	}

	// check for errors on stderr
	if errs.String() != "" {
		PrintUtil( "ERROR: Error removing image '%s':\n%s\n", img,
			errs.String())
		PrintUtil( "Exiting seed...\n")
		return errors.New(errs.String())
	}

	return nil
}

func RestartRegistry() error {
	PrintUtil("RESTARTING REGISTRY........................\n.\n.\n.\n.\n.\n")
	var errs bytes.Buffer

	PrintUtil( "INFO: Restarting test registry...\n")
	cmd := exec.Command("../restartRegistry.sh")
	cmd.Stderr = io.MultiWriter(os.Stderr, &errs)
	cmd.Stdout = os.Stdout

	err := cmd.Run()

	// check for errors on stderr first; it will likely have more explanation than cmd.Run
	if errs.String() != "" {
		PrintUtil( "ERROR: Error restarting registry. %s\n", errs.String())
		PrintUtil( "Exiting seed...\n")
		return errors.New(errs.String())
	}

	if err != nil {
		PrintUtil( "ERROR: Error restarting registry. %s\n",
			err.Error())
		return err
	}

	return nil
}