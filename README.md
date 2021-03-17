# Windows Service management for inlets PRO

## Status

This is an early prototype of a shim to add an inlets PRO HTTP client as a Windows Service, so that it can be started upon reboot, and managed by the system.

Windows Services also allow for restarting of the tunnel process and logging.

## Usage

* Download the binary to C:\windows\
* Create C:\inlets.json (with the contents as per below)

```json
{
 "upstreams": ["openfaas.exit.o6s.io=http://127.0.0.1:8080"],       
 "url": "wss://alex-tunnel.exit.o6s.io",
 "license-file": "C:\\license.txt",
 "token": "TOKEN-HERE",
 "auto-tls": false
}
```

* Create C:\license.txt (with the inlets PRO license)
* Install the service with `inlets-svc install`
* Start the service with `inlets-svc start`

Check the Event Viewer under Application Logs

## License

MIT
