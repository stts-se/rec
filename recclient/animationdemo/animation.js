// https://developer.mozilla.org/en-US/docs/Web/API/MediaStream_Recording_API 
// https://mozdevs.github.io/MediaRecorder-examples/record-live-audio.html
// https://github.com/mdn/voice-change-o-matic

'use strict'

var pos = 0;
var id = -1;
//var recChar = "&#x23FA;";
var pauseChar = "speak"; //"&#x23f8;";
var playChar = "playing"; //"&#x25b6;";
var state = "";
var playState = "play";
var pauseState = "pause";

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

function hasAudio(blob) {
    return false;
}

function updateAudio(blob) {
    currentBlob = blob;    
    var audio = document.getElementById('audio');
    console.log(blob);
    // use the blob from the MediaRecorder as source for the audio tag
    audio.src = URL.createObjectURL(blob);
    audio.play();
};

//You can force a dataavailable event to occur, thereby delivering the latest sound to you so you can filter it, save it, or whatever.

function init() {
    console.log("init: called");
    document.body.onkeyup = function(e){
	if(e.keyCode == 32){
            start();
	}
    }
    navigator.mediaDevices.getUserMedia({'audio': true, video: false}).then(function(stream) {
    	source = audioCtx.createMediaStreamSource(stream);
        source.connect(analyser);
    	visualize();
	console.log("init: creating MediaRecorder");
    	recorder = new MediaRecorder(stream);
    	recorder.addEventListener('dataavailable', function (evt) {
	    console.log("recorder: data available");
	    // if (hasAudio(evt.data)) {
	    // 	if (state === pauseState) {
	    // 	    unpause();
	    // 	} else if (state === playState)	{
	    // 	    pause();
	    // 	}
	    // } 
    	    updateAudio(evt.data); 
    	});
	
    	recorder.onstop = function(evt) {}
    });
}

function startRecorder() {
    console.log("startRecorder:",recorder);
    if (recorder != null) {
	//recorder.start(100);
	recorder.start();
    }
}

function stopRecorder() {
    console.log("stopRecorder:",recorder);
    if (recorder != null && recorder.state == "recording") {
	recorder.stop();
    }
}

function setPauseEnabled(elem) {
    console.log("setPauseEnabled");
    elem.innerHTML=pauseChar;
    elem.setAttribute("onClick", "pause()");
    document.body.onkeyup = function(e){
	if(e.keyCode == 32){
            pause();
	}
    }
}

function setUnpauseEnabled(elem) {
    console.log("setUnpauseEnabled");
    elem.innerHTML=playChar;
    elem.setAttribute("onClick", "unpause()");
    document.body.onkeyup = function(e){
	if(e.keyCode == 32){
            unpause();
	}
    }
}


function setPauseAndUnpauseDisabled(elem) {
    console.log("setPauseAndUnpauseDisabled");
    elem.innerHTML="";
    elem.setAttribute("onClick", "");
    document.body.onkeyup = function(e){
	if(e.keyCode == 32){
            // do nothing
	}
    }
}

function start() {
    stopRecorder();
    pos = 0;
    state = playState;
    var elem = document.getElementById("animate");
    elem.style.top = '0px'; 
    elem.style.left = '0px';
    setPauseEnabled(elem);
    document.getElementById("start").blur();
    var resetB = document.getElementById("reset");
    resetB.removeAttribute("disabled");
    console.log(recorder.state);
    // var muteB = document.getElementById("mute");
    // muteB.removeAttribute("disabled");
    // unmute();
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
    // var muteB = document.getElementById("mute");
    // muteB.setAttribute("disabled", "disabled");
    var elem = document.getElementById("animate");
    elem.innerHTML="";
    setPauseAndUnpauseDisabled(elem);
    state="";
    clearInterval(id);
    pos=0;
    id=-1;
    resetB.blur();
    // unmute();
}

function pause() {
    var elem = document.getElementById("animate");
    setUnpauseEnabled(elem);
    state = pauseState;
    recorder.stop();
}

function unpause() {
    var elem = document.getElementById("animate");
    setPauseEnabled(elem);
    state = playState;
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

// function mute() {
//     var b = document.getElementById("mute");
//     b.innerHTML="Unmute";
//     b.setAttribute("onClick","unmute()");
//     console.log("mute: recorder.state=",recorder.state);
//     if (recorder.state === "recording") {
// 	//recorder.pause();
// 	gainNode.gain.setTargetAtTime(0, audioCtx.currentTime, 1);
//     }
// }

// function unmute() {
//     var b = document.getElementById("mute");
//     b.innerHTML="Mute";
//     b.setAttribute("onClick","mute()");
//     console.log("unmute: recorder.state=",recorder.state);
//     if (recorder.state === "paused") {
// 	//recorder.resume();
// 	gainNode.gain.setTargetAtTime(0, audioCtx.currentTime, 1);
//     }
// }

function visualize() {
    var WIDTH = 400;//canvas.width;
    var HEIGHT = 50;//canvas.height;

    console.log(WIDTH);
    console.log(HEIGHT);
    
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
	for(var i = 0; i < bufferLengthAlt; i++) {
	    barHeight = dataArrayAlt[i];  
	    canvasCtx.fillStyle = 'rgb(' + (barHeight+100) + ',50,50)';
	    canvasCtx.fillRect(x,HEIGHT-barHeight/2,barWidth,barHeight/2);
	    
	    x += barWidth + 1;
	};
    };
    
    draw(); 
}

