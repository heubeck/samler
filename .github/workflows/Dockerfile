FROM golang:1.25

# Injected from the build
ARG VERSION

RUN apt-get update -y && apt-get install -y uuid-dev uuid-runtime

# make it
WORKDIR /build
COPY . .
RUN make

# Try it
RUN ./samler
