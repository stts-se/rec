function sendBlob(audioBlob, username, text, recording_id, onLoadEndFunc) {
    console.log("audio.js : CURRENT BLOB SIZE: "+ audioBlob.size);
    console.log("audio.js : CURRENT BLOB TYPE: "+ audioBlob.type);
    //clearResponse();
    
    // This is a bit backwards, since reader.readAsBinaryString below runs async.
    var reader = new FileReader();
    reader.addEventListener("loadend", function() {
	let rez = reader.result //contains the contents of blob as a typed array
	let payload = {
	    username : username,
	    audio : { file_type : currentBlob.type, data: btoa(rez)},
	    text : text,
	    recording_id : recording_id
	};
	
	sendJSON(payload, onLoadEndFunc);
	//sendButton.disabled = true;
    });
    
    reader.readAsBinaryString(audioBlob);
    
    console.log("SENDING BLOB"); 
};

function sendJSON(payload, onLoadEndFunc) {
    var xhr = new XMLHttpRequest();
    xhr.open("POST", baseURL + "/process/", true);
    xhr.setRequestHeader('Content-Type', 'application/json; charset=UTF-8');   
    xhr.onloadend = onLoadEndFunc
    xhr.send(JSON.stringify(payload));
}

