# syntax=docker/dockerfile:1

FROM golang:1.20.10-alpine AS build


WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /ptp .

FROM scratch

COPY --from=build /ptp /ptp

EXPOSE 80

ENTRYPOINT ["/ptp"]
