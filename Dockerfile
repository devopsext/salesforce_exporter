FROM golang:1.18.2-alpine3.15 AS build_base

ARG XPATE_GROUP_CI_TOKEN=abc

ENV GO_WORKDIR $GOPATH/prometheus/salesforce_exporter

# Set working directory
WORKDIR $GO_WORKDIR
COPY go.mod ./
COPY go.sum ./

# Install deps
RUN apk add --no-cache git
RUN go mod download
RUN go mod verify

COPY . .

RUN GOOS=linux go build -o /salesforce_exporter

# Start fresh from a smaller image
FROM alpine:3.15.4
RUN apk add ca-certificates tzdata

RUN addgroup --gid 2022 salesforce && \
    adduser --disabled-password --uid 2022 --ingroup salesforce --gecos salesforce salesforce

COPY --chown=saleforce:salesforce --from=build_base /salesforce_exporter /usr/local/salesforce_exporter

USER salesforce

EXPOSE 9141

CMD [ "/use/local/salesforce_exporter" ]
