FROM golang:1.21.4-alpine3.18

ENV GO111MODULE on
ENV CGO_ENABLED 1
ENV GOOS linux
ENV GOARCH amd64

WORKDIR /usr/src/app

COPY . .

# RUN go mod tidy

# RUN go mod vendor

RUN go build -o bin/main main/main.go

#CMD ["sudo", "chmod", "+x", "bin/main"]

CMD ["sh", "-c", "chmod +x bin/main && ./bin/main -mode prod"]
