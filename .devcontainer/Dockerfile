FROM golang:1.25-bookworm

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends git \
    && rm -rf /var/lib/apt/lists/*

RUN go install github.com/air-verse/air@latest \
    && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

COPY go.mod ./

RUN go mod download

COPY . .

COPY .devcontainer/docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.sh
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]

EXPOSE 3000

CMD ["air", "-c", ".air.toml"]
