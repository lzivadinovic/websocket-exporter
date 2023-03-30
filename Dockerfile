FROM golang:alpine as build
WORKDIR /app
COPY *.go .
COPY go.* .
RUN go get
RUN go build -o wsgo


FROM alpine:3
WORKDIR /app
ENV port=9143
ENV host=0.0.0.0
COPY --from=build /app/wsgo .
EXPOSE $port
ENTRYPOINT /app/wsgo --host ${host} --port ${port}