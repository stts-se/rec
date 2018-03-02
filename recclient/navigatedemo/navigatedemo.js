// https://developer.mozilla.org/en-US/docs/Web/API/MediaStream_Recordng_API 
// https://mozdevs.github.io/MediaRecorder-examples/record-live-audio.html
// https://github.com/mdn/voice-change-o-matic

'use strict'

var horizontal = 0;
var vertical = 0;
var pos = 0;
var id = -1;
var recChar = "<font color='#980000'>&#x23FA;</font>";
var pauseChar = "<font color='black'>&#x23f8;</font>";
var playChar =  "<font color='black'>&#x25b6;</font>";
var notStartedState = "not started";
var recState = "rec";
var pauseState = "pause";
var state = notStartedState;

var recordButtonN, stopButtonN;
// media recorder
var recorderN = null;
var currentBlobN;
var audioCtxN = new (window.AudioContext || window.webkitAudioContext)();
var sourceN;
var streamN;

// fork getUserMedia for multiple browser versions, for those that need prefixes
navigator.getUserMedia = (navigator.getUserMedia ||
                          navigator.webkitGetUserMedia ||
                          navigator.mozGetUserMedia ||
                          navigator.msGetUserMedia);



window.onload = function () {
    init();
}

function move0(event) {
    // listen for enter key
    if ( event.keyCode == 13 || event.which == 13 ) {
	var move = document.getElementById('speak0').value;
	move(move);
    }    
}

function move(strn) {
    if (strn.indexOf("right") !== -1) {
	moveRight();
    };
    if (strn.indexOf("left") !== -1) {
	moveLeft();
    };
    if (strn.indexOf("up") !== -1) {
	moveUp();
    };
    if (strn.indexOf("down") !== -1) {
	moveDown();
    };
}    




let baseURL = window.location.origin +"/rec";


function init() {
    console.log("init: called");
    
    navigator.mediaDevices.getUserMedia({'audio': true, video: false}).then(function(stream) {
    	sourceN = audioCtxN.createMediaStreamSource(stream);
	console.log("init: creating MediaRecorder");
    	recorderN = new MediaRecorder(stream);
	recorderN.addEventListener('dataavailable', function (evt) {
	    updateAudio(evt.data);
	    sendAndReceiveBlob();
	});

	recorderN.onstop = function(evt) {}
	
    });

    recordButtonN = document.getElementById('record');
    stopButtonN = document.getElementById('stop');
    stopButtonN.disabled = true;
}

function updateAudio(blob) {
    console.log("updateAudio()", blob.size);
    
    console.log("updateAudio(): "+ blob.type);

    currentBlobN = blob;
    
    var audio = document.getElementById('audio');
    // use the blob from the MediaRecorder as source for the audio tag
    audio.src = URL.createObjectURL(blob);
    audio.play();
};

function sendAndReceiveBlob() {
    console.log("sendAndReceiveBlob()");

    var onLoadEndFunc = function (data) {
	console.log("onLoadEndFunc data ", data);
	clearResponse();
	console.log("onLoadEndFunc|STATUS : "+ data.target.status + "/" + data.target.statusText);
	console.log("onLoadEndFunc|RESPONSE : "+ data.target.responseText);
	if (data.target.status === 200) {
	    var o = JSON.parse(data.target.responseText);
	    move(o.recognition_result);
	    showResponse(data.target.responseText);
	} else {
	    showError(data);
	}
    };

    
    AUDIO.sendBlob(currentBlobN,
		   "tmpuser",
		   "_",
		   "nav_rec",
		   onLoadEndFunc);

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

function showError(data) {
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
	"recording_id": "",
	"message": data.target.status + "/" + data.target.statusText + ": " + data.target.responseText.trim(),
    };
    
    var j = JSON.stringify(json, null, '\t');
    
    resp.innerHTML = j;
}


function sleep(milliseconds) {
  var start = new Date().getTime();
  for (var i = 0; i < 1e7; i++) {
    if ((new Date().getTime() - start) > milliseconds){
      break;
    }
  }
}



function record() {
    clearResponse();
    recordButtonN.disabled = true;
    stopButtonN.disabled = false;
    if (recorderN != null) {
	
	console.log("startRecorder:",recorderN);
	
	//recorderN.start(); // continuous input
	recorderN.start(); // input only on send
	var stopRecording = setInterval(function() {
	    
	    console.log("STOP RECORDING CALLED");
	    
	    stop();
	    clearInterval(stopRecording);
	}, 1500);
	
    }

}

function stop() {
    recorderN.stop();
    recordButtonN.disabled = false;
    stopButtonN.disabled = true;
    
}


function moveRight() {
    var elem = document.getElementById("animate");
    
    for(var i = 0; i < 100; i++) {
	horizontal++;
	elem.style.left = horizontal + 'px'; 
    }; 
}

function moveLeft() {
    var elem = document.getElementById("animate");
    
    for(var i = 0; i < 100; i++) {
	horizontal--;
	elem.style.left = horizontal + 'px'; 
    }; 
}

function moveUp() {
    var elem = document.getElementById("animate");
    
    for(var i = 0; i < 100; i++) {
	vertical--;
	elem.style.top = vertical + 'px'; 
    };
}

function moveDown() {
    var elem = document.getElementById("animate");
    
    for(var i = 0; i < 100; i++) {
	vertical++;
	elem.style.top = vertical + 'px'; 
    };
}

