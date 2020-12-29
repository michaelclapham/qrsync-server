FROM golang:1.15-buster

ENV APP_USER qrsync_user
ENV APP_HOME /go/src/qrsync_server

RUN groupadd $APP_USER && useradd -m -g $APP_USER -l $APP_USER
RUN mkdir -p $APP_HOME && chown -R $APP_USER:$APP_USER $APP_HOME

WORKDIR $APP_HOME
USER $APP_USER
COPY . $APP_HOME
RUN go build -o qrsync_server
EXPOSE 4001
CMD ["./qrsync_server"]
