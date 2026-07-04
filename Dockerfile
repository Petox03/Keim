FROM golang:1.26-alpine

WORKDIR /app

# git necesario para go get de dependencias
RUN apk add --no-cache git

# Redirigir cachés de Go a rutas que serán persistidas vía volúmenes nombrados
ENV GOCACHE=/root/.cache/go-build
ENV GOMODCACHE=/root/.cache/go-mod

CMD ["sleep", "infinity"]
