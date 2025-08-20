# Estágio de build
FROM golang:1.23-alpine AS builder

WORKDIR /app
RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api

# Estágio final
FROM golang:1.23-alpine

WORKDIR /app

# Instalar dependências de tempo de execução
RUN apk --no-cache add ca-certificates

# Copiar o binário do builder
COPY --from=builder /app/main .

# Copiar o arquivo .env (certifique-se de incluí-lo em .dockerignore se não for necessário em produção)
COPY .env .

# Expor a porta em que a aplicação é executada
EXPOSE 8080

# Comando para executar o executável
CMD ["./main"]