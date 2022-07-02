FROM golang:1.18

WORKDIR /awillettebackend
COPY . .

RUN apt update
RUN apt install sqlite3
RUN go get -d -v ./...
RUN go install -v ./...
RUN mkdir /awillettebackend/db
EXPOSE 9099

CMD ["willette_api"]
