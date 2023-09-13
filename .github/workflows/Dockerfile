FROM golang:1.21

# Injected from the build
ARG VERSION

RUN apt-get update && apt-get install uuid-dev uuid-runtime

# make it
WORKDIR /build
COPY . .
RUN make

# Try it
RUN ./samler
