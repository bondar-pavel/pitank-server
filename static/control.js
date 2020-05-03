document.onkeydown = function (e) { key(e.keyCode, true); }
document.onkeyup = function (e) { key(e.keyCode, false); }

//document.getElementById('video').src = 'http://' + location.hostname + ':8280/?action=stream';

// WebSocket connection to the server.
var socket = null;
// Currently pushed buttons.
var actions = [];
// Flag not to flood server with 'stop'.
var stopped = true;
// Flag to indicate program mode (no return).
var program = false;
// LocalDescription for WebRTC connection
var localDescription = "";
var localDescriptionSent = false;
// variables for tracking communication channel states
var WebsocketOpened = false;
var WebRTCOpened = false;

const urlParams = new URLSearchParams(window.location.search);
const tankName = urlParams.get('name') || 'pitank';
//const offer = urlParams.get('offer');
//console.log("Tank offer: ", offer);

var action_to_track = {
    'engine_forward': ['trackright_forward', 'trackleft_forward'],
    'engine_right': ['trackleft_forward'],
    'engine_left': ['trackright_forward'],
    'engine_reverse': ['trackleft_reverse', 'trackright_reverse'],
    'camera_start': ['camera_start'],
}

// Init WebRTC
let pc = new RTCPeerConnection({
    iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
})

pc.onsignalingstatechange = e => {
    warn("Signaling state change: " + pc.signalingState);
}
pc.oniceconnectionstatechange = e => {
    switch(pc.iceConnectionState) {
        case "connected":
            setWebRTCStatus("connected");
            break;
        case "disconnected":
            setWebRTCStatus(false);
            break;
    }

    warn("ICE state change:" + pc.iceConnectionState);
}
pc.onicecandidate = event => {
    warn("Generating localDescription");
    setWebRTCStatus("connecting");

    if (event.candidate === null) {
        warn("Generating localDescription done");
        // send Offer to tank
        fetch(`${window.location.origin}/api/tanks/${tankName}/offer`, {
            method: 'POST',
            body: btoa(JSON.stringify(pc.localDescription))
        }).then(response => {
            return response.text();
        }).then(data => {
            console.log("Returned answer: ", data);
        }).catch(warn);
    }
}

pc.onnegotiationneeded = e => {
    console.log("Starting offer generation");
    pc.createOffer().then(d => pc.setLocalDescription(d)).catch(warn);
}

let dc = pc.createDataChannel('commands');

let send_actions_webrts;

warn('New DataChannel ' + dc.label)
dc.onclose = () => {
    console.log('dc has closed');
}
dc.onopen = () => {
    console.log('dc has opened');

    send_actions_webrts = () => {
        let stop_now = actions.length == 0;
        if (!(stop_now && stopped)) {
            console.log("Sending commands via WebRTC channel");
            dc.send(JSON.stringify(get_commands(actions)));
        }
        stopped = stop_now;
    }
}
dc.onmessage = e => {
    try {
        var data = JSON.parse(e.data);
        if (data.time) {
            updateRoundTripTime(data.time);
        }
    } catch (err) {
        debug('WebRTC: message ' + e.data);
    }
}
//dc.send("Connection established!");

// Converts list of actions to command for the server
function get_commands(actions) {
    if (actions.length == 0) {
        return { commands: "stop", time: Date.now() };
    }

    // use dict for commands to make sure each command appears only once
    let commands = {};
    for (var i = 0; i < actions.length; ++i) {
        let cmd = action_to_track[actions[i]];
        console.log("Processing", actions[i], cmd);
        if (cmd) {
            console.log("right", cmd);
            for (var j = 0; j < cmd.length; ++j) {
                commands[cmd[j]] = true;
            }
        } else {
            console.log("wrong", actions[i]);
            commands[actions[i]] = true;
        }
    }
    console.log("Commands:", commands);
    return {
        commands: Object.keys(commands).join(','),
        time: Date.now()
    };
}
// Send 'actions' to 'socket', with reconnect availability.
function send_actions() {
    // Send commands via webrtc channel if available
    if (WebRTCOpened) {
        send_actions_webrts();
    } else {
        send_actions_websocket()
    }
}

function send_actions_websocket() {
    if (socket) {
        var stop_now = actions.length == 0;
        if (!(stop_now && stopped)) {
            socket.send(JSON.stringify(get_commands(actions)));
        }
        stopped = stop_now;
    } else {
        warn('connecting...');
        setWebsocketStatus("connecting");

        localDescriptionSent = false;
        socket = new WebSocket('ws://' + location.host + '/api/tanks/' + tankName + '/connect');

        socket.onopen = function (e) {
            warn(null);
            console.log('WebSocket: opened');
            setWebsocketStatus("connected");
            send_actions();
        }
        socket.onclose = function (e) {
            socket = null;
            warn('WebSocket: closed (' + e.code + ')');
            setWebsocketStatus(false);
        }
        socket.onmessage = function (e) {
            try {
                var data = JSON.parse(e.data);
                if (data.time) {
                    updateRoundTripTime(data.time);
                }
                if (data.answer) {
                    // Finish webrtc connection init with remote answer
                    console.log("Received remote answer:" + data.answer);
                    pc.setRemoteDescription(new RTCSessionDescription(JSON.parse(atob(data.answer)))).catch(warn);
                }
            } catch (err) {
                debug('WebSocket: message ' + e.data);
            }
        }
        socket.onerror = function (e) {
            warn('WebSocket: error');
        }
    }
}

function view(tank) {
    var tank_elements = document.getElementsByName('engine-tank');
    var classic_elements = document.getElementsByName('engine-classic');

    if (tank === 'program') {
        program = true;
        document.getElementById('controls').style.display = 'none';
        document.getElementById('program').style.display = '';
    } else {
        for (var i = 0; i < tank_elements.length; ++i) {
            tank_elements[i].style.display = tank ? '' : 'none';
        }
        for (var i = 0; i < classic_elements.length; ++i) {
            classic_elements[i].style.display = tank ? 'none' : '';
        }
    }
}

view(false);

setInterval(function () {
    send_actions(actions);
}, 500);

function setWebRTCStatus(status) {
    let webrtcDot = document.getElementById('webrtc');

    if (status == "connected") {
        webrtcDot.className = "dot connected";
        WebRTCOpened = true;
    } else if (status == "connecting") {
        webrtcDot.className = "dot connecting";
        WebRTCOpened = false;
    } else {
        webrtcDot.className = "dot disconnected";
        WebRTCOpened = false;
    }

    updateCommunicationChannelStatus();
}

function setWebsocketStatus(status) {
    let wsDot = document.getElementById('websocket');

    if (status == "connected") {
        wsDot.className = "dot connected";
        WebsocketOpened = true;
    } else if (status == "connecting") {
        wsDot.className = "dot connecting";
        WebsocketOpened = false;
    } else {
        wsDot.className = "dot disconnected";
        WebsocketOpened = false;
    }

    updateCommunicationChannelStatus();
}

// updateCommunicationChannelStatus updates status of communication channel currently in use
function updateCommunicationChannelStatus() {
    let commandChannelDot = document.getElementById('command_channel');
    let commandChannelName = document.getElementById('command_channel_name');

    if (!WebRTCOpened && !WebsocketOpened) {
        commandChannelDot.className = "dot disconnected";
        commandChannelName.innerText = "None";
    } else {
        commandChannelDot.className = "dot connected";
        commandChannelName.innerText = WebRTCOpened ? "WebRTC" : "Websocket";
    }
}

function updateRoundTripTime(time) {
    latency = Date.now() - time;

    let roundTripTime = document.getElementById('round_trip_time');
    roundTripTime.innerText = latency + " ms";

    // update rtt indicator color,
    // for now just reuse existent classes for colors green/yellow/red
    let roundTripDot = document.getElementById('round_trip');
    if (latency < 50) {
        roundTripDot.className = "dot connected";
    } else if (latency < 500) {
        roundTripDot.className = "dot connecting";
    } else {
        roundTripDot.className = "dot disconnected";
    }
}

function warn(text) {
    console.log(text)
    document.getElementById('warning').innerText = text;
}

function info(text) {
    document.getElementById('info').innerText = text;
}

function debug(text) {
    document.getElementById('debug').innerText = text;
}

function go(action, enabled) {
    var button = document.getElementById(action);
    button.className = button.className.replace(enabled ? ' inactive' : ' active',
        enabled ? ' active' : ' inactive');

    if (enabled && action.indexOf('track') >= 0) {
        view(true);
    } else if (enabled && action.indexOf('engine') >= 0) {
        view(false);
    }

    for (var i = 0; i < actions.length; ++i) {
        if (actions[i] == action) {
            if (enabled) {
                var found = true;
            } else {
                actions.splice(i, 1);
                send_actions();
            }
            break;
        }
    }
    if (enabled && !found) {
        actions.push(action);
        send_actions();
    }

    return false;
}

var key_action_mapping = {
    37: 'engine_left',
    38: 'engine_forward',
    39: 'engine_right',
    40: 'engine_reverse',
    87: 'engine_forward',
    83: 'engine_reverse',
    65: 'engine_left',
    68: 'engine_right',
    81: 'trackleft_forward',
    69: 'trackright_forward',
    90: 'trackleft_reverse',
    67: 'trackright_reverse',
    219: 'tower_left',
    221: 'tower_right'
};

function key(code, enabled) {
    if (program) {
        return true;
    }
    var action = key_action_mapping[code];
    if (action) {
        go(action, enabled);
        return false;
    }
    return true;
}

function button(btn, enabled) {
    go(btn.id, enabled);
}