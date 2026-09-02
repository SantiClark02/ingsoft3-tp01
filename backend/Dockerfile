# syntax=docker/dockerfile:1

# ---------- Etapa 1: build ----------
FROM golang:1.23-alpine AS build

WORKDIR /src

# Dependencias primero: esta capa se cachea y no se rehace
# mientras go.mod y go.sum no cambien.
COPY go.mod go.sum ./
RUN go mod download

# Recién ahora el código fuente
COPY . .

# CGO_ENABLED=0 produce un binario estático, sin dependencias
# de librerías del sistema. Es lo que permite una imagen final mínima.
# -ldflags="-s -w" quita símbolos de debug y achica el binario.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/server .

# ---------- Etapa 2: runtime ----------
FROM alpine:3.20

# ca-certificates para llamadas HTTPS salientes.
# Usuario sin privilegios: si alguien escapa del proceso, no es root.
RUN apk add --no-cache ca-certificates \
    && adduser -D -u 10001 appuser

WORKDIR /app
COPY --from=build /out/server /app/server

USER appuser
EXPOSE 8080

ENTRYPOINT ["/app/server"]