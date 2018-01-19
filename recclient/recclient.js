'use strict'
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
    sendButton = document.getElementById('send');
    sendButton.addEventListener('click', sendBlob);
    sendButton.disabled = true;
	
    navigator.mediaDevices.getUserMedia({'audio': true, video: false}).then(function(stream) {
	
		
	recorder = new MediaRecorder(stream);
	recorder.addEventListener('dataavailable', function (evt) {
	    updateAudio(evt.data); 
	});
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
	    username : "Mimmi Pigg",
	    audio : { file_type : currentBlob.type, data: btoa(rez)},
	    text : document.getElementById("text").value,//"fonclbt",
	    recording_id : "666"
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
