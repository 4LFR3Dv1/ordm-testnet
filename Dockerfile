# 🏭 Dockerfile para ORDM Blockchain - Executável Integrado
FROM golang:1.25-alpine AS builder

# Instalar dependências necessárias
RUN apk add --no-cache git ca-certificates tzdata

# Definir diretório de trabalho
WORKDIR /app

# Copiar arquivos de dependências
COPY go.mod go.sum ./

# Baixar dependências
RUN go mod download

# Copiar código fonte
COPY . .

# Compilar o executável integrado
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ordmd ./cmd/ordmd

# Imagem final
FROM alpine:latest

# Instalar ca-certificates para HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Criar usuário não-root
RUN addgroup -g 1001 -S blockchain && \
    adduser -u 1001 -S blockchain -G blockchain

# Definir diretório de trabalho
WORKDIR /app

# Copiar apenas o binário compilado
COPY --from=builder /app/ordmd .

# Criar diretório de dados
RUN mkdir -p /app/data && \
    chown -R blockchain:blockchain /app

# Mudar para usuário não-root
USER blockchain

# Expor portas necessárias
EXPOSE 8081 3000

# Variáveis de ambiente
ENV DATA_DIR=/app/data
ENV PORT=8081
ENV P2P_PORT=3000
ENV ORDM_NETWORK=testnet

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT}/api/v1/blockchain/info || exit 1

# Comando padrão
CMD ["./ordmd", "--mode", "both", "--rpc-port", "8081"]
