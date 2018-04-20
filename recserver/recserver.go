package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"

	"github.com/stts-se/rec"
	"github.com/stts-se/rec/adapters"
	"github.com/stts-se/rec/aggregator"
	"github.com/stts-se/rec/config"
)

// TODO Mutex per user, i.e., use lock for a specific user(name), not all users.
// Something like mutexMap[userName]*sync.Mutex perhaps
var writeMutex sync.Mutex

func getParam(paramName string, r *http.Request) string {
	//fmt.Println("getParam r.URL", r.URL)
	res := r.FormValue(paramName)
	if res != "" {
		return res
	}
	res = r.PostFormValue(paramName)
	if res != "" {
		return res
	}
	vars := mux.Vars(r)
	return vars[paramName]
}

func index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "../recclient/index.html")
}

const defaultExtension = "wav"

func mimeType(ext string) string {
	if ext == "mp3" {
		return "audio/mpeg"
	}
	return fmt.Sprintf("audio/%s", ext)
}

func checkProcessInput(input rec.ProcessInput) error {
	var errMsg []string

	if strings.TrimSpace(input.UserName) == "" {
		errMsg = append(errMsg, "no value for 'username'")
	}
	if strings.TrimSpace(input.Text) == "" {
		errMsg = append(errMsg, "no value for 'text'")
	}
	if strings.TrimSpace(input.RecordingID) == "" {
		errMsg = append(errMsg, "no value for 'recording_id'")
	}

	if len(input.Audio.Data) == 0 {
		errMsg = append(errMsg, "no 'audio.data'")
	}
	if strings.TrimSpace(input.Audio.FileType) == "" {
		errMsg = append(errMsg, "no value for 'audio.file_type'")
	}

	if len(errMsg) > 0 {
		return fmt.Errorf("missing values in input JSON: %s", strings.Join(errMsg, " : "))
	}

	return nil
}

func process(w http.ResponseWriter, r *http.Request) {
	verb := getParam("verb", r)
	if verb == "true" {
		process0(w, r, true)
	} else {
		process0(w, r, false)
	}

}

// verbMode includes all component results, instead of just one single selected result
func process0(w http.ResponseWriter, r *http.Request, verbMode bool) {

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		msg := fmt.Sprintf("failed to read request body : %v", err)
		log.Println(msg)
		// or return JSON response with error message?
		//res.Message = msg
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	input := rec.ProcessInput{}
	err = json.Unmarshal(body, &input)
	if err != nil {
		msg := fmt.Sprintf("failed to unmarshal incoming JSON : %v", err)
		log.Println(msg)
		log.Println("incoming JSON string : %s", string(body))
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	if len(input.Weights) > 0 {
		log.Printf("user set weights: %-v\n", input.Weights)
	}

	err = checkProcessInput(input)
	if err != nil {
		msg := fmt.Sprintf("incoming JSON was incomplete: %v", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	log.Printf("GOT username: %s\ttext: %s\t recording id: %s\n", input.UserName, input.Text, input.RecordingID)

	audioDir := rec.AudioDir{BaseDir: audioDir, UserDir: input.UserName}
	// writeAudioFile uses writeMutex internally
	audioFile, err := writeAudioFile(audioDir, input)
	if err != nil {
		msg := fmt.Sprintf("failed writing audio file : %v", err)
		log.Print(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	res, err := analyzeAudio(audioFile.Path(), input, verbMode, config.MyConfig.FailOnRecogniserError)
	//log.Printf("analyzeAudio.res = %s err= %v\n", res, err)
	if err != nil {
		msg := err.Error()
		log.Print(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	// baseFileName := strings.TrimSuffix(path.Base(audioFile), path.Ext(audioFile))
	// audioRef = rec.AudioRef{Dir: audioDir, BaseName: baseFileName}
	audioRef := audioFile.BasePath

	// writeJSONInfoFile defined in writeJSONInfoFile.go
	// uses writeMutex internally

	err = writeJSONInfoFile(audioRef, input, res)
	if err != nil {
		msg := fmt.Sprintf("failed writing info file : %v", err)
		log.Print(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	log.Printf("recserver result below:")
	log.Printf("%s\n", res)
	for _, r := range res.ComponentResults {
		log.Printf("%s\n", r.String())
	}
	resJSON, err := json.Marshal(res)
	if err != nil {
		msg := fmt.Sprintf("failed to marshal response : %v", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(resJSON))

}

type recresforchan struct {
	resp  rec.RecogniserResponse
	err   error
	index int
}

func runRecogniserChan(accres chan recresforchan, rc config.Recogniser, index int, wavFilePath string, input rec.ProcessInput) {
	log.Printf("running recogniser %s", rc.LongName())
	var res rec.RecogniserResponse
	var err error
	switch rc.Type {
	case config.Tensorflow:
		res, err = adapters.RunTensorflowCommand(rc, wavFilePath, input)
	case config.KaldiGStreamer:
		res, err = adapters.RunGStreamerKaldiFromURL(rc, wavFilePath, input)
	case config.PocketSphinx:
		res, err = adapters.CallExternalPocketsphinxDecoderServer(rc, wavFilePath, input)
	default:
		err = fmt.Errorf("unknown recogniser type: %s", rc.Type)
	}
	rchan := recresforchan{resp: res, err: err, index: index}
	accres <- rchan
	if err != nil {
		log.Printf("completed recogniser %s with an error : %v", rc.LongName(), err)
	} else {
		log.Printf("completed recogniser %s => %v", rc.LongName(), res)
	}
}

// runs parallell calls (using chan)
func analyzeAudio(audioFile string, input rec.ProcessInput, verbMode bool, failOnRecogError bool) (rec.ProcessResponse, error) {
	//fmt.Printf("analyzeAudio input : %s\n", input)
	var accres = make(chan recresforchan)
	var n = 0
	for index, rc := range config.MyConfig.EnabledRecognisers() {
		n++
		go runRecogniserChan(accres, rc, index, audioFile, input)
	}
	nRecs := len(config.MyConfig.EnabledRecognisers())
	res := make([]rec.RecogniserResponse, nRecs, nRecs)

	for i := 0; i < n; i++ {
		rr := <-accres
		res[rr.index] = rr.resp
		if rr.err != nil && failOnRecogError {
			return rec.ProcessResponse{}, rr.err
		}
	}

	final, err := aggregator.CombineResults(input, res, verbMode)
	if err != nil {
		return rec.ProcessResponse{}, fmt.Errorf("failed to combine results : %v", err)
	}
	return final, nil
}

type audioResponse struct {
	FileType string `json:"file_type"`
	Data     string `json:"data"`
	Message  string `json:"message"`
}

// TODO Protect with mutex?
func getAudio(w http.ResponseWriter, r *http.Request) {
	var res audioResponse
	vars := mux.Vars(r)
	userName := vars["username"]
	utteranceID := vars["utterance_id"]

	// // TODO Remove the "noise reduced" audio variants?
	// noiseRedS := getParam("noise_red", r)
	// useNoiseReduction := false
	// if onRegexp.MatchString(noiseRedS) {
	// 	useNoiseReduction = true
	// }
	// if useNoiseReduction {
	// 	msg := fmt.Sprintf("get_audio: noise_red option is deprecated")
	// 	log.Print(msg)
	// 	http.Error(w, msg, http.StatusBadRequest)
	// 	return
	// }
	var ext = vars["ext"]
	if ext == "" {
		ext = defaultExtension
	}
	audioFile := rec.NewAudioFile(audioDir, userName, utteranceID, "."+ext)
	_, err := os.Stat(audioFile.Path())
	if os.IsNotExist(err) {

		// No exact match of file name. Try to list files with same base name + running number
		basePath := filepath.Join(audioDir, userName, utteranceID)
		files, err := filepath.Glob(basePath + "_[0-9][0-9][0-9][0-9]." + ext)
		if err != nil {
			log.Printf("getAudio: problem listing files : %v\n", err)
		}
		highest := 0
		for _, f := range files {

			// numRE defined in generateNextFileNum
			numStr := numRE.FindStringSubmatch(f)
			if len(numStr) != 2 {
				log.Printf("getAudio: failed to match number in file name: '%s'\n", f)
				continue
			}
			n, err := strconv.Atoi(numStr[1])
			if err != nil {
				log.Printf("getAudio: failed to convert string to number: '%s' : %v\n", numStr, err)
				continue
			}

			if n > highest {
				highest = n
			}
		}

		if highest == 0 {
			msg := fmt.Sprintf("get_audio: no audio for user '%s'", userName)
			log.Print(msg)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}

		// We have found a matching file with the highest running number
		runningNum := fmt.Sprintf("_%04d", highest)
		utteranceID = utteranceID + runningNum

		audioFile = rec.NewAudioFile(audioDir, userName, utteranceID, "."+ext)

	}

	bytes, err := ioutil.ReadFile(audioFile.Path())
	if err != nil {
		msg := fmt.Sprintf("get_audio: failed to read audio file : %v", err)
		log.Print(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	data := base64.StdEncoding.EncodeToString(bytes)

	res.FileType = mimeType(ext)
	res.Data = data

	resJSON, err := prettyMarshal(res)
	if err != nil {
		msg := fmt.Sprintf("get_audio: failed to create JSON from struct : %v", res)
		log.Print(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s\n", string(resJSON))
}

func pingRecognisers(w http.ResponseWriter, r *http.Request) {
	res, err := testRecognisers(false)
	if err != nil {
		msg := fmt.Sprintf("server error : %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	html := "<table style=\"border-collapse: separate; border-spacing: 5px;\"><thead><tr><td><b>Server name</b></td><td><b>Status</b></td><td><b>Message</b></td></tr></thead><tbody>"
	for _, cr := range res.ComponentResults {
		var status string
		if cr.Status == true {
			status = "<font style=\"color:green\">OK</font>"
		} else {
			status = "<font style=\"color:red\">Not OK</font>"
		}
		html += fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%s</td></tr>", cr.Source, status, cr.Message)
	}
	html += "</tbody></table>"
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "%s\n", html)
}

func testRecognisers(failOnRecogError bool) (rec.ProcessResponse, error) {
	var err error
	log.Println("=== RUNNING PING RECOGNISERS ====")
	fileName := filepath.Join(audioDir, "silence_used_for_recserver_init_tests.wav")
	input := rec.ProcessInput{
		UserName:    "tmpuser0",
		Text:        "_silence_",
		RecordingID: "tmprecid0",
		Audio:       rec.Audio{Data: "", FileType: "audio/wav"}}
	verb := true
	res, err := analyzeAudio(fileName, input, verb, failOnRecogError)
	if err != nil {
		log.Printf("testRecognisers() failed : %v\n", err)
	} else {
		log.Println("testRecognisers() success")
	}
	log.Println("=== COMPLETED PING RECOGNISERS ====")
	return res, err
}

// The path to the directory where audio files are saved
var audioDir string

// Name of subdirectory in which to put the original input audio file
// recieved from client, before re-coding it into 16 kHz mono wav.
var inputAudioDir = "input_audio"

// This is filled in by main, listing the URLs handled by the router,
// so that these can be shown in the generated docs.
var walkedURLs []string

func main() {

	// /* print config sample instance to json */
	// fmt.Println(config.ConfigSample.PrettyString())
	// os.Exit(1)

	if len(os.Args) != 2 {
		fmt.Println("USAGE: go run recserver.go <json-config-file>")
		fmt.Println("sample config file: config/config-sample.json")
		os.Exit(1)
	}

	cfgFile := os.Args[1]
	cfg, cErr := config.NewConfig(cfgFile)
	if cErr != nil {
		log.Printf("Exiting. Failed to read config file : %v", cErr)
		os.Exit(1)
	}
	if len(cfg.Recognisers) == 0 {
		log.Printf("Exiting. No recognisers defined in config file : %s", cfgFile)
		os.Exit(1)
	}
	config.MyConfig = cfg
	log.Printf("Loaded recognisers from config: " + strings.Join(cfg.RecogniserNames(), ", "))

	if !validAudioFileExtension(defaultExtension) {
		log.Printf("Exiting! Unknown default audio file extension: %s", defaultExtension)
		os.Exit(1)
	}

	if !ffmpegEnabled() {
		log.Printf("Exiting! %s is required! Please install.", ffmpegCmd)
		os.Exit(1)
	}

	if !soxEnabled() {
		log.Printf("Exiting! %s is required! Please install.", soxCmd)
		os.Exit(1)
	}

	// TODO return text prompt etc as well

	//TODO Check Go ffmpeg, or similar, bindings instead of
	// external call

	audioDir = config.MyConfig.AudioDir
	log.Printf("recserver audioDir: %s\n", audioDir)
	_, sErr := os.Stat(audioDir)
	if os.IsNotExist(sErr) {
		os.Mkdir(audioDir, os.ModePerm)
	}

	//func loadUtteranceLists defined in getUtterance.go
	err := loadUtteranceLists(audioDir)
	if err != nil {
		msg := fmt.Sprintf("failed to load user utterance lists : %v", err)
		log.Print(msg)
		os.Exit(1)
	}
	//log.Printf("recserver Loaded utts\n")
	//uttLists = uls

	docs := make(map[string]string)

	p := config.MyConfig.ServerPort
	r := mux.NewRouter()
	r.StrictSlash(true)
	r.HandleFunc("/rec/", index)
	r.HandleFunc("/rec/process/", process).Methods("POST")
	docs["/rec/process/"] = "send param verb=true for verbose response"

	// generateDoc is definied in file generateDoc.go
	r.HandleFunc("/rec/doc/", generateDoc).Methods("GET")

	// TODO Should this rather be a POST request?
	r.HandleFunc("/rec/get_audio/{username}/{utterance_id}/{ext}", getAudio).Methods("GET")
	r.HandleFunc("/rec/get_audio/{username}/{utterance_id}", getAudio).Methods("GET") // with default extension

	// audioproc.go
	// r.HandleFunc("/rec/build_spectrogram/{username}/{utterance_id}/{ext}", buildSpectrogram).Methods("GET")
	// r.HandleFunc("/rec/build_spectrogram/{username}/{utterance_id}", buildSpectrogram).Methods("GET")
	//r.HandleFunc("/rec/analyse_audio/{username}/{utterance_id}/{ext}", analyseAudio).Methods("GET")
	//r.HandleFunc("/rec/analyse_audio/{username}/{utterance_id}", analyseAudio).Methods("GET")

	// Defined in getUtterance.go
	r.HandleFunc("/rec/get_next_utterance/{username}", getNextUtterance).Methods("GET")
	r.HandleFunc("/rec/get_previous_utterance/{username}", getPreviousUtterance).Methods("GET")

	r.HandleFunc("/rec/admin/ping_recognisers", pingRecognisers).Methods("GET")

	// List route URLs to use as simple on-line documentation
	r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		t, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		if info, ok := docs[t]; ok {
			t = fmt.Sprintf("%s - %s", t, info)
		}
		walkedURLs = append(walkedURLs, t)
		return nil
	})
	// Add paths that don't need to be in the generated
	// documentation after the r.Walk above

	// Defined in admin.go
	r.HandleFunc("/rec/admin/list_users", listUsers).Methods("GET")
	r.HandleFunc("/rec/admin/add_user/{username}", addUser).Methods("GET")
	//r.HandleFunc("/rec/admin/delete_user/{username}", deleteUser).Methods("GET")
	//r.HandleFunc("/rec/admin/get_utts/{username}", getUtts).Methods("GET")
	//r.HandleFunc("/rec/admin/list_files/{username}", listFiles).Methods("GET")

	// for ngrok access
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "%s\n", "server up and running")
	})

	// see navigatedemo.go
	r.HandleFunc("/rec/navigatedemo", navigateDemo)

	r.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "favicon.ico")
	})

	r.PathPrefix("/rec/recclient/").Handler(http.StripPrefix("/rec/recclient/", http.FileServer(http.Dir("../recclient"))))

	ps := fmt.Sprintf("%d", p)
	//addr := fmt.Sprintf("127.0.0.1:%s", ps) // access only from localhost
	addr := fmt.Sprintf(":%s", ps) // external access
	srv := &http.Server{
		Handler: r,
		Addr:    addr,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Printf("rec server started on %s/rec\n", addr)
	_, err = testRecognisers(true)
	if err != nil {
		log.Printf("Exiting! Recogniser tests failed : %v", err)
		os.Exit(1)
	}
	log.Fatal(srv.ListenAndServe())

	fmt.Println("No fun")
}
