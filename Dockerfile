FROM golang:1.16.3-buster as builder

WORKDIR /build

COPY go.mod go.sum main.go ./
RUN go mod download

RUN apt-get update && apt-get install -y unzip

RUN go build -o ./main ./main.go
RUN wget https://chromedriver.storage.googleapis.com/90.0.4430.24/chromedriver_linux64.zip
RUN unzip ./chromedriver_linux64.zip


FROM ubuntu:20.04 as runner

ENV DEBIAN_FRONTEND=noninteractive

WORKDIR /root

COPY --from=builder /build/main ./
COPY --from=builder /build/chromedriver /usr/local/bin

RUN apt-get update \
  && apt-get -y install wget gnupg \
  && wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add - \
  && echo 'deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main' | tee /etc/apt/sources.list.d/google-chrome.list \
  && apt-get update \
  && apt-get -y install google-chrome-stable ffmpeg \
  && apt-get -y clean \
  && rm -rf /var/lib/apt/lists/*

ENTRYPOINT [ "/root/main" ]
