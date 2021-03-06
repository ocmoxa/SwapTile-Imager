# SwapTile-Imager

![Version](https://img.shields.io/github/v/tag/ocmoxa/SwapTile-Imager)
[![Build Status](https://travis-ci.org/ocmoxa/SwapTile-Imager.svg?branch=main)](https://travis-ci.org/ocmoxa/SwapTile-Imager)
[![Go Report Card](https://goreportcard.com/badge/github.com/ocmoxa/SwapTile-Imager)](https://goreportcard.com/report/github.com/ocmoxa/SwapTile-Imager)
[![Coverage Status](https://coveralls.io/repos/github/ocmoxa/SwapTile-Imager/badge.svg?branch=main)](https://coveralls.io/github/ocmoxa/SwapTile-Imager?branch=master)

The server uses Minio for storing images and Redis for storing their
metadata. The server supports efficient image resizing.

# Development

```
docker-compose up
```

# Configuration

See [./config.example.jsonc](./config.example.jsonc).

# Requirnments

* Go 1.16
* libvips-dev

# Documentation
See [./docs/swagger.yml](./docs/swagger.yml) or browse [http://localhost:8080/swagger/ui](http://localhost:8080/swagger/ui).
