FROM golang:1.25 AS build

RUN apt update -y && apt install -y git make g++ bash

ENV GIT_TERMINAL_PROMPT=0
ENV GOPROXY=https://proxy.golang.org,https://goproxy.io,https://goproxy.dev

WORKDIR /hasher

RUN mkdir bin

COPY main.go /entrypoint/main.go

CMD ["go", "run", "/entrypoint/main.go"]
