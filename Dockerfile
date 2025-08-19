# Estágio de build
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Instala as dependências de compilação necessárias
RUN apk add --no-cache gcc musl-dev

# Copia os arquivos de dependências primeiro para aproveitar o cache
COPY go.mod go.sum ./
RUN go mod download

# Copia o restante do código
COPY . .

# Compila a aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api

# Estágio final
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copia o binário do estágio de build
COPY --from=builder /app/main .

# Expõe a porta que a aplicação vai rodar
EXPOSE 8080

# Comando para executar a aplicação
CMD ["./main"]