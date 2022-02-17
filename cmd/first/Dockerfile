FROM golang:1.17 as builder

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/src/app/app ./cmd/first/main.go

FROM golang:1.17

COPY --from=builder /usr/src/app/app /

CMD ["/app"]
