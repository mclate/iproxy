Simple SOCKS5/HTTP proxy to overcome tethering limitations.

This tool is initially designed and tested on iPhone 13.
It probably will work on Android, but it was never tested there.

## Features

* SOCKS5 proxy
* HTTP proxy
* Proxy auto discovery
* Prevents application hibernation by polling device location

## How to use 

### On your phone:

1. Install [iSH](https://ish.app/)
2. Fetch the latest release of iProxy
3. Run the app

CLI arguments:

* `-a <ip>` - use specific ip as an exposed host address for proxy.
  This should be the ip of your phone within the tethered wifi (for iPhone this most likely will be `172.20.10.1`)
  This ip will be used to bind proxy to.
* `-b <ip>` - address to bind to. By default, it's `0.0.0.0` meaning that we listen on all interfaces.
* `-d <port>` - enable Proxy auto configuration. Will serve requests on a given port under `/proxy` endpoint
* `-s <port>` - enable SOCKS5 on a given port
* `-p <port>` - enable HTTP proxy ona given port
* `-l` - enable location tracking. This helps with preventing app from hibernating.
  This is the only reason why app will ask for location permission.
  Make sure to select `Always allow`, otherwise it doesn't make sense to use it.
* `-h` - show help message with command line arguments
* `-v` - enable verbose output

### On your computer. 

There are two options: 

1. (Option 1) Configure Proxy auto discovery
2. (Option 2) Manually configure proxy for each application.

## Proxy auto discovery

When started with `-d <port>` flag, iProxy will allow one to use [Proxy auto config](https://en.wikipedia.org/wiki/Proxy_auto-config)
protocol under `/proxy` endpoint. Most applications and even OS itself are capable of using this feature. 

In order to set it up, one has to enable "Automatic Proxy Configuration" and provide the url of the phone, i.e.,
`http://172.20.10.1:<port>/proxy` where `172.20.10.1` is the phone router ip (when tethering) - this is the default
IP for iPhone, Android will have a different one. `<port>` is the one specified by `-d <port>` parameter.

Automatic Proxy Discovery will instruct the device to use SOCKS5 proxy first, if it is not available, it will try HTTP
proxy and in the end fallback to a direct connection. SOCKS and HTTP proxy will only be provided if they were enabled in
the iProxy command line parameters.

## Using proxy with different apps

In most cases, browsers will pull and use proxy configuration from the system.
Some other application might do it as well, however, in most cases you have to refer to the application documentation in 
order to find out how to configure it.

Most CLI applications will respect `http_proxy` or `all_proxy` environment variables.

## Known issues

### App is being hibernated/stop working

Yes, this is a known issue when iOS hibernates (or terminates) an app that is running in the background.
Allowing location use and running an app with `-l` parameter might help a bit, but in some cases app will be terminated regardless.

### How to make git work with proxy

Notoriously, git, when working through ssh, doesn't respect `all_proxy` env var.

One can use next env variable to make git work through SOCKS5:

```shell
export GIT_SSH_COMMAND='ssh -o ProxyCommand="nc -X 5 -x <phone ip>:<socks5 port> %h %p"'
```

### Missing proxy authentication

This is by design. Because the only use case for this app is to be working over tethered wifi, we assume that one only 
allows access to his hotspot to a known devices. Thus, authentication is done by the means of the hotspot.


## Contributions

Issues and PRs are welcome
