FROM golang:latest
WORKDIR /app

# Cache dependencies
# doesn't work
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copy source code after deps
COPY . .

# Build binary
RUN CGO_ENABLED=0 go build -o andrewwillettedotcom .

EXPOSE 80
EXPOSE 443

CMD ["./andrewwillettedotcom", "serve"]
