FROM golang:1.12
WORKDIR /app
ADD ./ .
RUN wget -O data.tar.gz https://geolite.maxmind.com/download/geoip/database/GeoLite2-City.tar.gz
RUN tar xvzf data.tar.gz --strip=1
