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

const urlParams = new URLSearchParams(window.location.search);
const tankName = urlParams.get('name') || 'pitank';
//const offer = urlParams.get('offer');
//console.log("Tank offer: ", offer);

var action_to_track = {
    'engine_forward': ['trackright_forward', 'trackleft_forward'],
    'engine_right': ['trackleft_forward'],
    'engine_left': ['trackright_forward'],
    'engine_reverse': ['trackleft_reverse', 'trackright_reverse'],
}

// Init WebRTC
let pc = new RTCPeerConnection({
    iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
})

pc.onsignalingstatechange = e => warn(pc.signalingState)
pc.oniceconnectionstatechange = e => warn(pc.iceConnectionState)
pc.onicecandidate = event => {
    warn("Generating localDescription");
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
    send_actions_webrts = nil;
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
dc.onmessage = e => warn(`Message from DataChannel '${dc.label}' payload '${e.data}'`)
//dc.send("Connection established!");

// Converts list of actions to command for the server
function get_commands(actions) {
    if (actions.length == 0) {
        return { commands: "stop" };
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
    if (send_actions_webrts) {
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
        localDescriptionSent = false;
        socket = new WebSocket('ws://' + location.host + '/api/tanks/' + tankName + '/connect');

        socket.onopen = function (e) {
            warn(null);
            console.log('WebSocket: opened');
            send_actions();
        }
        socket.onclose = function (e) {
            socket = null;
            warn('WebSocket: closed (' + e.code + ')');
        }
        socket.onmessage = function (e) {
            try {
                var data = JSON.parse(e.data);
                if (data.time) {
                    latency = Date.now() - data.time;
                    console.log('Round trip time ' + latency + " ms");
                    info('Round trip time ' + latency + " ms");
                }
                if (data.answer) {
                    // Finish webrtc connection init with remote answer
                    console.log("Received remote answer:" + data.answer);
                    pc.setRemoteDescription(new RTCSessionDescription(JSON.parse(atob(data.answer)))).catch(warn);
                }
            } catch (err) {
                debug('WebSocket: message ' + e.data);
            }

            // send localDescription if we have one
            if (localDescription != "" && !localDescriptionSent) {
                localDescriptionSent = true;
                console.log("Sending localDescription: " + localDescription);
                socket.send(JSON.stringify({
                    answer: localDescription
                }));
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