# Build stage
FROM golang:1.25-alpine AS builder

# Instalar dependências
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

# Final stage
FROM alpine:latest

# Instalar dependências necessárias
RUN apk --no-cache add ca-certificates tzdata bash

# Criar usuário não-root
RUN addgroup -g 1001 -S ordm && \
    adduser -u 1001 -S ordm -G ordm

# Definir diretório de trabalho
WORKDIR /app

# Copiar binários compilados
COPY --from=builder /app/ordm-node /app/ordm-explorer /app/ordm-monitor ./

# Copiar arquivos estáticos (apenas os que existem)
COPY --from=builder /app/cmd/gui/login_interface.html ./
COPY --from=builder /app/cmd/monitor/dashboard.html ./

# Copiar script de inicialização
COPY scripts/start.sh ./

# Criar diretórios necessários
RUN mkdir -p /tmp/ordm-data /tmp/ordm-data/wallets /tmp/ordm-data/blockchain && \
    chown -R ordm:ordm /app /tmp/ordm-data

# Tornar script executável
RUN chmod +x start.sh

# Mudar para usuário não-root
USER ordm

# Expor porta
EXPOSE 3000

# Comando de inicialização
CMD ["./start.sh"]
