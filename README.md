# WRL : Watch, Read, Listen

[![Go Report Card](https://goreportcard.com/badge/github.com/prfalken/wrl)](https://goreportcard.com/report/github.com/prfalken/wrl)
[![Build Status](https://travis-ci.com/prfalken/wrl.svg?branch=master)](https://travis-ci.com/prfalken/wrl)

Revamp application from http://github.com/shawnps/watchreadlisten. Thank you !

Search Imdb, Goodreads, and Spotify:

![Search](http://i.imgur.com/1KrZ0JY.png)

Then save items to a list:

![List](http://i.imgur.com/uYgOmqy.png)

## Installation:
`go get github.com/prfalken/wrl`

## Configure:

Get an OMDB key and a Goodreads keypair, then create a file named config.yml in `$GOPATH/src/github.com/shawnps/watchreadlisten` in the following format:

```YAML
imdb: 
  key: omdb-api-key
gr:
  key: goodreads-key
  secret: goodreads-secret
```
