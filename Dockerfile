#FROM golang:1.19-buster as builder
FROM golang:1.19-bullseye

RUN apt-get -y update && apt-get -y upgrade && apt-get install -y git python3-pip

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY ./main.go ./

RUN go test -v ./...
RUN go build -mod=readonly -v -o server

ENV HOME /root
ENV PATH=$HOME/.local/bin:$PATH

RUN python3 -m pip install --user qmk

RUN qmk setup --yes

EXPOSE 8088

ENTRYPOINT /app/server
