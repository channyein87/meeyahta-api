FROM golang:1.24-bookworm AS builder
WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0
RUN GOOS=linux GOARCH=arm GOARM=7 go build -o /out/meeyahta .

FROM gcr.io/distroless/static:nonroot
WORKDIR /app

# Provide timezone data for Australia/Sydney conversions.
COPY --from=builder /usr/share/zoneinfo/Australia/Sydney /usr/share/zoneinfo/Australia/Sydney
COPY --from=builder /out/meeyahta /app/meeyahta

ENTRYPOINT ["/app/meeyahta"]
