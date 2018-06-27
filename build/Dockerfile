FROM golang:1.10.3-stretch
RUN apt-get update && apt-get install -y zip
RUN go get -u github.com/Masterminds/glide
RUN go get -u -d github.com/golang-migrate/migrate/cli github.com/lib/pq
RUN go build -tags 'postgres' -o /usr/local/bin/migrate github.com/golang-migrate/migrate/cli

ENV DOCKERIZE_VERSION=v0.3.0
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && tar -C /usr/local/bin -xzvf dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && rm dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz

CMD [ "go", "version" ]
