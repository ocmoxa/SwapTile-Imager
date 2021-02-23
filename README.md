# SwapTile-Imager

![Version](https://img.shields.io/github/v/tag/ocmoxa/SwapTile-Imager)
[![Build Status](https://travis-ci.com/ocmoxa/SwapTile-Imager.svg?branch=master)](https://travis-ci.com/ocmoxa/SwapTile-Imager)
[![Go Report Card](https://goreportcard.com/badge/github.com/ocmoxa/SwapTile-Imager)](https://goreportcard.com/report/github.com/ocmoxa/SwapTile-Imager)
[![Coverage Status](https://coveralls.io/repos/github/ocmoxa/SwapTile-Imager/badge.svg?branch=master)](https://coveralls.io/github/ocmoxa/SwapTile-Imager?branch=master)

The server uses Minio for storing images and Redis for storing their
metadata.

# Development

```
docker-compose up
make run
```

# Configuration

See [./config.example.jsonc](./config.example.jsonc).

# Documentation
See [./docs/swagger.yml](./docs/swagger.yml) or browse [http://localhost:8080/swagger/ui](http://localhost:8080/swagger/ui).
