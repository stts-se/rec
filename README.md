# rec
Tiny demo server for capturing microphone sound via browser (Chrome, Firefox only) 

To build the server, you need to have Go installed.

Clone this repo.

cd rec/recserver

go get

__________________________

To compile and run the server:

go run *.go <json-config-file>

sample config file: config/config-sample.json

(or go build; ./recserver or go install; recserver)
