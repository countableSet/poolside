FROM golang:1.15 as build
WORKDIR /app/
COPY ./ .
RUN GOOS=linux go build -o app .

FROM ubuntu:20.04
ENV CONFIG_PATH=/app
WORKDIR /app/
COPY --from=build /app/app .
COPY --from=build /app/public/* public/
CMD ["./app"]