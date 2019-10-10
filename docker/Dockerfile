FROM golang
EXPOSE 8080
COPY ./ src/github.com/coff33un/game-server-ms
RUN go get github.com/gorilla/websocket && \
    go get github.com/segmentio/ksuid

WORKDIR src/github.com/coff33un/game-server-ms/src
CMD go run main.go
