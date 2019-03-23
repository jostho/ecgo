# Ecgo server

This is an echo server written in go. This can be used to simulate http responses.

## Environment

* fedora 29
* make 4.2
* golang 1.11
* buildah 1.6

## Build

To check the required dependencies

    make check && make prep

To run

    go run ecgoserver.go

To build run make. The compiled binary is ~ 8 MB in size.

    make
