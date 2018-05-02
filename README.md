# rec

[![Go Report Card](https://goreportcard.com/badge/github.com/stts-se/rec)](https://goreportcard.com/report/github.com/stts-se/rec)  [![Build Status](https://travis-ci.org/stts-se/rec.svg?branch=master)](https://travis-ci.org/stts-se/rec)


Tiny demo server for capturing microphone sound via browser (Chrome, Firefox only) 

To build the server, you need to have Go installed.

Clone this repo.

cd rec/recserver

go get

__________________________

To compile and run the server:

go run *.go &lt;json-config-file&gt;

sample config file: config/config-sample.json

(or go build; ./recserver or go install; recserver)
