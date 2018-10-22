"use strict";

//TODO: Clean up: refactor common functions and put into lib. Beware
//of HTML interaction sprinkled everywhere (see getElementById calls,
//for instance).

// See:
// https://developer.mozilla.org/en-US/docs/Web/API/MediaStream_Recording_API 
// https://mozdevs.github.io/MediaRecorder-examples/record-live-audio.html
// https://github.com/mdn/voice-change-o-matic


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
var recognition;

let defaultScriptName = "dictator";
let user = "anon";


function autosize(area){
    area.style.cssText = 'width: 100%; border: none; height:' + area.scrollHeight + 'px';
}


window.onload = function () {
    
    if (!('webkitSpeechRecognition' in window)) {
	alert("This browser does not support webkit speech recognition. Try Google Chrome.");
	return;
    };
    
    recognition = new webkitSpeechRecognition();
    
    let langSelect = document.getElementById("lang");
    langSelect.addEventListener("change", function(event) {
	
	var i  = langSelect.selectedIndex
	var lang = langSelect.options[i].value
	//if (startButton.disabled) {
	//    stopButton.click();
	//};
	recognition.lang = lang;
	
    });
    
    let tempResponse = document.getElementById("tempresponse");
    let finalResponse = document.getElementById("finalresponse");
    
    
    recognition.lang = langSelect.value;
    recognition.continuous = true;
    recognition.interimResults = true;
    
    recognition.onresult = function(event) {	
	for (var i = event.resultIndex; i < event.results.length; ++i) {
	    if (event.results[i].isFinal) {
		let full = finalResponse.value + '\n' + event.results[i][0].transcript.trim(); // + '<br>';
		finalResponse.value = full.trim();
		autosize(finalResponse);
		tempResponse.innerHTML = '';
		
	    } else {
		tempResponse.innerHTML = event.results[i][0].transcript;
	    }
	}
    };    
    
    recognition.onend = function() { // No 'event' arg?
	recButton.disabled = false;
	stopAndSendButton.disabled = true;
	document.getElementById("micimage").src = "mic.gif";
	//console.log("'onend' called!");
    };
    
    recognition.onerror = function(event) {
	console.log("Error: ", event);
	
	if (event.error == 'no-speech') {
	    document.getElementById("micimage").src = "mic.gif";
	    // TODO msg user
	    document.getElementById("msg").innerHTML = 'No speech<br>';
	    
	};
	if (event.error == 'audio-capture') {
	    document.getElementById("micimage").src = "mic.gif";
	    document.getElementById("msg").innerHTML = 'No microphone<br>';
	    
	};
	if (event.error == 'not-allowed') {
	    if (event.timeStamp - start_timestamp < 100) {
		document.getElementById("msg").innerHTML = 'Blocked<br>';
	    } else {
		document.getElementById("msg").innerHTML = 'Denied<br>';
	    }
	    
	};
	    if (event.error == 'network') {
		document.getElementById("msg").innerHTML = 'Network error<br>';
	    }
	};


    
    var url = new URL(document.URL);
    
    recButton = document.getElementById('rec');
    recButton.addEventListener('click', startRecording);
    recButton.disabled = false;
    
    stopAndSendButton = document.getElementById('stopandsend');
    stopAndSendButton.addEventListener('click', stopAndSend);
    stopAndSendButton.disabled = true;
    
    //console.log("navigator.mediaDevices:", navigator.mediaDevices);
    let mediaAccess = navigator.mediaDevices.getUserMedia({'audio': true, video: false});
    //console.log("navigator.mediaDevices.getUserMedia:", mediaAccess);
    
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


    // Set up abbreviations table, etc
        // Init abbrev hash table from server
    loadAbbrevTable();
    
    
    // Bootstrap already has JQuery as a dependancy

    
    $("#abbrev_table").on('click', 'tr', function(evt) {
	let row = $(this);
	//let row = row0[0];
	let dts = row.children('td');
	//console.log("KLIKKETIKLIKK ++", dts);
	//console.log("KLIKKETIKLIKK --", dts[0]);
	//console.log("KLIKKETIKLIKK --", dts[1]);
	//console.log("---------------------");
    } );
    
    
    $("#add_abbrev_button").on('click', function(evt) {
	let abbrev = document.getElementById("input_abbrev").value.trim();
	let expansion = document.getElementById("input_expansion").value.trim();
	
	// TODO add button should be disablem without text in both input fields, etc
	// TODO proper validation
	if (abbrev === "") {
	    document.getElementById("msg").innerText = "Cannot add empty abbreviation";
	    return;
	};
	if (expansion === "") {
	    document.getElementById("msg").innerText = "Cannot add empty expansion";
	    return;
	};
	
	abbrevMap[abbrev] = expansion;
	
	// TODO Nested async calls: NOT NICE, change to promises instead
	//addAbbrev contains a(n async) call to loadAbbrevTable();
	
	addAbbrev(abbrev, expansion);	

	
	//console.log("abbrev", abbrev);
	//console.log("expansion", expansion);
    });
    
    $("#delete_abbrev_button").on('click', function(evt) {
	let abbrev = document.getElementById("input_abbrev").value.trim();
	
	
	// TODO add button should be disablem without text in both input fields, etc
	// TODO proper validation
	if (abbrev === "") {
	    document.getElementById("msg").innerText = "Cannot delete empty abbreviation";
	    return;
	};
	
	delete abbrevMap[abbrev];
	
	// TODO Nested async calls: NOT NICE, change to promises instead
	//addAbbrev contains a(n async) call to loadAbbrevTable();
	
	deleteAbbrev(abbrev);	

	
	//console.log("abbrev", abbrev);
	//console.log("expansion", expansion);
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

    // TODO Don't clear between calls, but save text i text area
    document.getElementById("tempresponse").innerHTML = '';
    document.getElementById("finalresponse").value = '';
    
    recognition.start();
    
    document.getElementById("micimage").src = "mic-animate.gif";
    
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
    
    recognition.stop();
    
    // make MediaRecorder stop recording
    // eventually this will trigger the dataavailable event
    recorder.stop();
    
    document.getElementById("micimage").src = "mic.gif";
    
    stopAndSendButton.disabled = true;
    // stopButton.disabled = false;
    clearInterval(setIntFunc);
    document.getElementById("rec_progress").value = "0";
    document.getElementById("rec_progress").setAttribute("aria-valuenow", "0");
}

var setIntFunc;

function countDown() {
    var max = 15;
    let tick = 10;
    var dur = 0;

    document.getElementById("rec_progress").value = ""+ dur;
    document.getElementById("rec_progress").setAttribute("aria-valuenow", ""+ dur);
    
    setIntFunc = setInterval(function() {

	dur = dur + (tick / 1000);
	
	document.getElementById("rec_progress").value = ""+ dur;
	document.getElementById("rec_progress").setAttribute("aria-valuenow", ""+ dur);
	
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
		   defaultScriptName, //Woohoo, hardwired! ("dictator", see above)
		   user,              //Woohoo, hardvirew! ("anon", see above)
		   document.getElementById("finalresponse").value, // text TODO Send as argument to function rather that getting it from HTML...
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


// Stuff to add possibility of entering abbreviations that are
// automatically expanded in manually edited text

// Asks sever for list of persited abbrevisations, and fills in the
// clients hashmap
function loadAbbrevTable() {
    let xhr = new XMLHttpRequest();
    
    xhr.onload = function() {
	if ( xhr.readyState === 4 && 
     	     xhr.status === 200) {
	    
	    // TODO Catch errors here
	    let serverAbbrevs = JSON.parse(xhr.responseText);
	    //console.log("#######", serverAbbrevs);
	    abbrevMap = {};
	    for (var i = 0; i < serverAbbrevs.length; i++) {
		//console.log("i: ", i, serverAbbrevs[i]);
		let a = serverAbbrevs[i];
		abbrevMap[a.abbrev] = a.expansion;
	    };
	    updateAbbrevTable();
	    
	};
    };
    
    xhr.open("GET", baseURL+ "/list_abbrevs" , true)
    xhr.send();
};

function updateAbbrevTable() {
    let at = document.getElementById("abbrev_table_body");
    at.innerHTML = '';
    Object.keys(abbrevMap).forEach(function(k) {
	let v = abbrevMap[k];
	let tr = document.createElement('tr');
	let td1 = document.createElement('td');
	let td2 = document.createElement('td');

	td1.innerText = k;
	td2.innerText = v;
	
	tr.appendChild(td1);
	tr.appendChild(td2);
	at.appendChild(tr);
    });
    
    
}

function addAbbrev(abbrev, expansion) {
    let xhr = new XMLHttpRequest();
    
    //TODO Notify user of response
    // TODO error handling
    
    xhr.onload = function(resp) {
	//console.log("RESP", resp);

	// TODO Show response in client
	
	// TODO Nested async calls: NOT NICE, change to promises instead
	loadAbbrevTable();
    };
    
    xhr.open("GET", baseURL+ "/add_abbrev/"+ abbrev + "/"+ expansion , true)
    xhr.send();
};
function deleteAbbrev(abbrev) {
    let xhr = new XMLHttpRequest();
    
    //TODO Notify user of response
    // TODO error handling
    
    xhr.onload = function(resp) {
	//console.log("RESP", resp);

	// TODO Show response in client
	
	// TODO Nested async calls: NOT NICE, change to promises instead
	loadAbbrevTable();
    };
    
    xhr.open("GET", baseURL+ "/delete_abbrev/"+ abbrev, true)
    xhr.send();
};

