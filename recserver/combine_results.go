package main

import (
	"github.com/stts-se/rec"
	"sort"
	"strings"
)

func combineResults(input rec.ProcessInput, results []rec.ProcessResponse) rec.ProcessResponse {
	sorter := func(i, j int) bool {
		if results[i].Ok && results[j].Ok {
			return results[i].Confidence > results[j].Confidence
		} else {
			return results[i].Ok
		}
	}
	sort.Slice(results, sorter)
	var r1 rec.ProcessResponse
	if len(results) > 0 {
		r1 = results[0]
		sources := []string{}
		for _, r := range results {
			sources = append(sources, r.Source())
		}
		r1.Message = "Sources: " + strings.Join(sources, ", ")
	} else {
		r1 = rec.ProcessResponse{Ok: false,
			RecordingID:       input.RecordingID,
			Message:           "No result from server",
			RecognitionResult: ""}
	}
	return r1
}
