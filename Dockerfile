# build image
FROM golang:alpine as build
WORKDIR /build
RUN apk add --no-cache make
COPY go.sum go.mod Makefile /build/
RUN make go-fetch
COPY . /build/
RUN make

# runtime image
FROM alpine:3.18
COPY --from=build /build/discord-alertmanager /usr/local/bin/discord-alertmanager

# runtime params
WORKDIR /
ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
ENV LOG_JSON=true
CMD ["discord-alertmanager"]
