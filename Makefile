.PHONY: all

all:
	go build ./...
	patchelf --replace-needed libdevmapper.so.1.02 libdevmapper.so podman_prometheus_exporter
