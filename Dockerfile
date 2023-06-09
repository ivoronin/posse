FROM golang:1.20 as builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
ADD fsm ./fsm
ADD metrics ./metrics
ADD peer ./peer
ADD block ./block
ADD disk ./disk
ADD tun ./tun
ENV CGO_ENABLED=0
RUN go build -o posse

FROM scratch as runner
COPY --from=builder /build/posse /posse
ENTRYPOINT [ "/posse" ]