FROM golang:1.16.3-buster as builder

WORKDIR /build

COPY go.mod go.sum main.go ./
RUN go mod download

RUN apt-get update && apt-get install -y unzip

RUN go build -o ./main ./main.go
RUN wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
RUN wget https://chromedriver.storage.googleapis.com/90.0.4430.24/chromedriver_linux64.zip
RUN unzip ./chromedriver_linux64.zip


FROM ubuntu:20.04 as runner

ENV DEBIAN_FRONTEND=noninteractive

WORKDIR /root

COPY --from=builder /build/main ./
COPY --from=builder /build/chromedriver /usr/local/bin
COPY --from=builder /build/google-chrome-stable_current_amd64.deb ./

RUN apt-get update \
  && apt-get install -y \
    fonts-liberation \
    libasound2 \
    libatk-bridge2.0-0 \ 
    libatk1.0-0 \
    libatspi2.0-0 \
    libcairo2 \
    libcups2 \
    libdrm2 \
    libgbm1 \
    libgtk-3-0 \
    libnspr4 \
    libnss3 \
    libpango-1.0-0 \
    libxcomposite1 \
    libxdamage1 \
    libxfixes3 \
    libxkbcommon0 \
    libxrandr2 \
    libxshmfence1 \
    wget \
    xdg-utils \
    ffmpeg \
  && apt-get -y clean \
  && rm -rf /var/lib/apt/lists/*
RUN dpkg -i google-chrome-stable_current_amd64.deb

ENTRYPOINT [ "/root/main" ]
