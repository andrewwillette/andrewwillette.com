FROM alpine:3.19.1
RUN apk update && apk upgrade
RUN apk add --no-cache go=1.21.6-r0
EXPOSE 80
EXPOSE 443

ARG GIT_COMMIT_ARG=unspecified

ENV GIT_COMMIT=$GIT_COMMIT
WORKDIR /awillettebackend
COPY . .
ENV CGO_ENABLED=1
RUN go build .
CMD ["./willette_api"]
