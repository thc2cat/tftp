FROM golang:alpine as build
RUN apk --no-cache add tzdata
WORKDIR /app
ADD . .
RUN CGO_ENABLED=0 GOOS=linux go build -o myapp

FROM scratch as final
COPY --from=build /app/myapp tftpd
COPY --from=build /usr/local/go/lib/time/zoneinfo.zip /
ENV TZ=Europe/Paris
ENV ZONEINFO=/zoneinfo.zip
CMD ["/tftpd"]