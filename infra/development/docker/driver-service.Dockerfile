FROM alpine
WORKDIR /aclpp

ADD shared shared
ADD build build

ENTRYPOINT build/driver-service