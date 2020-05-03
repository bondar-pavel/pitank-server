# Pitank-Server (WIP)

Golang webserver for controlling and monitoring multiple pitanks.

This is Work-In-Progress project and it is far from being complete.

## Build Docker image with project

To build Docker image with binary, static files and templates run: 

```
make image
```

To start newly build image run script:

```
./restart_pitank_server.sh
```

It will start pitank_server on port 80, to test if it works, opend in browser:
```
http://localhost/
```

## Build Project Locally

To build project run:

```
make build
```

It produces `pitank_server` binary. By default server starts on port 8080:

```
$ ./pitank_server -h
Usage of ./pitank_server:
  -port string
        server port (default "8080")
```

## API

### GET /api/tanks

List all registered tanks in the system.

Example:
```
[
   {
      "name":"boo",
      "status":"disconnected",
      "last_registration":"2019-01-11T21:16:24.705987846Z",
      "last_deregistration":"2019-01-11T21:37:00.21155792Z"
   },
   {
      "name":"my-tank",
      "status":"connected",
      "last_registration":"2019-01-11T21:37:14.375007084Z",
      "last_deregistration":"0001-01-01T00:00:00Z"
   }
]
```

### GET /api/tanks/{name}

Show single tank info.

Example:

```
{
   "name":"my-tank",
   "status":"connected",
   "last_registration":"2019-01-11T21:37:14.375007084Z",
   "last_deregistration":"0001-01-01T00:00:00Z"
}
```

### WS /api/tanks/{name}/connect

Open Websocket from client to webserver. Used to send control commands to tank.

Example of data to be sent into WS from client:

```
{"commands": "stop"}
```

### WS /api/connect/{name}

Open Websocket from tank to webserver. Used to receive commands to tank.

Example of data received by tank from server:

```
{"commands": "engine_right engine_forward tower_right"}
```

## Improvement Proposal 1: Use WebRTS for client-tank communication (STATUS: implemented in v0.1.0)

Currently both client and tank are connected via Websocket to the server,
and server forwards commands from Client to Tank.
So data flow looks next:
Client -> Server -> Tank

In case if Client and Tank are located in the same area, but server is on another continent
we can get significant delay.

To improve latency in this case we can use WebRTC protocol to establish direct data channel between
Client and Tank, and Server will act as inremediary only during handshake stage.
In this case latency can be significantly reduced especially in case if Client and Tank are in the same location.

Proposed WebRTC flow:
1. Tank connects to the server over websocket
2. Tank create WebRTC offer
3. Tank send offer to WebSocket
4. Server assotiates offer with Tank
5. Client selects tank in UI
6. Client gets tank offer
7. Client generates answer for this offer
8. Client answer is sent to Websocket
9. Server forwards answer to tank
10. Tank accepts the answer
11. Webrtc connection is started
12. Client can send commands directly to tank over WebRTC channel
13. If WebRTC channel failed, then fallback to Websocket

## Improvement Proposal 2: Use Firestore as main communication media between tank, clients and server (STATUS: In design)

Current approach of communication between tank and clients relies heavily on a server as a central point, which always
keeps Websocket opened for connected parties.

Downsides of this approach are next:
- server contains state, so it is not horizontally scalable, and on restart state is lost;
- server can not be deployed on AppEngine (because of Websockets and state);

Switching to Firestore allows to redesing pitank-server in
cloud-native way to reduces role of the server, since many of it's functions are transferred to Firestore.

Additional benefits of switching to Firestore is an easisy way to add user authentication, which was in TODO list for quite a while.

Implemenentation of WebRTC for pitank to client communication,
allowed to reduce role of Websockets as a media for transferring realtime commands. This allows to replace Websockets with Firebase to configuraton data (WebRTC offer exchange) and WebRTC for realtime, low-delay commands.

TODO: Detailed desing

# History

## 0.1.0

- Implemented WebRTC flow on both sides: pitank-client and pitank-server
- Added status icons for WebRTC/websocket connection on tank control page