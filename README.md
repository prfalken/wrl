# watchreadlisten

[![Go Report Card](https://goreportcard.com/badge/github.com/shawnps/watchreadlisten)](https://goreportcard.com/report/github.com/shawnps/watchreadlisten)

Search Rotten Tomatoes, Goodreads, and Spotify:

![Search](http://i.imgur.com/1KrZ0JY.png)

Then save items to a list:

![List](http://i.imgur.com/uYgOmqy.png)

## Installation:
`go get github.com/shawnps/watchreadlisten`

## Configure:

Get a Rotten Tomatoes key and a Goodreads keypair, then create a file named config.yml in `$GOPATH/src/github.com/shawnps/watchreadlisten` in the following format:

```YAML
rt: rotten-tomatoes-api-key
gr:
  key: goodreads-key
  secret: goodreads-secret
```
