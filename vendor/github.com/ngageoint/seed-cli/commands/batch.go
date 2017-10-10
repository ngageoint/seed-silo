package commands

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ngageoint/seed-cli/constants"
	"github.com/ngageoint/seed-cli/objects"
	"github.com/ngageoint/seed-cli/util"
)

type BatchIO struct {
	Inputs []string
	Outdir string
}

func BatchRun(batchDir, batchFile, imageName, outputDir, metadataSchema string, settings, mounts []string, rmFlag bool) error {
	if imageName == "" {
		return errors.New("ERROR: No input image specified.")
	}

	if exists, err := util.ImageExists(imageName); !exists {
		return err
	}

	if batchDir == "" {
		batchDir = "."
	}

	batchDir = util.GetFullPath(batchDir, "")

	seed := objects.SeedFromImageLabel(imageName)

	outdir := getOutputDir(outputDir, imageName)

	var inputs []BatchIO
	var err error

	if batchFile != "" {
		inputs, err = ProcessBatchFile(seed, batchFile, outdir)
		if err != nil {
			util.PrintUtil( "ERROR: Error processing batch file: %s\n", err.Error())
			return err
		}
	} else {
		inputs, err = ProcessDirectory(seed, batchDir, outdir)
		if err != nil {
			util.PrintUtil( "ERROR: Error processing batch directory: %s\n", err.Error())
			return err
		}
	}


	out := "Results: \n"
	for _, in := range inputs {
		exitCode, err := DockerRun(imageName, in.Outdir, metadataSchema, in.Inputs, settings, mounts, rmFlag, true)

		//trim inputs to print only the key values and filenames
		truncatedInputs := []string{}
		for _, i := range in.Inputs {
			begin := strings.Index(i, "=") + 1
			end := strings.LastIndex(i, "/")
			truncatedInputs = append(truncatedInputs, i[0:begin] + "..." + i[end:])
		}

		//trim path to specified (or generated) batch output directory
		truncatedOut := "..." + strings.Replace(in.Outdir, outdir, filepath.Base(outdir), 1)

		if err != nil {
			out += fmt.Sprintf("FAIL: Input = %v \t ExitCode = %d \t Error = %s \n", truncatedInputs, exitCode, err.Error())
		} else {
			out += fmt.Sprintf("PASS: Input = %v \t ExitCode = %d \t Output = %s \n", truncatedInputs, exitCode, truncatedOut)
		}
	}

	util.InitPrinter(false)
	util.PrintUtil("%v", out)

	return err
}

//PrintBatchUsage prints the seed batch usage arguments, then exits the program
func PrintBatchUsage() {
	util.PrintUtil( "\nUsage:\tseed batch -in IMAGE_NAME [OPTIONS] \n")

	util.PrintUtil( "\nRuns Docker image defined by seed spec.\n")

	util.PrintUtil( "\nOptions:\n")
	util.PrintUtil( "  -%s -%s Docker image name to run\n",
		constants.ShortImgNameFlag, constants.ImgNameFlag)
	util.PrintUtil( "  -%s  -%s Optional file specifying input keys and file mapping for batch processing. Supersedes directory flag.\n",
		constants.ShortBatchFlag, constants.BatchFlag)
	util.PrintUtil( "  -%s  -%s Alternative to batch file.  Specifies a directory of files to batch process (default is current directory)\n",
		constants.ShortJobDirectoryFlag, constants.JobDirectoryFlag)
	util.PrintUtil( "  -%s  -%s \t Specifies the key/value setting values of the seed spec in the format SETTING_KEY=VALUE\n",
		constants.ShortSettingFlag, constants.SettingFlag)
	util.PrintUtil( "  -%s  -%s \t Specifies the key/value mount values of the seed spec in the format MOUNT_KEY=HOST_PATH\n",
		constants.ShortMountFlag, constants.MountFlag)
	util.PrintUtil( "  -%s  -%s \t Job Output Directory Location\n",
		constants.ShortJobOutputDirFlag, constants.JobOutputDirFlag)
	util.PrintUtil( "  -%s \t\t Automatically remove the container when it exits (docker run --rm)\n",
		constants.RmFlag)
	util.PrintUtil( "  -%s  -%s \t External Seed metadata schema file; Overrides built in schema to validate side-car metadata files\n",
		constants.ShortSchemaFlag, constants.SchemaFlag)
	panic(util.Exit{0})
}

func getOutputDir(outputDir, imageName string) string {
	if outputDir == "" {
		outputDir = "batch-" + imageName + "-" + time.Now().Format(time.RFC3339)
		outputDir = strings.Replace(outputDir, ":", "_", -1)
	}

	outdir := util.GetFullPath(outputDir, "")

	// Check if outputDir exists. Create if not
	if _, err := os.Stat(outdir); os.IsNotExist(err) {
		// Create the directory
		// Didn't find the specified directory
		util.PrintUtil( "INFO: %s not found; creating directory...\n",
			outdir)
		os.Mkdir(outdir, os.ModePerm)
	}
	return outdir
}

func ProcessDirectory(seed objects.Seed, batchDir, outdir string) ([]BatchIO, error) {
	key := ""
	unrequired := ""
	for _, f := range seed.Job.Interface.Inputs.Files {
		if f.Multiple {
			continue
		}
		if f.Required {
			if key != "" {
				return nil, errors.New("ERROR: Multiple required inputs are not supported when batch processing directories.")
			}
			key = f.Name
		} else if unrequired == "" {
			unrequired = f.Name
		}
	}

	if key == "" {
		key = unrequired
	}

	if key == "" {
		return nil, errors.New("ERROR: Could not determine which input to use from Seed manifest.")
	}

	files, err := ioutil.ReadDir(batchDir)
	if err != nil {
		return nil, err
	}

	batchIO := []BatchIO{}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fileDir := filepath.Join(outdir, file.Name())
		filePath := filepath.Join(batchDir, file.Name())
		fileInputs := []string{}
		fileInputs = append(fileInputs, key+"="+filePath)
		row := BatchIO{fileInputs, fileDir}
		batchIO = append(batchIO, row)
	}

	util.PrintUtil("Batch Input Dir = %v \t Batch Output Dir = %v \n", batchDir, outdir)

	return batchIO, err
}

func ProcessBatchFile(seed objects.Seed, batchFile, outdir string) ([]BatchIO, error) {
	lines, err := util.ReadLinesFromFile(batchFile)
	if err != nil {
		return nil, err
	}

	if len(lines) == 0 {
		return nil, errors.New("ERROR: Empty batch file")
	}

	keys := strings.Split(lines[0], ",")
	extraKeys := keys

	if len(keys) == 0 || len(keys[0]) == 0{
		return nil, errors.New("ERROR: Empty keys list on first line of batch file.")
	}

	for _, f := range seed.Job.Interface.Inputs.Files {
		hasKey := util.ContainsString(keys, f.Name)
		if f.Required && !hasKey {
			msg := fmt.Sprintf("ERROR: Batch file is missing required key %v", f.Name)
			return nil, errors.New(msg)
		} else if !hasKey {
			fmt.Println("WARN: Missing input for key " + f.Name)
		}
		extraKeys = util.RemoveString(extraKeys, f.Name)
	}

	if len(extraKeys) > 0 {
		msg := fmt.Sprintf("WARN: These input keys don't match any specified keys in the Seed manifest: %v\n", extraKeys)
		fmt.Println(msg)
	}

	batchIO := []BatchIO{}
	for i, line := range lines {
		if i == 0 {
			continue
		}
		values := strings.Split(line, ",")
		fileInputs := []string{}
		inputNames := fmt.Sprintf("%d", i)
		for j, file := range values {
			if j > len(keys) {
				fmt.Println("WARN: More files provided than keys")
			}
			fileInputs = append(fileInputs, keys[j]+"="+file)
			inputNames += "-" + filepath.Base(file)
		}
		fileDir := filepath.Join(outdir, inputNames)
		row := BatchIO{fileInputs, fileDir}
		batchIO = append(batchIO, row)
	}

	util.PrintUtil("Batch Input = %s \t", batchFile)
	util.PrintUtil("Batch Output Dir = %s \n", outdir)

	return batchIO, err
}
