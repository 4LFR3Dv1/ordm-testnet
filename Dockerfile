# Dockerfile para ORDM Testnet
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

# Compilar aplicações
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ordm-node ./cmd/gui
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ordm-explorer ./cmd/explorer
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ordm-monitor ./cmd/monitor

# Imagem final
FROM alpine:latest

# Instalar ca-certificates para HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Criar usuário não-root
RUN addgroup -g 1001 -S ordm && \
    adduser -u 1001 -S ordm -G ordm

# Definir diretório de trabalho
WORKDIR /app

# Copiar binários compilados
COPY --from=builder /app/ordm-node /app/
COPY --from=builder /app/ordm-explorer /app/
COPY --from=builder /app/ordm-monitor /app/

# Copiar arquivos estáticos
COPY --from=builder /app/cmd/monitor/dashboard.html /app/

# Criar diretórios necessários
RUN mkdir -p /app/data /app/logs /app/backups /app/wallets && \
    chown -R ordm:ordm /app

# Mudar para usuário não-root
USER ordm

# Expor portas
EXPOSE 3000 8080 9090

# Script de inicialização
COPY --chown=ordm:ordm scripts/start.sh /app/
RUN chmod +x /app/start.sh

# Comando padrão
CMD ["/app/start.sh"]
