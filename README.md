# Pitank-Server (WIP)

Golang webserver for controlling and monitoring multiple pitanks.

This is Work-In-Progress project and far from being complete.

## Build Project

To build project run, it will product `pitank_server` binary:

```
make build
```

`pitank_server` by default start on port 8080:

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

Example of data to send into WS from client:

```
{"outputs": "12,14,16"}
```

### WS /api/connect/{name}

Open Websocket from tank to webserver. Used to receive commands to tank.

Example of data received by tank from server:

```
{"outputs": "12,14,16"}
```
