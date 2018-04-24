var AUDIO = {};

AUDIO.sendBlob = function(audioBlob, username, text, recording_id, onLoadEndFunc) {
    console.log("audio.js : BLOB SIZE: "+ audioBlob.size);
    console.log("audio.js : BLOB TYPE: "+ audioBlob.type);
    
    // This is a bit backwards, since reader.readAsBinaryString below runs async.
    var reader = new FileReader();
    reader.addEventListener("loadend", function() {
	let rez = reader.result //contains the contents of blob as a typed array
	let payload = {
	    username : username,
	    audio : { file_type : audioBlob.type, data: btoa(rez)},
	    text : text,
	    recording_id : recording_id
	};
	
	AUDIO.sendJSON(payload, onLoadEndFunc);
    });
    
    reader.readAsBinaryString(audioBlob);
    
    console.log("audio.js : SENDING BLOB"); 
};

AUDIO.sendJSON = function(payload, onLoadEndFunc) {
    var xhr = new XMLHttpRequest();
    xhr.open("POST", baseURL + "/process/?verb=true", true);
    xhr.setRequestHeader('Content-Type', 'application/json; charset=UTF-8');   
    xhr.onloadend = onLoadEndFunc;    
    xhr.send(JSON.stringify(payload));
};


