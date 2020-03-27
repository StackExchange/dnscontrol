Gandi Go library
================

WIP: migrating from https://github.com/tiramiseb/go-gandi-livedns

[![GoDoc](https://godoc.org/github.com/tiramiseb/go-gandi?status.svg)](https://godoc.org/github.com/tiramiseb/go-gandi)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/tiramiseb/go-gandi/master/LICENSE)
![Go](https://github.com/tiramiseb/go-gandi/workflows/Go/badge.svg)

This library interacts with [Gandi's API](https://api.gandi.net/docs/), to manage Gandi services. This API returns some data as HTTP headers, please note those are ignored by this library.

A simple CLI is also shipped with this library. It returns responses to the requests in JSON. Build it with `go build -o gandi ./cmd`.
