# build
FROM golang:alpine AS build
RUN apk --no-cache add curl build-base gcc
ADD . /src
WORKDIR /src
RUN make build

# final
FROM alpine
COPY --from=build /src/unshort.link /src/unshort.link
EXPOSE 8080
ENTRYPOINT ["/src/unshort.link"]