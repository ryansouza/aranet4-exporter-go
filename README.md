# aranet4-exporter-go

aranet4-exporter-go is a Prometheus exporter for the [Aranet4](https://aranet.com/products/aranet4/) environmental
sensing devices. Use this to add the CO2, temperature, and humidity readings from your Aranet4 devices to your
Prometheus dashboards.

## Requirements

aranet4-exporter-go should work on most modern Linux system with Bluetooth support.

It has been tested on these platforms:

* x86 Ubuntu 22.04
* Raspberry Pi / Raspbian 3b 32-bit

## Installing

Debian-based distributions can download .deb packages from
the [releases](https://github.com/ryansouza/aranet4-exporter-go/releases).

## Setup

These instructions assume you are on a system using systemd and bluetoothctl.

1. Use bluetoothctl to pair with your Aranet4 device as follows:

    ```shell
    $ bluetoothctl
    > power on
    > scan on
    > scan off
    # find the MAC address of the Aranet4 device in the output.
    > pair XX:XX:..
    # ... enter passcode shown on Aranet4 ...
    > trust XX:XX:..
    ```

2. If `bluetoothctl` is not installed, you may need to install it:

    ```shell
    # Raspbian
    $ sudo apt-get install --no-install-recommends bluetooth pi-bluetooth bluez
    # Other distros
    $ sudo apt-get install --no-install-recommends bluetooth bluez
    ```

3. Add the MAC address and optional nickname of your devices to  `/etc/default/aranet4-exporter`.

4. Run `systemctl restart aranet4-exporter`.

5. Run `journalctl status aranet4-exporter` to get the port that the daemon was started on, and add the URL to your
   Prometheus collector. Example: TODO(ryansouze)

## Releasing

### Building release with goreleaser

To add a new versioned release:

```shell
$ git tag -a v0.1.0 -m "v0.1.0"
$ git push origin v0.1.0
```

### Building snapshot release with goreleaser

Snapshot build:

```shell
$ podman run --rm --privileged \
    -v $PWD:/go/src/github.com/user/repo \
    -w /go/src/github.com/user/repo \
    -e GITHUB_TOKEN \
    docker.io/goreleaser/goreleaser release --snapshot --rm-dist
```

