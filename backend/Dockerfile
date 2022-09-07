FROM alpine:3.16.2
# RUN apt update
# RUN apt install sqlite3
RUN apk update && apk upgrade
RUN apk add --no-cache sqlite
# RUN go get -d -v ./...
# RUN go install -v ./...
EXPOSE 9099

ARG GIT_COMMIT_ARG=unspecified

ENV GIT_COMMIT=$GIT_COMMIT
WORKDIR /awillettebackend
RUN mkdir /awillettebackend/db
COPY willette_api .

# CMD ["./willette_api"]
ENTRYPOINT ["tail", "-f", "/dev/null"]
