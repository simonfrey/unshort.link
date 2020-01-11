# build
FROM golang:alpine AS build
RUN apk --no-cache add curl build-base gcc
ADD . /src
WORKDIR /src
RUN go generate ./... && go generate ./db/db.go && GOOS=linux go build -o app

# final
FROM alpine
WORKDIR /app
COPY --from=build /src/app /app/
ENTRYPOINT ["./app"]