FROM golang:latest
EXPOSE 80
EXPOSE 443
WORKDIR /awillettebackend
COPY . .
ENV CGO_ENABLED=0
RUN go build .
CMD ["./willette_api"]
