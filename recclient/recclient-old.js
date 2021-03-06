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




let recButton, stopButton, sendButton, getAudioButton, prevButton, nextButton;
let baseURL = window.location.origin +"/rec";
console.log(baseURL);
var currentBlob;
var recorder;
// var wavesurfer;

window.onload = function () {

    var url = new URL(document.URL);
    var user = url.searchParams.get('username');
    if (user != null && user != "") {
	document.getElementById('username').setAttribute('value',user);
	document.getElementById('username2').setAttribute('value',user);
	console.log("Setting user", user);
    }
    
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
    });
    stopButton.disabled = true;
    
    // stop+send not used as separate buttons, instead see stopButton also sends to server
    sendButton = document.getElementById('send');
    sendButton.addEventListener('click', sendAndReceiveBlob);
    sendButton.disabled = true;

    getAudioButton = document.getElementById('get_audio');
    getAudioButton.addEventListener('click', function() {
	getAudio();
    });


    console.log("navigator.mediaDevices:", navigator.mediaDevices);
    mediaAccess = navigator.mediaDevices.getUserMedia({'audio': true, video: false});
    console.log("navigator.mediaDevices.getUserMedia:", mediaAccess);
    
    //navigator.mediaDevices.getUserMedia({'audio': true, video: false}).then(function(stream) {
    mediaAccess.then(function(stream) {
	console.log("navigator.mediaDevices.getUserMedia was called")

	source = audioCtx.createMediaStreamSource(stream);
        source.connect(analyser);
	visualize();	
	//HB
	console.log("visualize was called")
	
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

    //initWavesurferJS();

    // TODO Remove temporary initialization

    //HB not sure about this..
    //maybe we can get the first utterance instead?
    prevButton.click();

    // getAudioButton.click(); HL using this for quicker dev with spectrograms
};


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
	document.getElementById("recording_id2").setAttribute('value',resp.recording_id);
	document.getElementById("text").innerHTML = resp.text;
	document.getElementById("num").innerHTML = resp.num +"/"+ resp.of;
	document.getElementById("message").innerHTML = resp.message;
    };

    xhr.send();

}

function initWavesurferJS() {
    // https://wavesurfer-js.org/doc/class/src/plugin/spectrogram.js~SpectrogramPlugin.html
    wavesurfer = WaveSurfer.create({
    	container: '#js-wavesurfer-wav',
    	waveColor: '#6699FF',
    	progressColor: '#517acc', //'#46B54D',
    	labels: true,
    	controls: true,
    });
    
    wavesurfer.on('ready', function () {
    	//console.log("wavesurfer.js sample rate", wavesurfer.backend.ac.sampleRate);
    	// var spectrogram = Object.create(WaveSurfer.Spectrogram);
    	// spectrogram.init({
    	//     wavesurfer: wavesurfer,
    	//     container: "#js-wavesurfer-spectrogram",
    	//     labels: true,
    	// });
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
    document.getElementById("js-wavesurfer-wav").setAttribute("style", maxWidth);
    //document.getElementById("js-wavesurfer-spectrogram").setAttribute("style", maxWidth);
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

    if (recorder == null) {
	msg = "Cannot record -- recorder is undefined"
	console.log(msg);
	alert(msg);
    }
    

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

    const scriptName = "default";
    AUDIO.sendBlob(currentBlob,
		   scriptName,
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

    //HB added
    var e = document.getElementById("recognition_result");
    if (e !== null) {
	e.innerHTML = o.recognition_result;
    }


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

    //sendButton.disabled = false;
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

//HB chrome gives a warning..
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

	//HB true to display the frequency bars, false for alternative display
	var drawFrequencyBars = false;
	
	// Only draw frequency bars when recording
	// When recording, the stop button is enabled 
	if (stopButton.disabled === false) { 
	    if ( drawFrequencyBars === true ) {
		//HB this displays the frequency bars
		for(var i = 0; i < bufferLengthAlt; i++) {
		    barHeight = dataArrayAlt[i];
		    //console.log(barHeight);
		    canvasCtx.fillStyle = 'rgb(' + (barHeight+100) + ',50,50)';
		    canvasCtx.fillRect(x,HEIGHT-barHeight/2,barWidth,barHeight/2);
		    
		    x += barWidth + 1;
		}
	    } else {
		//Alternative visualisation: simple amplitude meter
	        var values = 0;
		for (var i = 0; i < bufferLengthAlt; i++) {
		    values += (dataArrayAlt[i]);
		}
		
		var average = values / bufferLengthAlt;
		
		//console.log(Math.round(average - 40));

		barWidth = WIDTH;
		barHeight = average;
		canvasCtx.fillStyle = 'rgb(' + (barHeight+100) + ',50,50)';
		canvasCtx.fillRect(x,HEIGHT-barHeight/2,barWidth,barHeight/2);
		/*
		  canvasCtx.clearRect(0, 0, 150, 300);
		  canvasCtx.fillStyle = '#BadA55';
		  canvasCtx.fillRect(0, 300 - average, 150, 300);
		  canvasCtx.fillStyle = '#262626';
		  canvasCtx.font = "48px impact";
		  canvasCtx.fillText(Math.round(average - 40), -2, 300);
		*/
	    }
	};
    };
    
    draw(); 
}

