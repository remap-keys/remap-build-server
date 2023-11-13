#FROM golang:1.19-buster as builder
FROM golang:1.21-bullseye

RUN apt-get -y update && apt-get -y upgrade && \
    apt-get install -y \
        python3-pip build-essential clang-format diffutils gcc git unzip wget zip \
        binutils-avr gcc-avr avr-libc binutils-arm-none-eabi \
        gcc-arm-none-eabi libnewlib-arm-none-eabi avrdude dfu-programmer \
        dfu-util teensy-loader-cli libhidapi-hidraw0 libusb-dev

WORKDIR /app

ENV HOME /root
ENV PATH=$HOME/.local/bin:$PATH

RUN python3 -m pip install --user qmk

RUN mkdir -p /root/versions/0.22.14
RUN qmk setup --yes --home /root/versions/0.22.14 --branch 0.22.14

COPY go.* ./
RUN go mod download
COPY ./main.go ./
COPY ./auth/*.go ./auth/
COPY ./build/*.go ./build/
COPY ./database/*.go ./database/
COPY ./parameter/*.go ./parameter/
COPY ./web/*.go ./web/
# RUN go test -v ./...
RUN go build -mod=readonly -v -o server

# For local development environment only
#COPY service-account-remap-b2d08-70b4596e8a05.json ./

# EXPOSE 8088

ENTRYPOINT /app/server
