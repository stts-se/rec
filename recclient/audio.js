var AUDIO = {};

AUDIO.sendBlob = function(audioBlob, scriptname, username, text, recording_id, onLoadEndFunc) {
    //console.log("audio.js : BLOB SIZE: "+ audioBlob.size);
    //console.log("audio.js : BLOB TYPE: "+ audioBlob.type);
    
    // This is a bit backwards, since reader.readAsBinaryString below runs async.
    var reader = new FileReader();
    reader.addEventListener("loadend", function() {
	let rez = reader.result //contains the contents of blob as a typed array
	let payload = {
	    scriptname : scriptname,
	    username : username,
	    audio : { file_type : audioBlob.type, data: btoa(rez)},
	    text : text,
	    recording_id : recording_id
	};
	
	AUDIO.sendJSON(payload, onLoadEndFunc);
    });
    
    reader.readAsBinaryString(audioBlob);
    
    //console.log("audio.js : SENDING BLOB"); 
};

AUDIO.sendJSON = function(payload, onLoadEndFunc) {
    //console.log("PAYLOAD:", payload);

    var xhr = new XMLHttpRequest();
    xhr.open("POST", baseURL + "/process/?verb=true", true);
    xhr.setRequestHeader('Content-Type', 'application/json; charset=UTF-8');   
    xhr.onloadend = onLoadEndFunc;    
    xhr.send(JSON.stringify(payload));
};



// Snippet lifted from https://stackoverflow.com/questions/105034/create-guid-uuid-in-javascript#2117523:
function uuidv4() {
  return ([1e7]+-1e3+-4e3+-8e3+-1e11).replace(/[018]/g, c =>
    (c ^ crypto.getRandomValues(new Uint8Array(1))[0] & 15 >> c / 4).toString(16)
  )
}

