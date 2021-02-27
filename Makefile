.POSIX:

SHELL:=/bin/sh

GO:=`which go`
PATCHELF:=`which patchelf`

TARGET:=./podman_prometheus_exporter

.PHONY: default
default: all

podman_prometheus_exporter: main.go
	@$(GO) build -o $(TARGET) ./main.go
	@$(PATCHELF) --replace-needed libdevmapper.so.1.02 libdevmapper.so $(TARGET)

.PHONY: clean
clean:
	@rm -f $(TARGET)

.PHONY: all
all: podman_prometheus_exporter
