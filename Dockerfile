FROM alpine:latest
RUN apk update && apk upgrade
RUN apk add --no-cache go
EXPOSE 80
EXPOSE 443

ARG GIT_COMMIT_ARG=unspecified

ENV GIT_COMMIT=$GIT_COMMIT
WORKDIR /awillettebackend
RUN mkdir -p /awillettebackend/logging
COPY . .
ENV CGO_ENABLED=1
RUN go build .
CMD ["./willette_api"]
