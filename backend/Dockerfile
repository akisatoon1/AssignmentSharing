FROM golang:1.25-bookworm AS builder

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -v -o /bin/app ./...

FROM scratch
COPY --from=builder /bin/app /bin/app

EXPOSE 8080

CMD ["/bin/app"]
