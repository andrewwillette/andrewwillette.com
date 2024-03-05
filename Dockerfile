FROM alpine:latest
RUN apk add --no-cache go
RUN apk update && apk upgrade
EXPOSE 80
EXPOSE 443

ARG GIT_COMMIT_ARG=unspecified

ENV GIT_COMMIT=$GIT_COMMIT
WORKDIR /awillettebackend
COPY . .
ENV CGO_ENABLED=1
RUN go build .
CMD ["./willette_api"]
