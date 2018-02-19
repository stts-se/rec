var pos = 0;
var id = -1;
var pauseChar = "&#x23f8;";
var playChar = "&#x25b6;";
var pausing = false;

function start() {
    pos = 0;
    pausing = false;
    var elem = document.getElementById("animate");
    elem.innerHTML=pauseChar;
    elem.style.top = '0px'; 
    elem.style.left = '0px'; 
    elem.setAttribute("onClick", "pause()");
    var stopB = document.getElementById("stop");
    stopB.removeAttribute("disabled");
    if (id < 0) {
	id = setInterval(run, 20);
    }
}

function reset() {
    var elem = document.getElementById("animate");
    elem.style.top = '0px'; 
    elem.style.left = '0px'; 
    stop();
}

function stop() {
    var stopB = document.getElementById("stop");
    stopB.setAttribute("disabled","disabled");
    var elem = document.getElementById("animate");
    elem.innerHTML="";
    elem.setAttribute("onClick", "");
    pausing=false;
    clearInterval(id);
    pos=0;
    id=-1;
}

function pause() {
    var elem = document.getElementById("animate");
    elem.innerHTML=playChar;
    elem.setAttribute("onClick", "unpause()");
    pausing = true;
}

function unpause() {
    var elem = document.getElementById("animate");
    elem.innerHTML=pauseChar;
    elem.setAttribute("onClick", "pause()");
    pausing = false;
}

function run() {
    var elem = document.getElementById("animate");
    if (pausing == true) {
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
