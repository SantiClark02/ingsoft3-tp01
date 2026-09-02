# syntax=docker/dockerfile:1

# Imagen base de Golang
FROM golang:1.21

# Variable de entorno para no cachear módulos
ENV GO111MODULE=on

# Directorio de trabajo dentro del contenedor
WORKDIR /app

# Copiamos go.mod y go.sum primero para aprovechar cache
COPY go.mod ./
COPY go.sum ./

# Descargar dependencias
RUN go mod download

# Copiar el resto del código
COPY . ./

# Dar permisos de ejecución al wait-for-it
RUN chmod +x wait-for-it.sh

# Compilar binario
RUN go build -o main .

# Expone el puerto del servidor
EXPOSE 8080

# Comando por defecto
CMD ["./wait-for-it.sh", "mysql:3306", "-t", "60", "--", "./main"]
