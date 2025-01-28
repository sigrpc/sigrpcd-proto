FROM golang:1.22-bookworm AS builder

COPY ./ /sigrpcd
WORKDIR /sigrpcd/cmd/x64
RUN go build sigrpcd.go

FROM gcr.io/distroless/base-debian12

COPY --from=builder /sigrpcd/cmd/x64/sigrpcd /usr/local/bin/sigrpcd

USER 1001
