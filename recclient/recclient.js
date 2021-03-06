// See:
// https://developer.mozilla.org/en-US/docs/Web/API/MediaStream_Recording_API 
// https://mozdevs.github.io/MediaRecorder-examples/record-live-audio.html
// https://github.com/mdn/voice-change-o-matic

//'use strict'

// //
// //
//
// https://github.com/mdn/voice-change-o-matic/blob/gh-pages/scripts/app.js:
//
// fork getUserMedia for multiple browser versions, for those
// that need prefixes

navigator.getUserMedia = (navigator.getUserMedia ||
                          navigator.webkitGetUserMedia ||
                          navigator.mozGetUserMedia ||
                          navigator.msGetUserMedia);

// navigator.mediaDevices.getUserMedia = (navigator.mediaDevices.getUserMedia ||
// 					navigator.mediaDevices.webkitGetUserMedia ||
// 					navigator.mediaDevices.mozGetUserMedia ||
// 					navigator.mediaDevices.msGetUserMedia);




let recButton, stopAndSendButton, /*stopButton,*/ getAudioButton, sendButton;
let baseURL = window.location.origin +"/rec"; // TODO: should probably be : let baseURL = window.location.protocol + '//' + window.location.host + window.location.pathname.replace(/\/$/g,"");

console.log(baseURL);
var currentBlob;
var recorder;
// var wavesurfer;
let user = "anon";

window.onload = function () {

    var url = new URL(document.URL);
    
    recButton = document.getElementById('rec');
    recButton.addEventListener('click', startRecording);
    recButton.disabled = false;
    
    // stopButton = document.getElementById('stop');
    // stopButton.addEventListener('click', stopRecording);
    // stopButton.disabled = true;
    
    stopAndSendButton = document.getElementById('stopandsend');
    stopAndSendButton.addEventListener('click', stopAndSend);
    stopAndSendButton.disabled = true;
    
    console.log("navigator.mediaDevices:", navigator.mediaDevices);
    mediaAccess = navigator.mediaDevices.getUserMedia({'audio': true, video: false});
    console.log("navigator.mediaDevices.getUserMedia:", mediaAccess);
    
    //navigator.mediaDevices.getUserMedia({'audio': true, video: false}).then(function(stream) {
	mediaAccess.then(function(stream) {
	console.log("navigator.mediaDevices.getUserMedia was called")
	source = audioCtx.createMediaStreamSource(stream);
        source.connect(analyser);
	visualize();	
	recorder = new MediaRecorder(stream);
	recorder.addEventListener('dataavailable', function (evt) {
	    updateAudio(evt.data);
	    sendAndReceiveBlob();
	});
	
	recorder.onstop = function(evt) {}
    });
    //navigator.mediaDevices.getUserMedia({'audio': true, video: false}).catch(function(err) {
    mediaAccess.catch(function(err) {
	console.log("error from getUserMedia:", err);
	alert("Couldn't initialize recorder: " + err);
    });

};

function startRecording() {

    if (recorder == null) {
	msg = "Cannot record -- recorder is undefined"
	console.log(msg);
	alert(msg);
    }
    

    // TODO set max recording time limit
    
    recButton.disabled = true;
    // stopButton.disabled = false;
    stopAndSendButton.disabled = false;
    recorder.start();

    clearResponse();
    countDown();
}
// function stopRecording() {
//     console.log("stopRecording()");

//     recButton.disabled = false;
    
//     // make MediaRecorder stop recording
//     // eventually this will trigger the dataavailable event
//     recorder.cancel();
//     stopAndSendButton.disabled = false;
//     // stopButton.disabled = false;
//     clearInterval(setIntFunc);
//     document.getElementById("rec_progress").value = "0";
// }

function stopAndSend() {
    console.log("stopAndSend()");

    recButton.disabled = false;
    
    // make MediaRecorder stop recording
    // eventually this will trigger the dataavailable event
    recorder.stop();
    stopAndSendButton.disabled = true;
    // stopButton.disabled = false;
    clearInterval(setIntFunc);
    document.getElementById("rec_progress").value = "0";
}

var setIntFunc;

function countDown() {
    var max = 5;
    let tick = 10;
    var dur = 0;

    document.getElementById("rec_progress").value = ""+ dur;
    
    setIntFunc = setInterval(function() {

	dur = dur + (tick / 1000);
	
	document.getElementById("rec_progress").value = ""+ dur;
	
	if (dur > max + 1) {
	    clearInterval(setIntFunc);
	    stopAndSendButton.click();
	};
	
    }, tick);

    setIntFunc;
}

function sendAndReceiveBlob() {
    console.log("sendAndReceiveBlob()");

    var onLoadEndFunc = function (data) {
	//console.log("onLoadEndFunc data ", data);
	clearResponse(); // originally called before sending
	stopAndSendButton.disabled = true; // originally called after sendJSON
	console.log("onLoadEndFunc|STATUS : "+ data.target.status + "/" + data.target.statusText);
	console.log("onLoadEndFunc|RESPONSE : "+ data.target.responseText);
	if (data.target.status === 200) {
	    showResponse(data.target.responseText);
	} else {
	    showError(data, document.getElementById("recording_id").innerHTML);
	}
    };

    AUDIO.sendBlob(currentBlob,
		   "", //TODO scrtptnam
		   user,
		   "", // input text
		   "", // rec id
		   onLoadEndFunc);
}

function showError(data, recordingId) {
    var resp = document.getElementById("response");
    clearResponse();

//     type processResponse struct {
// 	Ok                bool    `json:"ok"`
// 	Confidence        float64 `json:"confidence"`
// 	RecognitionResult string  `json:"recognition_result"`
// 	RecordingID       string  `json:"recording_id"`
// 	Message           string  `json:"message"`
// }

    var json = {
	"ok": false,
	"confidence": -1,
	"recognition_result": "",
	"recording_id": recordingId,
	"message": data.target.status + "/" + data.target.statusText + ": " + data.target.responseText.trim(),
    };
    
    var j = JSON.stringify(json, null, '\t');
    
    resp.innerHTML = j;
}

function showResponse(json) {
    var resp = document.getElementById("response");
    clearResponse();
    var o = JSON.parse(json);
    var j = JSON.stringify(o, null, '\t');
    j = j.replace(/("input_confidence": {)\n\s*/g, "$1");
    j = j.replace(/("(?:config[^"]*|combined|recogniser|user)": [0-9.]+,?)\n\s*(}?)/g, "$1$2 ");
    j = j.replace(/} ,/g, "},");
    j = j.replace(/("(?:config[^"]*|combined|recogniser|user|confidence)": [0-9].[0-9]{4})[0-9]+/g, "$1");
    resp.innerHTML = j;
};

function clearResponse() {
    document.getElementById("response").innerHTML = "";
}

// function showJSAudioPane() {
//     console.log("showJSAudioPane()");
//     ele = document.getElementById("js-wavesurfer");
//     ele.style.visibility = "visible";
// }

// function hideJSAudioPane() {
//     console.log("hideJSAudioPane()");
//     ele = document.getElementById("js-wavesurfer");
//     ele.style.visibility = "hidden";
// }


function updateAudio(blob) {
    console.log("updateAudio()", blob.size);
    //console.log("UPDATE AUDIO: "+ blob.size);
    //console.log("UPDATE AUDIO: "+ blob.type);

    currentBlob = blob;
    
    var audio = document.getElementById('audio');
    // use the blob from the MediaRecorder as source for the audio tag
    audio.src = URL.createObjectURL(blob);

};

function uint8ArrayToArrayBuffer(input) {
    var res = new ArrayBuffer(input.length);
    for(var i = 0; i < input.length; i++) {
        res[i] = input[i];
    }

    return res;
}

function getAudio() {

    console.log("getAudio()");
    //hideJSAudioPane();
    
    let userName = document.getElementById('username2').value;
    let utteranceID = document.getElementById('recording_id2').value;
    let audio = document.getElementById('audio_from_server');

    let audioURL = baseURL + "/get_audio/" + userName + "/" + utteranceID;
    console.log("getAudio URL " + audioURL);
    let xhr = new XMLHttpRequest();
    xhr.open("GET", audioURL, true);
    xhr.setRequestHeader('Content-Type', 'application/json; charset=UTF-8');

    // TODO error handling
    
    
    xhr.onloadend = function () {
     	// done
	console.log("STATUS: "+ xhr.statusText);
	audio.src = "";
	let resp = JSON.parse(xhr.response);

	// https://stackoverflow.com/questions/16245767/creating-a-blob-from-a-base64-string-in-javascript#16245768
	let byteCharacters = atob(resp.data);  

	var byteNumbers = new Array(byteCharacters.length);
	for (var i = 0; i < byteCharacters.length; i++) {
	    byteNumbers[i] = byteCharacters.charCodeAt(i);
	}
	var byteArray = new Uint8Array(byteNumbers);

	let blob = new Blob([byteArray], {'type' : resp.file_type});
	audio.src = URL.createObjectURL(blob);
	console.log("getAudio onloadend")
	//audio.play();

	//wavesurfer.loadBlob(blob);
    };
    
    xhr.send();

}

//
// https://github.com/mdn/voice-change-o-matic/blob/gh-pages/scripts/app.js:
//
// set up forked web audio context, for multiple browsers
// window. is needed otherwise Safari explodes

var audioCtx = new (window.AudioContext || window.webkitAudioContext)();
var source;
var stream;

//set up the different audio nodes we will use for the app

var analyser = audioCtx.createAnalyser();
analyser.minDecibels = -90;
analyser.maxDecibels = -10;
analyser.smoothingTimeConstant = 0.85;

// set up canvas context for visualizer

var canvas = document.querySelector('.visualizer');
var canvasCtx = canvas.getContext("2d");

var intendedWidth = document.querySelector('.wrapper').clientWidth;

canvas.setAttribute('width',intendedWidth / 2);

var drawVisual;

function visualize() {
    var WIDTH = canvas.width;
    var HEIGHT = canvas.height;
    
    
    analyser.fftSize = 256;
    var bufferLengthAlt = analyser.frequencyBinCount;
    //console.log(bufferLengthAlt);
    var dataArrayAlt = new Uint8Array(bufferLengthAlt);
    
    canvasCtx.clearRect(0, 0, WIDTH, HEIGHT);
    
    var draw = function() {
	drawVisual = requestAnimationFrame(draw);
	
	analyser.getByteFrequencyData(dataArrayAlt);
	
	canvasCtx.fillStyle = 'rgb(0, 0, 0)';
	canvasCtx.fillRect(0, 0, WIDTH, HEIGHT);
	
	var barWidth = (WIDTH / bufferLengthAlt) * 2.5;
	var barHeight;
	var x = 0;
	
	// Only draw frequency bars when recording
	// When recording, the stop button is enabled 
	if (stopAndSendButton.disabled === false) { 
	    for(var i = 0; i < bufferLengthAlt; i++) {
		barHeight = dataArrayAlt[i];
		
		canvasCtx.fillStyle = 'rgb(' + (barHeight+100) + ',50,50)';
		canvasCtx.fillRect(x,HEIGHT-barHeight/2,barWidth,barHeight/2);
		
		x += barWidth + 1;
	    }
	};
    };
    
    draw(); 
}

