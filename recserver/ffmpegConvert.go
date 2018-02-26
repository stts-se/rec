package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func ffmpegConvert(inFilePath, outFilePath string, removeInputFile bool) error {

	_, pErr := exec.LookPath("ffmpeg")
	if pErr != nil {
		log.Printf("ffmpegConvert failure : %v\n", pErr)
		return fmt.Errorf("ffmpegConvert failed to find the external 'ffmpeg' command : %v", pErr)
	}

	// '-y' means write over if output file already exists
	//HB cmd := exec.Command("ffmpeg", "-y", "-i", inFilePath, outFilePath)
	sampleRate := "16000"
	cmd := exec.Command("ffmpeg", "-y", "-i", inFilePath, "-ac", "1", "-ar", sampleRate, outFilePath)
	var out bytes.Buffer
	var sterr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &sterr

	err := cmd.Run()
	if err != nil {
		log.Printf("%s\n", sterr.String())
		return fmt.Errorf("ffmpegConvert failed running '%s': %v\n", cmd.Path, err)

	}

	// Command appears to have worked out.
	// Delete original file?
	if removeInputFile {
		err := os.Remove(inFilePath)
		if err != nil {
			log.Printf("failed to remove input file : %v\n", err)
		}
	}

	return nil
}
