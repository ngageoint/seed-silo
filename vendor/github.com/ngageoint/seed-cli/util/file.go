package util

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/ngageoint/seed-cli/constants"
)

//GetFullPath returns the full path of the given file. This expands relative file
// paths and verifes non-relative paths
// Validate path for file existance??
func GetFullPath(rFile, directory string) string {

	// Normalize
	rFile = filepath.Clean(filepath.ToSlash(rFile))

	if !filepath.IsAbs(rFile) {

		// Expand relative file path
		// Define the current working directory
		curDir, _ := os.Getwd()

		// Test relative to current directory
		dir := filepath.Join(curDir, rFile)
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			rFile = filepath.Clean(dir)

			// see if parent directory exists. If so, assume this directory will be created
		} else if _, err := os.Stat(filepath.Dir(dir)); !os.IsNotExist(err) {
			rFile = filepath.Clean(dir)
		}

		// Test relative to working directory
		if directory != "." {
			dir = filepath.Join(directory, rFile)
			if _, err := os.Stat(dir); !os.IsNotExist(err) {
				rFile = filepath.Clean(dir)

				// see if parent directory exists. If so, assume this directory will be created
			} else if _, err := os.Stat(filepath.Dir(dir)); !os.IsNotExist(err) {
				rFile = filepath.Clean(dir)
			}
		}
	}

	return rFile
}

//DockerfileRegistry attempts to find the registry for a dockerfile's base image, if any
func DockerfileBaseRegistry(dir string) (string, error) {
	registry := ""

	// Define the current working directory
	curDirectory, _ := os.Getwd()

	dockerfile := "Dockerfile"
	if dir == "." {
		dockerfile = filepath.Join(curDirectory, dockerfile)
	} else {
		if filepath.IsAbs(dir) {
			dockerfile = filepath.Join(dir, dockerfile)
		} else {
			dockerfile = filepath.Join(curDirectory, dir, dockerfile)
		}
	}

	// Verify dockerfile exists within specified directory.
	_, err := os.Stat(dockerfile)
	if os.IsNotExist(err) {
		PrintUtil( "ERROR: %s cannot be found.\n",
			dockerfile)
		PrintUtil( "Make sure you have specified the correct directory.\n")
	}

	file, err := os.Open(dockerfile)
	if err == nil {

		// make sure it gets closed
		defer file.Close()

		// create a new scanner and read the file line by line
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			i := strings.Index(line, "/")
			if i > 5 && strings.HasPrefix(line, "FROM ") {
				registry = line[5:i]
				break
			}
		}

		// check for errors
		err = scanner.Err()
	}

	return registry, err
}

//GetSeedFileName Finds and returns the full filepath to the seed.manifest.json
// The second return value indicates whether it exists or not.
func GetSeedFileName(dir string) (string, bool, error) {
	// Define the current working directory
	curDirectory, _ := os.Getwd()

	// set path to seed file -
	// 	Either relative to current directory or located in given directory
	//  	-d directory might be a relative path to current directory
	seedFileName := constants.SeedFileName
	if dir == "." {
		seedFileName = filepath.Join(curDirectory, seedFileName)
	} else {
		if filepath.IsAbs(dir) {
			seedFileName = filepath.Join(dir, seedFileName)
		} else {
			seedFileName = filepath.Join(curDirectory, dir, seedFileName)
		}
	}

	// Check to see if seed.manifest.json exists within specified directory.
	_, err := os.Stat(seedFileName)
	return seedFileName, !os.IsNotExist(err), err
}

//SeedFileName Finds and returns the full filepath to the seed.manifest.json
func SeedFileName(dir string) (string, error) {
	seedFileName, exists, err := GetSeedFileName(dir)
	if !exists {
		PrintUtil( "ERROR: %s cannot be found.\n",
			seedFileName)
		PrintUtil( "Make sure you have specified the correct directory.\n")
	}

	return seedFileName, err
}

//RemoveAllFiles removes all files in the specified directory
func RemoveAllFiles(v string) {
	err := os.RemoveAll(v)
	if err != nil {
		PrintUtil( "Error removing directory: %s\n", err.Error())
	}
}

func ReadLinesFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	lines := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}