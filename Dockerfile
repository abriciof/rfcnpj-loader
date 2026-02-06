# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/rfcnpj-loader ./cmd/rfcnpj-loader

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /out/rfcnpj-loader /usr/local/bin/rfcnpj-loader
# default paths inside container
ENV OUTPUT_FILES_PATH=/data/output     EXTRACTED_FILES_PATH=/data/extracted
VOLUME ["/data"]
ENTRYPOINT ["rfcnpj-loader"]
