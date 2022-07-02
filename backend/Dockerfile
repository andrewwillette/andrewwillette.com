FROM golang:1.18

WORKDIR /goApp
COPY . .

RUN apt update
RUN apt install sqlite3
RUN go get -d -v ./...
RUN go install -v ./...
# add directory to insert mounted sqlite db to
RUN mkdir /goApp/db
EXPOSE 9099

CMD ["willette_api"]
