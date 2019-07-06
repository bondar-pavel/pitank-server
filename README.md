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
