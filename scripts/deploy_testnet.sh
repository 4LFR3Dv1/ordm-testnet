#!/bin/bash

# 🚀 Script de Deploy da Testnet - Seed Nodes
# Este script configura um seed node da testnet em um VPS

set -e

# Configurações
NODE_VERSION="1.0.0"
GO_VERSION="1.25"
NODE_PORT="3001"
API_PORT="8080"
EXPLORER_PORT="8080"
SERVICE_NAME="ordm-testnet"
USER="ordm"

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Função para log colorido
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
}

info() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] INFO: $1${NC}"
}

# Verificar se está rodando como root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        error "Este script deve ser executado como root"
        exit 1
    fi
}

# Atualizar sistema
update_system() {
    log "Atualizando sistema..."
    apt update && apt upgrade -y
    apt install -y curl wget git build-essential ufw
}

# Instalar Go
install_go() {
    log "Instalando Go $GO_VERSION..."
    
    if command -v go &> /dev/null; then
        current_version=$(go version | awk '{print $3}' | sed 's/go//')
        if [[ "$current_version" == "$GO_VERSION" ]]; then
            info "Go $GO_VERSION já está instalado"
            return
        fi
    fi
    
    # Baixar e instalar Go
    wget https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz
    tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
    rm go${GO_VERSION}.linux-amd64.tar.gz
    
    # Configurar PATH
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    source /etc/profile
    
    log "Go $GO_VERSION instalado com sucesso"
}

# Criar usuário para o serviço
create_user() {
    log "Criando usuário $USER..."
    
    if id "$USER" &>/dev/null; then
        info "Usuário $USER já existe"
    else
        useradd -m -s /bin/bash $USER
        usermod -aG sudo $USER
        echo "$USER ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers
        log "Usuário $USER criado"
    fi
}

# Configurar firewall
setup_firewall() {
    log "Configurando firewall..."
    
    ufw --force reset
    ufw default deny incoming
    ufw default allow outgoing
    
    # Portas necessárias
    ufw allow ssh
    ufw allow $NODE_PORT/tcp
    ufw allow $API_PORT/tcp
    ufw allow $EXPLORER_PORT/tcp
    
    # Portas para P2P
    ufw allow 3001:3010/tcp
    ufw allow 3001:3010/udp
    
    ufw --force enable
    log "Firewall configurado"
}

# Baixar e compilar o projeto
build_project() {
    log "Baixando e compilando o projeto..."
    
    # Mudar para o usuário
    su - $USER << 'EOF'
    
    # Criar diretório do projeto
    mkdir -p ~/ordm-testnet
    cd ~/ordm-testnet
    
    # Clonar repositório (substitua pela URL real)
    if [ ! -d "ordm-main" ]; then
        git clone https://github.com/seu-usuario/ordm-main.git
    fi
    
    cd ordm-main
    
    # Baixar dependências
    go mod tidy
    
    # Compilar binários
    go build -o ordm-node ./cmd/node
    go build -o ordm-explorer ./cmd/explorer
    go build -o ordm-backend ./cmd/backend
    
    # Tornar executáveis
    chmod +x ordm-node ordm-explorer ordm-backend
    
    log "Projeto compilado com sucesso"
EOF
}

# Criar diretórios de dados
create_data_dirs() {
    log "Criando diretórios de dados..."
    
    su - $USER << 'EOF'
    mkdir -p ~/ordm-testnet/data
    mkdir -p ~/ordm-testnet/logs
    mkdir -p ~/ordm-testnet/wallets
    mkdir -p ~/ordm-testnet/backups
EOF
}

# Criar arquivo de configuração
create_config() {
    log "Criando arquivo de configuração..."
    
    cat > /home/$USER/ordm-testnet/config.json << 'EOF'
{
    "network": "testnet",
    "node": {
        "port": 3001,
        "api_port": 8080,
        "explorer_port": 8080,
        "max_peers": 50,
        "heartbeat": 30,
        "timeout": 60
    },
    "seed_nodes": [
        "/ip4/18.188.123.45/tcp/3001/p2p/QmSeedNode1",
        "/ip4/52.15.67.89/tcp/3001/p2p/QmSeedNode2",
        "/ip4/34.201.234.56/tcp/3001/p2p/QmSeedNode3"
    ],
    "faucet": {
        "enabled": true,
        "max_amount": 50,
        "daily_limit": 100,
        "rate_limit": 1
    },
    "logging": {
        "level": "info",
        "file": "/home/ordm/ordm-testnet/logs/node.log"
    }
}
EOF

    chown $USER:$USER /home/$USER/ordm-testnet/config.json
}

# Criar systemd service
create_service() {
    log "Criando systemd service..."
    
    cat > /etc/systemd/system/$SERVICE_NAME.service << EOF
[Unit]
Description=ORDM Testnet Node
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=/home/$USER/ordm-testnet
ExecStart=/home/$USER/ordm-testnet/ordm-main/ordm-node
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable $SERVICE_NAME
}

# Criar script de monitoramento
create_monitor_script() {
    log "Criando script de monitoramento..."
    
    cat > /home/$USER/ordm-testnet/monitor.sh << 'EOF'
#!/bin/bash

# Script de monitoramento da testnet

LOG_FILE="/home/ordm/ordm-testnet/logs/monitor.log"
NODE_PID=$(pgrep ordm-node)

log() {
    echo "[$(date)] $1" >> $LOG_FILE
}

# Verificar se o node está rodando
if [ -z "$NODE_PID" ]; then
    log "Node não está rodando, reiniciando..."
    systemctl restart ordm-testnet
    exit 1
fi

# Verificar uso de memória
MEMORY_USAGE=$(ps -p $NODE_PID -o %mem --no-headers)
if (( $(echo "$MEMORY_USAGE > 80" | bc -l) )); then
    log "Uso de memória alto: ${MEMORY_USAGE}%, reiniciando..."
    systemctl restart ordm-testnet
fi

# Verificar uso de CPU
CPU_USAGE=$(ps -p $NODE_PID -o %cpu --no-headers)
if (( $(echo "$CPU_USAGE > 90" | bc -l) )); then
    log "Uso de CPU alto: ${CPU_USAGE}%, reiniciando..."
    systemctl restart ordm-testnet
fi

log "Node funcionando normalmente - Mem: ${MEMORY_USAGE}%, CPU: ${CPU_USAGE}%"
EOF

    chmod +x /home/$USER/ordm-testnet/monitor.sh
    chown $USER:$USER /home/$USER/ordm-testnet/monitor.sh
}

# Configurar cron job para monitoramento
setup_monitoring() {
    log "Configurando monitoramento..."
    
    # Adicionar cron job para monitoramento a cada 5 minutos
    (crontab -u $USER -l 2>/dev/null; echo "*/5 * * * * /home/$USER/ordm-testnet/monitor.sh") | crontab -u $USER -
    
    # Adicionar backup diário
    (crontab -u $USER -l 2>/dev/null; echo "0 2 * * * tar -czf /home/$USER/ordm-testnet/backups/backup-\$(date +\%Y\%m\%d).tar.gz /home/$USER/ordm-testnet/data") | crontab -u $USER -
}

# Iniciar serviços
start_services() {
    log "Iniciando serviços..."
    
    systemctl start $SERVICE_NAME
    
    # Aguardar um pouco e verificar status
    sleep 5
    
    if systemctl is-active --quiet $SERVICE_NAME; then
        log "Serviço iniciado com sucesso"
    else
        error "Falha ao iniciar o serviço"
        systemctl status $SERVICE_NAME
        exit 1
    fi
}

# Mostrar informações finais
show_info() {
    log "Deploy concluído com sucesso!"
    echo
    info "Informações do Seed Node:"
    echo "  - Usuário: $USER"
    echo "  - Diretório: /home/$USER/ordm-testnet"
    echo "  - Porta do Node: $NODE_PORT"
    echo "  - Porta da API: $API_PORT"
    echo "  - Serviço: $SERVICE_NAME"
    echo
    info "Comandos úteis:"
    echo "  - Status: systemctl status $SERVICE_NAME"
    echo "  - Logs: journalctl -u $SERVICE_NAME -f"
    echo "  - Reiniciar: systemctl restart $SERVICE_NAME"
    echo "  - Parar: systemctl stop $SERVICE_NAME"
    echo
    info "Monitoramento:"
    echo "  - Script: /home/$USER/ordm-testnet/monitor.sh"
    echo "  - Logs: /home/$USER/ordm-testnet/logs/"
    echo "  - Backups: /home/$USER/ordm-testnet/backups/"
    echo
    info "Firewall:"
    echo "  - Status: ufw status"
    echo "  - Regras: ufw status numbered"
    echo
    warn "IMPORTANTE: Configure o IP público nos seed nodes da rede!"
}

# Função principal
main() {
    log "Iniciando deploy da testnet..."
    
    check_root
    update_system
    install_go
    create_user
    setup_firewall
    build_project
    create_data_dirs
    create_config
    create_service
    create_monitor_script
    setup_monitoring
    start_services
    show_info
    
    log "Deploy concluído!"
}

# Executar função principal
main "$@"
