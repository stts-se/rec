// https://developer.mozilla.org/en-US/docs/Web/API/MediaStream_Recording_API 
// https://mozdevs.github.io/MediaRecorder-examples/record-live-audio.html
// https://github.com/mdn/voice-change-o-matic

'use strict'

var pos = 0;
var id = -1;
var recChar = "<font color='#980000'>&#x23FA;</font>";
var pauseChar = "<font color='black'>&#x23f8;</font>";
var playChar =  "<font color='black'>&#x25b6;</font>";
var notStartedState = "not started";
var recState = "rec";
var pauseState = "pause";
var state = notStartedState;

// media recorder
var recorder = null;
var currentBlob;
var audioCtx = new (window.AudioContext || window.webkitAudioContext)();
var source;
var stream;

// fork getUserMedia for multiple browser versions, for those that need prefixes
navigator.getUserMedia = (navigator.getUserMedia ||
                          navigator.webkitGetUserMedia ||
                          navigator.mozGetUserMedia ||
                          navigator.msGetUserMedia);

// visualization
var drawVisual;
var analyser = audioCtx.createAnalyser();
analyser.minDecibels = -90;
analyser.maxDecibels = -10;
analyser.smoothingTimeConstant = 0.85;

// set up canvas context for visualizer
var canvas = document.querySelector('.visualizer');
var canvasCtx = canvas.getContext("2d");
// var intendedWidth = document.querySelector('.wrapper').clientWidth;
// canvas.setAttribute('width',intendedWidth / 2);


window.onload = function () {
    init();
}

let baseURL = window.location.origin +"/rec/animationdemo";

function hasAudio(blob) {
    return false;
}

function replayAudio(blob) {
    currentBlob = blob;    
    var audio = document.getElementById('audio');
    //console.log(blob);

    // use the blob from the MediaRecorder as source for the audio tag
    audio.src = URL.createObjectURL(blob);
    audio.play();
};

function init() {
    console.log("init: called");
    // document.body.onkeyup = function(e){
	// if (e.keyCode == 32)
           // start();
    // }

    navigator.mediaDevices.getUserMedia({'audio': true, video: false}).then(function(stream) {
    	source = audioCtx.createMediaStreamSource(stream);
        source.connect(analyser);
	visualize();
	console.log("init: creating MediaRecorder");
    	recorder = new MediaRecorder(stream);


	
	// VAD | https://github.com/kdavis-mozilla/vad.js
	var options = {
	    source: source,
	    voice_stop: function() {}, 
	    voice_start: function() {
		console.log('vad: voice_start');
		if (state === notStartedState)
    		    start();
    		else if (state === pauseState)
    		    unpause();
    		else if (state === recState)
    		    pause();
	    }
	};
	var vad = new VAD(options);

	// recorder.addEventListener('dataavailable', function (evt) {
	//     console.log("recorder: data available");
	//     if (hasAudio(evt.data)) {
	//     	if (state === pauseState) {
	//     	    unpause();
	//     	} else if (state === recState)	{
	//     	    pause();
	//     	}
	//     } 
    	//     //replayAudio(evt.data); 
    	// });
	
    	recorder.onstop = function(evt) {}
    });
}

function startRecorder() {
    console.log("startRecorder:",recorder);
    if (recorder != null) {
	recorder.start(100); // continuous input
	//recorder.start(); // input only on send
    }
}

function stopRecorder() {
    console.log("stopRecorder:",recorder);
    if (recorder != null && recorder.state == "recording") {
	recorder.stop();
    }
}

function setPauseEnabled(elem) {
    //console.log("setPauseEnabled");
    elem.innerHTML=playChar;
    elem.setAttribute("onClick", "pause()");
    // document.body.onkeyup = function(e){
    // 	if(e.keyCode == 32){
    //         pause();
    // 	}
    // }
}

function setUnpauseEnabled(elem) {
    //console.log("setUnpauseEnabled");
    elem.innerHTML=pauseChar;
    elem.setAttribute("onClick", "unpause()");
    // document.body.onkeyup = function(e){
	// if(e.keyCode == 32){
           // unpause();
	// }
    // }
}


function setPauseAndUnpauseDisabled(elem) {
    //console.log("setPauseAndUnpauseDisabled");
    elem.innerHTML="";
    elem.setAttribute("onClick", "");
    document.body.onkeyup = function(e){
	if(e.keyCode == 32){
            // do nothing
	}
    }
}

function start() {
    console.log("start called");
    stopRecorder();
    pos = 0;
    state = recState;
    var elem = document.getElementById("animate");
    elem.style.top = '0px'; 
    elem.style.left = '0px';
    setPauseEnabled(elem);
    document.getElementById("start").blur();
    var resetB = document.getElementById("reset");
    resetB.removeAttribute("disabled");
    //console.log(recorder.state);
    if (id < 0) {
	id = setInterval(run, 20);
    }
    startRecorder();
}

function reset() {
    console.log("reset called");
    var elem = document.getElementById("animate");
    elem.style.top = '0px'; 
    elem.style.left = '0px'; 
    stop();
}

function stop() {
    console.log("stop called");
    stopRecorder();
    var resetB = document.getElementById("reset");
    resetB.setAttribute("disabled","disabled");
    var elem = document.getElementById("animate");
    elem.innerHTML="";
    setPauseAndUnpauseDisabled(elem);
    state=notStartedState;
    clearInterval(id);
    pos=0;
    id=-1;
    resetB.blur();
}

function pause() {
    console.log("pause called");
    var elem = document.getElementById("animate");
    setUnpauseEnabled(elem);
    state = pauseState;
    recorder.stop();
}

function unpause() {
    console.log("unpause called");
    var elem = document.getElementById("animate");
    setPauseEnabled(elem);
    state = recState;
    recorder.start();
}

function run() {
    var elem = document.getElementById("animate");
    if (state == pauseState) {
	// do nothing
    } else if (pos == 350) {
	stop();	
	elem.innerHTML="End";
    } else {
	pos++; 
	elem.style.top = pos + 'px'; 
	elem.style.left = pos + 'px'; 
    }
}

function visualize() {
    var WIDTH = 400;//canvas.width;
    var HEIGHT = 50;//canvas.height;

    analyser.fftSize = 256;

    var bufferLengthAlt = analyser.frequencyBinCount;
    //console.log(bufferLengthAlt);
    var dataArrayAlt = new Uint8Array(bufferLengthAlt);
    
    canvasCtx.clearRect(0, 0, WIDTH, HEIGHT);
    
    var draw = function() {
	// console.log("frequencyBinCount", analyser.frequencyBinCount);
	// console.log("draw called with state:", state);
	drawVisual = requestAnimationFrame(draw);
	
	analyser.getByteFrequencyData(dataArrayAlt);

	canvasCtx.fillStyle = 'rgb(0, 0, 0)';
	canvasCtx.fillRect(0, 0, WIDTH, HEIGHT);
	
	var barWidth = (WIDTH / bufferLengthAlt) * 2.5;
	var barHeight;
	var x = 0;

	// Only draw frequency bars when recording
	// When recording, the stop button is enabled 
	for(var i = 0; i < bufferLengthAlt; i++) {
	    barHeight = dataArrayAlt[i];
	    canvasCtx.fillStyle = 'rgb(' + (barHeight+100) + ',50,50)';
	    canvasCtx.fillRect(x,HEIGHT-barHeight/2,barWidth,barHeight/2);
	    
	    x += barWidth + 1;
	};

    };
    
    draw(); 
}

