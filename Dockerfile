FROM golang:1.15-buster

ENV APP_USER qrsync_user
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPATH=

RUN groupadd $APP_USER && useradd -m -g $APP_USER -l $APP_USER

USER $APP_USER
COPY . .
RUN go build -o qrsync_server
EXPOSE 4010
RUN ls -l
CMD ./qrsync_server
