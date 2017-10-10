package objects

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/ngageoint/seed-cli/constants"
	"github.com/ngageoint/seed-cli/util"
)

//Seed represents a seed.manifest.json object.
type Seed struct {
	SeedVersion string `json:"seedVersion"`
	Job         Job    `json:"job"`
}

type Job struct {
	Name           string     `json:"name"`
	JobVersion     string     `json:"jobVersion"`
	PackageVersion string     `json:"packageVersion"`
	Title          string     `json:"title,omitempty"`
	Description    string     `json:"description,omitempty"`
	Tags           []string   `json:"tags,omitempty"`
	Maintainer     Maintainer `json:"maintainer"`
	Timeout        int        `json:"timeout,omitempty"`
	Interface      Interface  `json:"interface,omitempty"`
	Resources      Resources  `json:"resources,omitempty"`
	Errors         []ErrorMap `json:"errors,omitempty"`
}

type Maintainer struct {
	Name         string `json:"name"`
	Organization string `json:"organization,omitempty"`
	Email        string `json:"email"`
	Url          string `json:"url,omitempty"`
	Phone        string `json:"phone,omitempty"`
}

type Interface struct {
	Command  string    `json:"command"`
	Inputs   Inputs    `json:"inputs,omitempty"`
	Outputs  Outputs   `json:"outputs,omitempty"`
	Mounts   []Mount   `json:"mounts,omitempty"`
	Settings []Setting `json:"settings,omitempty"`
}

type Resources struct {
	Scalar []Scalar `json:"scalar"`
}

type Scalar struct {
	Name            string  `json:"name"`
	Value           float64 `json:"value"`
	InputMultiplier float64 `json:"inputMultiplier"`
}

type Inputs struct {
	Files []InFile `json:"files,ommitempty"`
	Json  []InJson `json:"json,omitempty"`
}

type InFile struct {
	Name       string   `json:"name"`
	MediaTypes []string `json:"mediaTypes"`
	Multiple   bool     `json:"multiple"`
	Required   bool     `json:"required"`
}

func (o *InFile) UnmarshalJSON(b []byte) error {
	type xInFile InFile
	xo := &xInFile{Multiple: false, Required: true}
	if err := json.Unmarshal(b, xo); err != nil {
		return err
	}
	*o = InFile(*xo)
	return nil
}

type InJson struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

func (o *InJson) UnmarshalJSON(b []byte) error {
	type xInJson InJson
	xo := &xInJson{Required: true}
	if err := json.Unmarshal(b, xo); err != nil {
		return err
	}
	*o = InJson(*xo)
	return nil
}

type Outputs struct {
	Files []OutFile `json:"files,omitempty"`
	JSON  []OutJson `json:"json,omitempty"`
}

type OutFile struct {
	Name      string `json:"name"`
	MediaType string `json:"mediaType"`
	Count     string `json:"count"`
	Pattern   string `json:"pattern"`
	Required  bool   `json:"required"`
}

func (o *OutFile) UnmarshalJSON(b []byte) error {
	type xOutFile OutFile
	xo := &xOutFile{Count: "1", Required: true}
	if err := json.Unmarshal(b, xo); err != nil {
		return err
	}
	*o = OutFile(*xo)
	return nil
}

type OutJson struct {
	Name     string `json:"name"`
	Key      string `json:"key"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

func (o *OutJson) UnmarshalJSON(b []byte) error {
	type xOutJson OutJson
	xo := &xOutJson{Required: true}
	if err := json.Unmarshal(b, xo); err != nil {
		return err
	}
	*o = OutJson(*xo)
	return nil
}

type Mount struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Mode string `json:"mode"`
}

func (o *Mount) UnmarshalJSON(b []byte) error {
	type xMount Mount
	xo := &xMount{Mode: "ro"}
	if err := json.Unmarshal(b, xo); err != nil {
		return err
	}
	*o = Mount(*xo)
	return nil
}

type Setting struct {
	Name   string `json:"name"`
	Secret bool   `json:"secret"`
}

func (o *Setting) UnmarshalJSON(b []byte) error {
	type xSetting Setting
	xo := &xSetting{Secret: false}
	if err := json.Unmarshal(b, xo); err != nil {
		return err
	}
	*o = Setting(*xo)
	return nil
}

type ErrorMap struct {
	Code        int    `json:"code"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

func (o *ErrorMap) UnmarshalJSON(b []byte) error {
	type xErrorMap ErrorMap
	xo := &xErrorMap{Category: "job"}
	if err := json.Unmarshal(b, xo); err != nil {
		return err
	}
	*o = ErrorMap(*xo)
	return nil
}

//GetManifestLabel returns the seed.manifest.json as LABEL
//  com.ngageoint.seed.manifest contents
func GetManifestLabel(seedFileName string) string {
	// read the seed.manifest.json into a string
	seedbytes, err := ioutil.ReadFile(seedFileName)
	if err != nil {
		util.PrintUtil( "ERROR: Error reading %s. %s\n", seedFileName,
			err.Error())
		os.Exit(1)
	}
	var seedbuff bytes.Buffer
	json.Compact(&seedbuff, seedbytes)
	seedbytes, err = json.Marshal(seedbuff.String())
	if err != nil {
		util.PrintUtil( "ERROR: Error marshalling seed manifest. %s\n",
			err.Error())
	}

	// Escape forward slashes and dollar signs
	seed := string(seedbytes)
	seed = strings.Replace(seed, "$", "\\$", -1)
	seed = strings.Replace(seed, "/", "\\/", -1)

	return seed
}

//SeedFromImageLabel returns seed parsed from the docker image LABEL
func SeedFromImageLabel(imageName string) Seed {
	cmdStr := "inspect -f '{{index .Config.Labels \"com.ngageoint.seed.manifest\"}}'" + imageName
	util.PrintUtil(
		"INFO: Retrieving seed manifest from %s LABEL=com.ngageoint.seed.manifest\n",
		imageName)

	inspectCommand := exec.Command("docker", "inspect", "-f",
		"'{{index .Config.Labels \"com.ngageoint.seed.manifest\"}}'", imageName)

	errPipe, errr := inspectCommand.StderrPipe()
	if errr != nil {
		util.PrintUtil(
			"ERROR: error attaching to docker inspect command stderr. %s\n",
			errr.Error())
	}

	// Attach stdout pipe
	outPipe, errr := inspectCommand.StdoutPipe()
	if errr != nil {
		util.PrintUtil(
			"ERROR: error attaching to docker inspect command stdout. %s\n",
			errr.Error())
	}

	// Run docker inspect
	if err := inspectCommand.Start(); err != nil {
		util.PrintUtil( "ERROR: error executing docker %s. %s\n", cmdStr,
			err.Error())
	}

	// Print out any std out
	seedBytes, err := ioutil.ReadAll(outPipe)
	if err != nil {
		fmt.Fprintf(os.Stdout, "ERROR: Error retrieving docker %s stdout.\n%s\n",
			cmdStr, err.Error())
	}

	// check for errors on stderr
	slurperr, _ := ioutil.ReadAll(errPipe)
	if string(slurperr) != "" {
		util.PrintUtil( "ERROR: Error executing docker %s:\n%s\n",
			cmdStr, string(slurperr))
		util.PrintUtil( "Exiting seed...\n")
		os.Exit(1)
	}

	// un-escape special characters
	seedStr := string(seedBytes)
	seedStr = strings.Replace(seedStr, "\\\"", "\"", -1)
	seedStr = strings.Replace(seedStr, "\\\"", "\"", -1) //extra replace to fix extra back slashes added by docker build command
	seedStr = strings.Replace(seedStr, "\\$", "$", -1)
	seedStr = strings.Replace(seedStr, "\\/", "/", -1)
	seedStr = strings.TrimSpace(seedStr)
	seedStr = strings.TrimSuffix(strings.TrimPrefix(seedStr, "'\""), "\"'")

	seed := &Seed{}

	err = json.Unmarshal([]byte(seedStr), &seed)
	if err != nil {
		util.PrintUtil( "ERROR: Error unmarshalling seed: %s\n", err.Error())
	}

	return *seed
}

//SeedFromManifestFile returns seed struct parsed from seed file
func SeedFromManifestFile(seedFileName string) Seed {

	// Open and parse seed file into struct
	seedFile, err := os.Open(seedFileName)
	if err != nil {
		util.PrintUtil( "ERROR: Error opening %s. Error received is: %s\n",
			seedFileName, err.Error())
		util.PrintUtil( "Exiting seed...\n")
		os.Exit(1)
	}
	jsonParser := json.NewDecoder(seedFile)
	var seed Seed
	if err = jsonParser.Decode(&seed); err != nil {
		util.PrintUtil(
			"ERROR: A valid %s must be present in the working directory. Error parsing %s.\nError received is: %s\n",
			constants.SeedFileName, seedFileName, err.Error())
		util.PrintUtil( "Exiting seed...\n")
		os.Exit(1)
	}

	return seed
}

//BuildImageName extracts the Docker Image name from the seed.json
// 	jobName-jobVersion-seed:pkgVersion
func BuildImageName(seed *Seed) string {
	var buffer bytes.Buffer

	buffer.WriteString(seed.Job.Name)
	buffer.WriteString("-")
	buffer.WriteString(seed.Job.JobVersion)
	buffer.WriteString("-seed")
	buffer.WriteString(":")
	buffer.WriteString(seed.Job.PackageVersion)

	return buffer.String()
}
