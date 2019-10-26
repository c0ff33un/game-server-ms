FROM golang

COPY ./src/src /app/

WORKDIR /app

ENTRYPOINT ["./src"]
