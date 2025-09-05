FROM golang:latest
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o andrewwillettedotcom .
EXPOSE 80
EXPOSE 443
CMD ["./andrewwillettedotcom", "serve"]
