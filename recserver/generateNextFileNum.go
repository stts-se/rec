package main

import (
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/stts-se/rec"
)

var numRE = regexp.MustCompile("^.*_([0-9]{4})[.][^0-9]+$")

func generateNextFileNum(audioDir rec.AudioDir, fileNameBase string) string {
	res := "0001"

	highest := 0

	matches, err := filepath.Glob(filepath.Join(audioDir.Path(), fileNameBase+"_[0-9][0-9][0-9][0-9].*"))
	if err != nil {
		log.Printf("generateNextFileNum: failed to list files, returning default")
		return res
	}

	if len(matches) == 0 {
		return res
	}

	for _, m := range matches {
		numStr := numRE.FindStringSubmatch(m)
		if len(numStr) != 2 {
			log.Printf("generateNextFileNum failed to match number in file name: '%s'\n", m)
			continue
		}
		i, err := strconv.Atoi(numStr[1])
		if err != nil {
			log.Printf("generateNextFileNum failed to convert string to number: '%s' : %v\n", numStr, err)
			continue
		}
		if i > highest {
			highest = i
		}
	}

	highest++
	res = fmt.Sprintf("%04d", highest)
	return res
}
