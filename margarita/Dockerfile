FROM golang:1.14-alpine as build
WORKDIR /app/
COPY ./ .
RUN GOOS=linux go build -o app .

FROM alpine:latest
WORKDIR /app/
COPY --from=build /app/app .
CMD ["./app"]