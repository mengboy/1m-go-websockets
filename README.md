# Going Infinite, handling 1M websockets connections in Go
This repository holds the complete implementation of the examples seen in Gophercon Israel talk, 2019.

> Going Infinite, handling 1 millions websockets connections in Go / Eran Yanay &mdash; [ [Video](https://www.youtube.com/watch?v=LI1YTFMi8W4) | [Slides](https://speakerdeck.com/eranyanay/going-infinite-handling-1m-websockets-connections-in-go) ]

It doesnt intend or claim to serve as a better, more optimal implementation than other libraries that implements the websocket protocol, it simply shows a set of tools, all combined together to demonstrate a server written in pure Go that is able to serve more than a million websockets connections with less than 1GB of ram.

# Usage
This repository demonstrates how a very high number of websockets connections can be maintained efficiently in Linux

Everything is written in pure Go

Each folder shows an example of a server implementation that overcomes various issues raised by the OS, by the hardware or the Go runtime itself, as shown during the talk.

`setup.sh` is a wrapper to running multiple instances using Docker. See content of the script for more details of how to use it.

`destroy.sh` is a wrapper to stop all running clients.

A single client instance can be executed by running `go run client.go -conn=<# connections to establish>`


# Attention
Because of gorilla/websocket not support evented read，there will be some issue when client send message high frequency.
There is a [discussion](https://github.com/gorilla/websocket/issues/481) about if support evented read.

由于gorilla/websocket不支持事件读，在epoll用会有一些问题，参考[gorilla/websocket是否支持事件读的讨论](https://github.com/gorilla/websocket/issues/481)。
具体表现为客户端高频向后端发消息是，不会每次都触发epoll事件。
