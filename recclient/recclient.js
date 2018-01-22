// See:
// https://developer.mozilla.org/en-US/docs/Web/API/MediaStream_Recording_API 
// https://mozdevs.github.io/MediaRecorder-examples/record-live-audio.html
// https://github.com/mdn/voice-change-o-matic

'use strict'

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


let recButton, stopButton, sendButton;
let baseURL = window.location.origin;
var currentBlob;
var recorder;

window.onload = function () {
    recButton = document.getElementById('rec');
    recButton.addEventListener('click', startRecording);
    recButton.disabled = false;
    
    stopButton = document.getElementById('stop');
    stopButton.addEventListener('click', stopRecording);
    stopButton.disabled = true;
    
    sendButton = document.getElementById('send');
    sendButton.addEventListener('click', sendBlob);
    sendButton.disabled = true;
	
    navigator.mediaDevices.getUserMedia({'audio': true, video: false}).then(function(stream) {
	source = audioCtx.createMediaStreamSource(stream);
        source.connect(analyser);
	visualize();	
	recorder = new MediaRecorder(stream);
	recorder.addEventListener('dataavailable', function (evt) {
	    updateAudio(evt.data); 
	});
	
	// ??
	recorder.onstop = function(evt) {}
    });
};    
function startRecording() {
    
    recButton.disabled = true;
    stopButton.disabled = false;
    sendButton.disabled = true;
    recorder.start();

    clearResponse();
    
}
function stopRecording() {
    
    recButton.disabled = false;
    stopButton.disabled = true;
    
    // make MediaRecorder stop recording
    // eventually this will trigger the dataavailable event
    recorder.stop();
    sendButton.disabled = false;
}
function sendBlob() {
    console.log("CURRENT BLOB SIZE: "+ currentBlob.size);
    console.log("CURRENT BLOB TYPE: "+ currentBlob.type);
    clearResponse();
    
    // This is a bit backwards, since reader.readAsBinaryString below runs async.
    var reader = new FileReader();
    reader.addEventListener("loadend", function() {
	let rez = reader.result //contains the contents of blob as a typed array
	let payload = {
	    username : document.getElementById("username").value,
	    audio : { file_type : currentBlob.type, data: btoa(rez)},
	    text : document.getElementById("text").value,//"fonclbt",
	    recording_id : document.getElementById("recording_id").value
	};
	
	sendJSON(payload);
	sendButton.disabled = true;
    });
    
    reader.readAsBinaryString(currentBlob);
    
    console.log("SENDING BLOB"); 
};

function sendJSON(payload) {
    var xhr = new XMLHttpRequest();
    xhr.open("POST", baseURL + "/process/", true);
    xhr.setRequestHeader('Content-Type', 'application/json; charset=UTF-8');
   
    // TODO error handling
    
    xhr.onloadend = function () {
     	// done
	console.log("STATUS: "+ xhr.statusText);
	console.log("STATUS: "+ JSON.stringify(xhr.response));
	showResponse(xhr.response);
    };

    
    
    xhr.send(JSON.stringify(payload));
}

function showResponse(json) {

    clearResponse();
    var resp = document.getElementById("response");

    var node = document.createTextNode(JSON.stringify(json));

    resp.appendChild(node);
};

function clearResponse() {
    document.getElementById("response").innerHTML = "";
}


function updateAudio(blob) {
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

//var visualSelect = document.getElementById("visual");

var drawVisual;

function visualize() {
    var WIDTH = canvas.width;
    var HEIGHT = canvas.height;
    
    
    analyser.fftSize = 256;
    var bufferLengthAlt = analyser.frequencyBinCount;
    console.log(bufferLengthAlt);
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

