FROM golang:1.19

RUN apt-get update && apt-get install uuid-dev uuid-runtime
WORKDIR /build
COPY . .
RUN make
