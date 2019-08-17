# build stage
FROM golang:alpine AS build-env
WORKDIR /root
RUN apk --no-cache add build-base git
ADD . /root
RUN env GO111MODULE=on go build -o exporter cmd/main.go

# final stage
FROM alpine
WORKDIR /root
RUN apk --no-cache add bash
COPY run.sh .
RUN chmod u+x run.sh
COPY --from=build-env /root/exporter /root
CMD [ "./run.sh" ]