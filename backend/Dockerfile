FROM golang:1.18

ARG GIT_COMMIT_ARG=unspecified

ENV GIT_COMMIT=$GIT_COMMIT
WORKDIR /awillettebackend
COPY . .

RUN apt update
RUN apt install sqlite3
RUN go get -d -v ./...
RUN go install -v ./...
RUN mkdir /awillettebackend/db
EXPOSE 9099

CMD ["willette_api"]
