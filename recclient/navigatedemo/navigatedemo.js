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

// visualization
// var drawVisual;
// var analyser = audioCtx.createAnalyser();
// analyser.minDecibels = -90;
// analyser.maxDecibels = -10;
// analyser.smoothingTimeConstant = 0.85;

// set up canvas context for visualizer
//var canvas = document.querySelector('.visualizer');
//var canvasCtx = canvas.getContext("2d");
// var intendedWidth = document.querySelector('.wrapper').clientWidth;
// canvas.setAttribute('width',intendedWidth / 2);


window.onload = function () {
    //var spk = document.getElementById('speak');

    init();
}


// function recording() {

//     // TODO set max recording time limit
    
//     recordButtonN.disabled = true;
//     //stopButton.disabled = false;
//     //sendButton.disabled = true;
//     recorderN.start();

//     clearResponse();
//     countDown();
// }



function move0(event) {
    // listen for enter key
    if ( event.keyCode == 13 || event.which == 13 ) {
	var move = document.getElementById('speak0').value;
	if (move === "right") {
	    moveRight();
	};
	if (move === "left") {
	    moveLeft();
	};
	if (move === "up") {
	    moveUp();
	};
	if (move === "down") {
	    moveDown();
	};
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




let baseURL = window.location.origin +"/rec" //navigatedemo";

// function hasAudio(blob) {
//     return false;
// }

// function replayAudio(blob) {
//     currentBlob = blob;    
//     var audio = document.getElementById('audio');
//     //console.log(blob);

//     // use the blob from the MediaRecorder as source for the audio tag
//     audio.src = URL.createObjectURL(blob);
//     audio.play();
// };

function init() {
    console.log("init: called");
    // document.body.onkeyup = function(e){
	// if (e.keyCode == 32)
           // start();
    // }

    
    navigator.mediaDevices.getUserMedia({'audio': true, video: false}).then(function(stream) {
    	sourceN = audioCtxN.createMediaStreamSource(stream);
        //source.connect(analyser);
	//visualize();
	console.log("init: creating MediaRecorder");
    	recorderN = new MediaRecorder(stream);
	recorderN.addEventListener('dataavailable', function (evt) {
	    updateAudio(evt.data);
	    sendAndReceiveBlob();
	});

	recorderN.onstop = function(evt) {}
	
	// // VAD | https://github.com/kdavis-mozilla/vad.js
	// var options = {
	//     source: source,
	//     voice_stop: function() {}, 
	//     voice_start: function() {
	// 	console.log('vad: voice_start');
	// 	if (state === notStartedState)
    	// 	    start();
    	// 	else if (state === pauseState)
    	// 	    unpause();
    	// 	else if (state === recState)
    	// 	    pause();
	//     }
	// };
	// var vad = new VAD(options);

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

function sendAndReceiveBlob() {
    console.log("sendAndReceiveBlob()");

    var onLoadEndFunc = function (data) {
	console.log("onLoadEndFunc data ", data);
	clearResponse(); // originally called before sending
	//sendButton.disabled = true; // originally called after sendJSON
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

    
    // AUDIO.sendBlob(currentBlob,
    // 	     document.getElementById("username").value,
    // 	     document.getElementById("text").innerHTML,
    // 	     document.getElementById("recording_id").innerHTML,
    // 	     onLoadEndFunc);

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
    clearResponse()
    recordButtonN.disabled = true;
    stopButtonN.disabled = false;
    if (recorderN != null) {
	
	console.log("startRecorder:",recorderN);
	
	//recorderN.start(); // continuous input
	recorderN.start(); // input only on send
    }

    //setTimeout(stop(), 1500);
    // record for a while
    //sleep(1000);
    //recorderN.stop();
    
    //recordButtonN.disabled = false;
}

function stop() {
    recorderN.stop();
    recordButtonN.disabled = false;
    stopButtonN.disabled = true;
    
}

// function stopRecorder() {
//     console.log("stopRecorder:",recorder);
//     if (recorder != null && recorder.state == "recording") {
// 	recorder.stop();
//     }
// }

// function setPauseEnabled(elem) {
//     //console.log("setPauseEnabled");
//     elem.innerHTML=playChar;
//     elem.setAttribute("onClick", "pause()");
//     // document.body.onkeyup = function(e){
//     // 	if(e.keyCode == 32){
//     //         pause();
//     // 	}
//     // }
// }

// function setUnpauseEnabled(elem) {
//     //console.log("setUnpauseEnabled");
//     elem.innerHTML=pauseChar;
//     elem.setAttribute("onClick", "unpause()");
//     // document.body.onkeyup = function(e){
// 	// if(e.keyCode == 32){
//            // unpause();
// 	// }
//     // }
// }


// function setPauseAndUnpauseDisabled(elem) {
//     //console.log("setPauseAndUnpauseDisabled");
//     elem.innerHTML="";
//     elem.setAttribute("onClick", "");
//     document.body.onkeyup = function(e){
// 	if(e.keyCode == 32){
//             // do nothing
// 	}
//     }
// }

// function start() {
//     console.log("start called");
//     stopRecorder();
//     pos = 0;
//     state = recState;
//     var elem = document.getElementById("animate");
//     elem.style.top = '0px'; 
//     elem.style.left = '0px';
//     setPauseEnabled(elem);
//     document.getElementById("start").blur();
//     var resetB = document.getElementById("reset");
//     resetB.removeAttribute("disabled");
//     //console.log(recorder.state);
//     if (id < 0) {
// 	id = setInterval(run, 20);
//     }
//     startRecorder();
// }

// function reset() {
//     console.log("reset called");
//     var elem = document.getElementById("animate");
//     elem.style.top = '0px'; 
//     elem.style.left = '0px'; 
//     stop();
// }

// function stop() {
//     console.log("stop called");
//     stopRecorder();
//     var resetB = document.getElementById("reset");
//     resetB.setAttribute("disabled","disabled");
//     var elem = document.getElementById("animate");
//     elem.innerHTML="";
//     setPauseAndUnpauseDisabled(elem);
//     state=notStartedState;
//     clearInterval(id);
//     pos=0;
//     id=-1;
//     resetB.blur();
// }

// function pause() {
//     console.log("pause called");
//     var elem = document.getElementById("animate");
//     setUnpauseEnabled(elem);
//     state = pauseState;
//     recorder.stop();
// }

// function unpause() {
//     console.log("unpause called");
//     var elem = document.getElementById("animate");
//     setPauseEnabled(elem);
//     state = recState;
//     recorder.start();
// }

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



// function run() {
//     var elem = document.getElementById("animate");
//     if (state == pauseState) {
// 	// do nothing
//     } else if (pos == 350) {
// 	stop();	
// 	elem.innerHTML="End";
//     } else {
// 	pos++; 
// 	elem.style.top = pos + 'px'; 
// 	elem.style.left = pos + 'px'; 
//     }
// }

// function visualize() {
//     var WIDTH = 400;//canvas.width;
//     var HEIGHT = 50;//canvas.height;

//     analyser.fftSize = 256;

//     var bufferLengthAlt = analyser.frequencyBinCount;
//     //console.log(bufferLengthAlt);
//     var dataArrayAlt = new Uint8Array(bufferLengthAlt);
    
//     canvasCtx.clearRect(0, 0, WIDTH, HEIGHT);
    
//     var draw = function() {
// 	// console.log("frequencyBinCount", analyser.frequencyBinCount);
// 	// console.log("draw called with state:", state);
// 	drawVisual = requestAnimationFrame(draw);
	
// 	analyser.getByteFrequencyData(dataArrayAlt);

// 	canvasCtx.fillStyle = 'rgb(0, 0, 0)';
// 	canvasCtx.fillRect(0, 0, WIDTH, HEIGHT);
	
// 	var barWidth = (WIDTH / bufferLengthAlt) * 2.5;
// 	var barHeight;
// 	var x = 0;

// 	// Only draw frequency bars when recording
// 	// When recording, the stop button is enabled 
// 	for(var i = 0; i < bufferLengthAlt; i++) {
// 	    barHeight = dataArrayAlt[i];
// 	    canvasCtx.fillStyle = 'rgb(' + (barHeight+100) + ',50,50)';
// 	    canvasCtx.fillRect(x,HEIGHT-barHeight/2,barWidth,barHeight/2);
	    
// 	    x += barWidth + 1;
// 	};

//     };
    
//     draw(); 
// }

