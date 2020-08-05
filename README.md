Podman Prometheus Exporter
==========================

This application exports basic podman statistics (`man podman-stats`) for
prometheus to scrape. It does not call `podman stats` but instead uses the
`libpod` (now `podman`) library to fetch the statistics.

Build dependencies
------------------

For building on ubuntu/debian you need `libdevmapper-dev`, `libgpgme-dev` and
`libbtrfs-dev`.

Build
-----

Clone and run `go build`

If built on Gentoo (or probably any non debian based system) you may need to
patch the devmapper shared library

```
patchelf --replace-needed libdevmapper.so.1.02 libdevmapper.so \
           podman_prometheus_exporter
```

Or you can use the Dockerfile to build

Usage
-----

```
Usage of podman_prometheus_exporter:
  -l string
          Address to listen on (default ":9901")
```

Install
-------

(optional) Copy the binary to somewhere on your path such as `/usr/local/bin/`.

The included systemd service script can be used to start the service on
boot/keep it running.

```
cp podman_prometheus_exporter.service /etc/systemd/system/podman_prometheus_exporter.service
```

