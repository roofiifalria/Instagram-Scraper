# Etapa de construcción
FROM golang:alpine AS builder

# Establece las variables de entorno necesarias
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Configura el directorio de trabajo
WORKDIR /build

# Copia y descarga las dependencias usando go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copia el código fuente al contenedor
COPY . .

# Compila la aplicación
RUN go build -o main .

# Mueve a /dist el archivo binario compilado
WORKDIR /dist
RUN cp /build/main .

# Imagen mínima para el contenedor final
FROM scratch

# Copia el binario desde la etapa de construcción
COPY --from=builder /dist/main /

# Agrega certificados CA para permitir la verificación del certificado SSL
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Comando para ejecutar la aplicación
ENTRYPOINT ["/main"]
