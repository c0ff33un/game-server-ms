FROM golang

RUN go get github.com/gorilla/websocket && \
    go get github.com/segmentio/ksuid && \
    go get go.mongodb.org/mongo-driver/mongo

COPY ./ src/github.com/coff33un/game-server-ms

WORKDIR src/github.com/coff33un/game-server-ms/src

EXPOSE 8080

CMD go run *.go
