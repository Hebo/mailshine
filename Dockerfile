# build stage
FROM golang as build

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build ./cmd/mailshine

# Can't use scratch because we need cgo for sqlite
FROM debian

WORKDIR /app

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /app/ /app/

ARG DATA_DIR=/data
RUN mkdir -p ${DATA_DIR}
ENV DB_PATH=${DATA_DIR}/shine.db

EXPOSE 8000

VOLUME ["/data"]
ENTRYPOINT ["/app/mailshine", "-port", "8000"]
