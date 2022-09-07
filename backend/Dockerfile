FROM golang:1.18
RUN apt update
RUN apt install sqlite3
# RUN go get -d -v ./...
# RUN go install -v ./...
EXPOSE 9099

ARG GIT_COMMIT_ARG=unspecified

ENV GIT_COMMIT=$GIT_COMMIT
WORKDIR /awillettebackend
RUN mkdir /awillettebackend/db
COPY willette_api .

CMD ["./willette_api"]
