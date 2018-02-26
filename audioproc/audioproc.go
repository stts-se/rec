package audioproc

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	//"strings"
)

func execCmd(cmd *exec.Cmd) (bytes.Buffer, error) {
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	//log.Printf("execCmd: %s", strings.Join(cmd.Args, " "))

	return stderr, cmd.Run()
}

func Analyse(inFilePath string) (map[string]string, error) {
	res := make(map[string]string)
	res["input_file"] = inFilePath
	return res, nil
}

const soxCmd = "sox"

func SoxEnabled() bool {
	funcId := "SoxEnabled"
	_, pErr := exec.LookPath(soxCmd)
	if pErr != nil {
		log.Printf("%s failure : %v\n", funcId, pErr)
		return false
	}
	return true
}

func NoiseReduce(inFilePath, outFilePath string) error {

	noiseProfPath := inFilePath + "-noiseprof"

	funcId := "NoiseReduce"

	_, pErr := exec.LookPath(soxCmd)
	if pErr != nil {
		log.Printf("%s failure : %v\n", funcId, pErr)
		return fmt.Errorf("%s failed to find the external 'sox' command : %v", funcId, pErr)
	}

	// (1) noise profile
	// sox /tmp/rec_0001.wav -n noiseprof /tmp/rec_0001-tmp.noiseprof
	cmd := exec.Command(soxCmd, inFilePath, "-n", "noiseprof", noiseProfPath)
	stderr, err := execCmd(cmd)
	if err != nil {
		log.Printf("%s\n", stderr.String())
		return fmt.Errorf("%s failed running '%s': %v\n", funcId, cmd.Path, err)

	}

	// (2) noise reduction
	// sox /tmp/rec_0001.wav /tmp/rec_0001-tmp.wav noisered /tmp/rec_0001-tmp.noiseprof 0.21
	cmd = exec.Command(soxCmd, inFilePath, outFilePath, "noisered", noiseProfPath, "0.21")
	stderr, err = execCmd(cmd)
	if err != nil {
		log.Printf("%s\n", stderr.String())
		return fmt.Errorf("%s failed running '%s': %v\n", funcId, cmd.Path, err)

	}

	// (3) remove noise profile
	err = os.Remove(noiseProfPath)
	if err != nil {
		log.Printf("failed to remove file : %v\n", err)
	}

	return nil
}

var genTmpFileRE = regexp.MustCompile("([.][^.]+)$")

func BuildSoxSpectrogram(inFilePath, outFilePath string, useNoiseReduction bool) error {

	noiseRedFilePath := genTmpFileRE.ReplaceAllString(inFilePath, "-tmp$1")
	noiseProfPath := genTmpFileRE.ReplaceAllString(inFilePath, "-tmp.noiseprof")

	funcId := "BuildSoxSpectrogram"

	//log.Printf("%s input %s %s %v", funcId, inFilePath, outFilePath, useNoiseReduction)

	_, pErr := exec.LookPath(soxCmd)
	if pErr != nil {
		log.Printf("%s failure : %v\n", funcId, pErr)
		return fmt.Errorf("%s failed to find the external 'sox' command : %v", funcId, pErr)
	}

	specInputFile := inFilePath
	if useNoiseReduction {
		specInputFile = noiseRedFilePath

		// (1) noise profile
		// sox /tmp/rec_0001.wav -n noiseprof /tmp/rec_0001-tmp.noiseprof
		cmd := exec.Command(soxCmd, inFilePath, "-n", "noiseprof", noiseProfPath)
		stderr, err := execCmd(cmd)
		if err != nil {
			log.Printf("%s\n", stderr.String())
			return fmt.Errorf("%s failed running '%s': %v\n", funcId, cmd.Path, err)

		}

		// (2) noise reduction
		// sox /tmp/rec_0001.wav /tmp/rec_0001-tmp.wav noisered /tmp/rec_0001-tmp.noiseprof 0.21
		cmd = exec.Command(soxCmd, inFilePath, noiseRedFilePath, "noisered", noiseProfPath, "0.21")
		stderr, err = execCmd(cmd)
		if err != nil {
			log.Printf("%s\n", stderr.String())
			return fmt.Errorf("%s failed running '%s': %v\n", funcId, cmd.Path, err)

		}
	}

	// (3) spectrogram | -m for monochrome, -l for light background
	cmd := exec.Command(soxCmd, specInputFile, "-n" /*"rate", "7k",*/, "spectrogram", "-m", "-l", "-x", "1100", "-y", "300", "-z", "90", "-o", outFilePath)
	stderr, err := execCmd(cmd)
	if err != nil {
		log.Printf("%s\n", stderr.String())
		return fmt.Errorf("%s failed running '%s': %v\n", funcId, cmd.Path, err)

	}

	if useNoiseReduction {
		err = os.Remove(noiseRedFilePath)
		if err != nil {
			log.Printf("failed to remove noise reduced tmp file : %v\n", err)
		}
		err = os.Remove(noiseProfPath)
		if err != nil {
			log.Printf("failed to remove noise profile tmp file : %v\n", err)
		}
	}

	return nil
}
