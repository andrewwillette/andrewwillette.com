FROM alpine:3.16.2
RUN apk update && apk upgrade
RUN apk add --no-cache go
EXPOSE 9099

ARG GIT_COMMIT_ARG=unspecified

ENV GIT_COMMIT=$GIT_COMMIT
WORKDIR /awillettebackend
COPY . .
ENV CGO_ENABLED=1
RUN go build .
CMD ["./willette_api"]
