# --- Estágio 1: Builder (Compilação) ---
FROM golang:1.23-alpine AS builder

# Instala certificados SSL (necessário para HTTPS/Bancos na nuvem) e git
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Otimização de cache: Copia apenas arquivos de dependência primeiro
COPY go.mod go.sum ./
RUN go mod download

# Copia o código fonte
COPY . .

# Compila o binário
# -ldflags="-w -s": Remove informações de debug para diminuir o tamanho do binário
# -o api: Nome do executável
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o api ./cmd/api/main.go

# --- Estágio 2: Final (Produção) ---
# Usamos 'scratch' (imagem vazia) para segurança máxima e tamanho mínimo
FROM scratch

WORKDIR /app

# Copia os certificados SSL e Zona de Tempo do estágio anterior
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copia APENAS o binário compilado
COPY --from=builder /app/api .

# Expõe a porta (apenas documentação)
EXPOSE 8080

# Define o usuário como não-root (segurança extra, id 65532 é comum em distroless)
USER 65532:65532

# Comando para rodar
ENTRYPOINT ["./api"]