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


// //
// //


let recButton, stopButton, sendButton, getAudioButton, getSpecButton, prevButton, nextButton;
let baseURL = window.location.origin +"/rec";
console.log(baseURL);
var currentBlob;
var recorder;
var wavesurfer;

window.onload = function () {

    prevButton  = document.getElementById('prev_button');
    prevButton.addEventListener('click', getPrev)
    
    nextButton  = document.getElementById('next_button');
    nextButton.addEventListener('click', getNext)
    
    recButton = document.getElementById('rec');
    recButton.addEventListener('click', startRecording);
    recButton.disabled = false;
    
    // stopButton = document.getElementById('stop');
    // stopButton.addEventListener('click', stopRecording);
    // stopButton.disabled = true;
    
    stopButton = document.getElementById('stopandsend');
    stopButton.addEventListener('click', function() {
	stopRecording();
	//sendAndReceiveBlob(); // => added to dataavailable event listener below
    });
    stopButton.disabled = true;
    
    // stop+send not used as separate buttons, instead see stopButton also sends to server
    sendButton = document.getElementById('send');
    sendButton.addEventListener('click', sendAndReceiveBlob);
    sendButton.disabled = true;

    getAudioButton = document.getElementById('get_audio');
    getAudioButton.addEventListener('click', function() {
	getAudio();
	if (document.getElementById('get_audio_include_spectrogram').checked) {
	    getServerSpectrogram();
	}
    });

    
    navigator.mediaDevices.getUserMedia({'audio': true, video: false}).then(function(stream) {
	source = audioCtx.createMediaStreamSource(stream);
	//console.log("source.connect(analyser)");
        source.connect(analyser);
	visualize();	
	recorder = new MediaRecorder(stream);
	recorder.addEventListener('dataavailable', function (evt) {
	    updateAudio(evt.data);
	    sendAndReceiveBlob();
	});
	
	recorder.onstop = function(evt) {}
    });

    initWavesurferJS();

    initSoxStatus()
    
    // TODO Remove temporary initialization
    prevButton.click();

    // getAudioButton.click(); HL using this for quicker dev with spectrograms
};


function initSoxStatus() {
    var xhr = new XMLHttpRequest();
    xhr.open("GET", baseURL + "/enabled/sox" , true);

    xhr.onloadend = function () {
	let resp = JSON.parse(xhr.response);
	let enabled = (resp.enabled == "true")
	//console.log("initSoxStatus soxEnabledResp", resp)
	console.log("initSoxStatus soxEnabled", enabled)
	if (enabled) {
	    document.getElementById("get_audio_include_spectrogram").removeAttribute("disabled");
	    document.getElementById("noise_red_spectrogram").removeAttribute("disabled");
	} else {
	    document.getElementById("get_audio_include_spectrogram").setAttribute("disabled","disabled");
	    document.getElementById("noise_red_spectrogram").setAttribute("disabled","disabled");
	}
    };    
    xhr.send();
    
}

function getPrev() {

    document.getElementById("message").innerHTML = "";
    document.getElementById("num").innerHTML = "";
    
    // TODO Error check user name
    
    let userName = document.getElementById('username').value
    
    var xhr = new XMLHttpRequest();
    xhr.open("GET", baseURL + "/get_previous_utterance/" + userName , true);

    
    // TODO error handling
    
    
    xhr.onloadend = function () {
	
	let resp = JSON.parse(xhr.response);
	
	document.getElementById("recording_id").innerHTML = resp.recording_id;
	document.getElementById("text").innerHTML = resp.text;
	document.getElementById("num").innerHTML = resp.num +"/"+ resp.of;
	document.getElementById("message").innerHTML = resp.message;
    };
    
    xhr.send();

}

function initWavesurferJS() {
    // https://wavesurfer-js.org/doc/class/src/plugin/spectrogram.js~SpectrogramPlugin.html
    wavesurfer = WaveSurfer.create({
    	container: '#js-wavesurfer',
    	waveColor: '#6699FF',
    	progressColor: '#517acc', //'#46B54D',
    	labels: true,
    	//fftSamples: 256, //512,//1024,//1024,
    	controls: true,
    });
    
    wavesurfer.on('ready', function () {
    	//console.log("wavesurfer.js sample rate", wavesurfer.backend.ac.sampleRate);
    	var spectrogram = Object.create(WaveSurfer.Spectrogram);
    	spectrogram.init({
    	    wavesurfer: wavesurfer,
    	    container: "#js-wavesurfer-spectrogram",
    	    //fftSamples: 256, //512,//1024,//1024,
    	    labels: true,
    	});
    	var timeline = Object.create(WaveSurfer.Timeline);
    	timeline.init({
            wavesurfer: wavesurfer,
            container: "#js-wavesurfer-timeline",
    	    labels: true,
    	});
    	wavesurfer.play();
    });

    let maxWidth = "max-width: 1244px";
    document.getElementById("js-wavesurfer").setAttribute("style", maxWidth);
    document.getElementById("js-wavesurfer-spectrogram").setAttribute("style", maxWidth);
    document.getElementById("js-wavesurfer-timeline").setAttribute("style", maxWidth);
}

function getNext() {

    document.getElementById("num").innerHTML = "";
    document.getElementById("message").innerHTML = "";
    
    // TODO Error check user name

    let userName = document.getElementById('username').value
    
    var xhr = new XMLHttpRequest();
    xhr.open("GET", baseURL + "/get_next_utterance/" + userName , true);

    
    // TODO error handling
    
    
    xhr.onloadend = function () {
	
	let resp = JSON.parse(xhr.response);
	
	document.getElementById("recording_id").innerHTML = resp.recording_id;
	document.getElementById("text").innerHTML = resp.text;
	document.getElementById("num").innerHTML = resp.num +"/"+ resp.of;
	document.getElementById("message").innerHTML = resp.message;
    };
    
    xhr.send();
    
}

function startRecording() {

    // TODO set max recording time limit
    
    recButton.disabled = true;
    stopButton.disabled = false;
    sendButton.disabled = true;
    recorder.start();

    clearResponse();
    countDown();
}
function stopRecording() {
    console.log("stopRecording()");

    recButton.disabled = false;
    stopButton.disabled = true;
    
    // make MediaRecorder stop recording
    // eventually this will trigger the dataavailable event
    recorder.stop();
    sendButton.disabled = false;
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
	    stopButton.click();
	};
	
    }, tick);

    setIntFunc;
}

function sendAndReceiveBlob() {
    console.log("sendAndReceiveBlob()");

    var onLoadEndFunc = function (data) {
	//console.log("onLoadEndFunc data ", data);
	clearResponse(); // originally called before sending
	sendButton.disabled = true; // originally called after sendJSON
	console.log("onLoadEndFunc|STATUS : "+ data.target.status + "/" + data.target.statusText);
	console.log("onLoadEndFunc|RESPONSE : "+ data.target.responseText);
	if (data.target.status === 200) {
	    showResponse(data.target.responseText);
	} else {
	    showError(data, document.getElementById("recording_id").innerHTML);
	}
    };

    AUDIO.sendBlob(currentBlob,
	     document.getElementById("username").value,
	     document.getElementById("text").innerHTML,
	     document.getElementById("recording_id").innerHTML,
	     onLoadEndFunc);
}

function showError(data, recordingId) {
    var resp = document.getElementById("response");
    clearResponse();

//     type processResponse struct {
// 	Ok                bool    `json:"ok"`
// 	Confidence        float32 `json:"confidence"`
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
    
    resp.innerHTML = j;
};

function clearResponse() {
    document.getElementById("response").innerHTML = "";
}

function clearServerSpectrogram() {
    var ele = document.getElementById("spectrogram_from_server");
    if (ele != null)
	ele.removeAttribute("src");
}


function updateAudio(blob) {
    console.log("updateAudio()", blob.size);
    //console.log("UPDATE AUDIO: "+ blob.size);
    //console.log("UPDATE AUDIO: "+ blob.type);

    currentBlob = blob;
    
    var audio = document.getElementById('audio');
    // use the blob from the MediaRecorder as source for the audio tag
    audio.src = URL.createObjectURL(blob);
    audio.play();
    // var xhr = new XMLHttpRequest();
    // xhr.open('GET', audio.src, true);
    // xhr.responseType = 'blob';
    // xhr.onload = function(e) {
    // 	if (this.status == 200) {
    // 	    currentBlob = this.response;
    // 	    // myBlob is now the blob that the object URL pointed to.
    // 	}
    // };
    // xhr.send();
    
    
    //sendButton.disabled = false;
};

function getServerSpectrogram() {
    console.log("getServerSpectrogram()");
    let userName = document.getElementById('username2').value;
    let utteranceID = document.getElementById('recording_id2').value;
    let useNoiseRed = document.getElementById('noise_red_spectrogram').checked;
    let spec = document.getElementById('spectrogram_from_server');

    var xhr = new XMLHttpRequest();
    xhr.open("GET", baseURL + "/build_spectrogram/" + userName + "/" + utteranceID + "?noise_red=" + useNoiseRed, true);
    xhr.setRequestHeader('Content-Type', 'application/json; charset=UTF-8');
    
    // TODO error handling
    
    
    xhr.onloadend = function () {
     	// done
	console.log("STATUS: "+ xhr.statusText);
	spec.src = "";
	let resp = JSON.parse(xhr.response);

	//console.log("TODO: CHECK FILE TYPE: " + resp.file_type);
	
	// https://stackoverflow.com/questions/16245767/creating-a-blob-from-a-base64-string-in-javascript#16245768
	let byteCharacters = atob(resp.data);  

	var byteNumbers = new Array(byteCharacters.length);
	for (var i = 0; i < byteCharacters.length; i++) {
	    byteNumbers[i] = byteCharacters.charCodeAt(i);
	}
	var byteArray = new Uint8Array(byteNumbers);
	
	let blob = new Blob([byteArray], {'type' : resp.file_type});
	spec.src = URL.createObjectURL(blob);
    };

    xhr.send();   
}


function uint8ArrayToArrayBuffer(input) {
    var res = new ArrayBuffer(input.length);
    for(var i = 0; i < input.length; i++) {
        res[i] = input[i];
    }

    return res;
}

function getAudio() {

    console.log("getAudio()");
    clearServerSpectrogram();
    
    let userName = document.getElementById('username2').value;
    let utteranceID = document.getElementById('recording_id2').value;
    let audio = document.getElementById('audio_from_server');

    let audioURL = baseURL + "/get_audio/" + userName + "/" + utteranceID;
    console.log("getAudio URL " + audioURL);
    let xhr = new XMLHttpRequest();
    xhr.open("GET", audioURL, true);
    xhr.setRequestHeader('Content-Type', 'application/json; charset=UTF-8');

    let noiseRedSpec = document.getElementById('noise_red_spectrogram').checked;
    console.log("getAudio noiseRedSpec", noiseRedSpec);

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
	//audio.play();

	if (noiseRedSpec !== true) {
	    wavesurfer.loadBlob(blob);
	}
    };

    if (noiseRedSpec) {
	getAudioForSpectrogram(audioURL, noiseRedSpec);
    }
    
    xhr.send();

}

// TODO Clear old spectrogramme when no audio found for request/before call to generate new spectrogramme

function getAudioForSpectrogram(audioURL, noiseRedSpec) {
    console.log("getAudioForSpectrogram()");
    let audioURLForSpec = audioURL + "?noise_red=" + noiseRedSpec
    console.log("getAudio URL for spectrogram " + audioURLForSpec);
    let xhr = new XMLHttpRequest();
    xhr.open("GET", audioURLForSpec, true);
    xhr.setRequestHeader('Content-Type', 'application/json; charset=UTF-8');

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

	wavesurfer.loadBlob(blob);
    };
  
    xhr.send();

}

// //
// //
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
	if (stopButton.disabled === false) { 
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

