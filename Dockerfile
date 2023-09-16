# syntax=docker/dockerfile:1


FROM golang:1.21 AS BUILD



WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /pull_topic


FROM scratch
COPY --from=BUILD /pull_topic /pull_topic

EXPOSE 8082
CMD [ "/pull_topic" ]

