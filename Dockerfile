FROM golang:1.12 AS build
WORKDIR /app
ENV CGO_ENABLED=0
ENV GOOS=linux
ADD ./ .
RUN wget -O data.tar.gz https://geolite.maxmind.com/download/geoip/database/GeoLite2-City.tar.gz
RUN tar xvzf data.tar.gz --strip=1
RUN go build -o app.cmd -ldflags "-s -w -extldflags -static" .

FROM scratch
ENV GIN_MODE=release
COPY --from=build /app/app.cmd /
COPY --from=build /app/GeoLite2-City.mmdb /
CMD ["/app.cmd"]