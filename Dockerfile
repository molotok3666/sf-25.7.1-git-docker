FROM golang AS compiling_stage
RUN mkdir -p /go/src/pipeline
WORKDIR /go/src/pipeline
ADD main.go .
ADD go.mod .
RUN go install .

FROM alpine:latest
LABEL version="1.0.0"
LABEL maintainer="Maksim Bogatyrev<molotok3666@mail.ru>"
WORKDIR /root/
COPY --from=compiling_stage /go/bin/website .
ENTRYPOINT ./pipeline
EXPOSE 8080