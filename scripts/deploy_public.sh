#!/bin/bash

# Script de Deploy Público para ORDM Testnet
# Configuração completa para VPS com domínio e SSL

set -e

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
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
}

# Verificar se é root
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
    apt install -y curl wget git build-essential ufw nginx certbot python3-certbot-nginx
}

# Instalar Go
install_go() {
    log "Instalando Go 1.25..."
    wget https://go.dev/dl/go1.25.linux-amd64.tar.gz
    tar -C /usr/local -xzf go1.25.linux-amd64.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    source /etc/profile
    rm go1.25.linux-amd64.tar.gz
}

# Criar usuário
create_user() {
    log "Criando usuário ordm..."
    useradd -m -s /bin/bash ordm
    usermod -aG sudo ordm
    echo "ordm:ordm_secure_password_2024" | chpasswd
}

# Configurar firewall
setup_firewall() {
    log "Configurando firewall..."
    ufw --force reset
    ufw default deny incoming
    ufw default allow outgoing
    ufw allow ssh
    ufw allow 80/tcp
    ufw allow 443/tcp
    ufw allow 3000/tcp  # Node GUI
    ufw allow 8080/tcp  # Explorer
    ufw allow 9090/tcp  # Monitor
    ufw allow 3001/tcp  # P2P
    ufw --force enable
}

# Configurar Nginx
setup_nginx() {
    log "Configurando Nginx..."
    
    # Configuração para o Explorer
    cat > /etc/nginx/sites-available/ordm-explorer << 'EOF'
server {
    listen 80;
    server_name explorer.ordm-testnet.com;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
EOF

    # Configuração para o Monitor
    cat > /etc/nginx/sites-available/ordm-monitor << 'EOF'
server {
    listen 80;
    server_name monitor.ordm-testnet.com;
    
    location / {
        proxy_pass http://localhost:9090;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
EOF

    # Configuração para a API
    cat > /etc/nginx/sites-available/ordm-api << 'EOF'
server {
    listen 80;
    server_name api.ordm-testnet.com;
    
    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
EOF

    # Ativar sites
    ln -sf /etc/nginx/sites-available/ordm-explorer /etc/nginx/sites-enabled/
    ln -sf /etc/nginx/sites-available/ordm-monitor /etc/nginx/sites-enabled/
    ln -sf /etc/nginx/sites-available/ordm-api /etc/nginx/sites-enabled/
    
    # Remover site padrão
    rm -f /etc/nginx/sites-enabled/default
    
    # Testar e reiniciar Nginx
    nginx -t
    systemctl restart nginx
    systemctl enable nginx
}

# Configurar SSL
setup_ssl() {
    log "Configurando SSL..."
    
    # Certbot para cada domínio
    certbot --nginx -d explorer.ordm-testnet.com --non-interactive --agree-tos --email admin@ordm-testnet.com
    certbot --nginx -d monitor.ordm-testnet.com --non-interactive --agree-tos --email admin@ordm-testnet.com
    certbot --nginx -d api.ordm-testnet.com --non-interactive --agree-tos --email admin@ordm-testnet.com
    
    # Renovação automática
    echo "0 12 * * * /usr/bin/certbot renew --quiet" | crontab -
}

# Compilar projeto
build_project() {
    log "Compilando projeto..."
    cd /home/ordm
    
    # Clonar repositório
    git clone https://github.com/your-username/ordm-main.git
    cd ordm-main
    
    # Instalar dependências
    go mod tidy
    
    # Compilar binários
    go build -o ordm-node ./cmd/gui
    go build -o ordm-explorer ./cmd/explorer
    go build -o ordm-monitor ./cmd/monitor
    
    # Mover para /usr/local/bin
    mv ordm-node /usr/local/bin/
    mv ordm-explorer /usr/local/bin/
    mv ordm-monitor /usr/local/bin/
    
    # Criar diretórios de dados
    mkdir -p /var/lib/ordm/{data,logs,backups}
    chown -R ordm:ordm /var/lib/ordm
}

# Criar serviços systemd
create_services() {
    log "Criando serviços systemd..."
    
    # Serviço do Node
    cat > /etc/systemd/system/ordm-node.service << 'EOF'
[Unit]
Description=ORDM Testnet Node
After=network.target

[Service]
Type=simple
User=ordm
WorkingDirectory=/var/lib/ordm
ExecStart=/usr/local/bin/ordm-node
Restart=always
RestartSec=10
Environment=HOME=/home/ordm

[Install]
WantedBy=multi-user.target
EOF

    # Serviço do Explorer
    cat > /etc/systemd/system/ordm-explorer.service << 'EOF'
[Unit]
Description=ORDM Testnet Explorer
After=network.target

[Service]
Type=simple
User=ordm
WorkingDirectory=/var/lib/ordm
ExecStart=/usr/local/bin/ordm-explorer
Restart=always
RestartSec=10
Environment=HOME=/home/ordm

[Install]
WantedBy=multi-user.target
EOF

    # Serviço do Monitor
    cat > /etc/systemd/system/ordm-monitor.service << 'EOF'
[Unit]
Description=ORDM Testnet Monitor
After=network.target

[Service]
Type=simple
User=ordm
WorkingDirectory=/var/lib/ordm
ExecStart=/usr/local/bin/ordm-monitor
Restart=always
RestartSec=10
Environment=HOME=/home/ordm

[Install]
WantedBy=multi-user.target
EOF

    # Recarregar systemd
    systemctl daemon-reload
    
    # Habilitar serviços
    systemctl enable ordm-node
    systemctl enable ordm-explorer
    systemctl enable ordm-monitor
}

# Configurar monitoramento
setup_monitoring() {
    log "Configurando monitoramento..."
    
    # Script de monitoramento
    cat > /usr/local/bin/ordm-monitor.sh << 'EOF'
#!/bin/bash

# Verificar se os serviços estão rodando
check_service() {
    local service=$1
    if ! systemctl is-active --quiet $service; then
        echo "$(date): $service não está rodando. Reiniciando..."
        systemctl restart $service
    fi
}

# Verificar serviços
check_service ordm-node
check_service ordm-explorer
check_service ordm-monitor

# Verificar uso de disco
DISK_USAGE=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
if [ $DISK_USAGE -gt 80 ]; then
    echo "$(date): Uso de disco alto: ${DISK_USAGE}%"
fi

# Verificar uso de memória
MEM_USAGE=$(free | awk 'NR==2{printf "%.0f", $3*100/$2}')
if [ $MEM_USAGE -gt 80 ]; then
    echo "$(date): Uso de memória alto: ${MEM_USAGE}%"
fi
EOF

    chmod +x /usr/local/bin/ordm-monitor.sh
    
    # Adicionar ao crontab
    echo "*/5 * * * * /usr/local/bin/ordm-monitor.sh >> /var/log/ordm-monitor.log 2>&1" | crontab -
}

# Configurar backup
setup_backup() {
    log "Configurando backup automático..."
    
    # Script de backup
    cat > /usr/local/bin/ordm-backup.sh << 'EOF'
#!/bin/bash

BACKUP_DIR="/var/lib/ordm/backups"
DATE=$(date +%Y%m%d_%H%M%S)

# Criar backup dos dados
tar -czf $BACKUP_DIR/ordm_data_$DATE.tar.gz -C /var/lib/ordm data/

# Manter apenas os últimos 7 backups
find $BACKUP_DIR -name "ordm_data_*.tar.gz" -mtime +7 -delete

echo "Backup criado: ordm_data_$DATE.tar.gz"
EOF

    chmod +x /usr/local/bin/ordm-backup.sh
    
    # Backup diário às 2h da manhã
    echo "0 2 * * * /usr/local/bin/ordm-backup.sh >> /var/log/ordm-backup.log 2>&1" | crontab -
}

# Iniciar serviços
start_services() {
    log "Iniciando serviços..."
    systemctl start ordm-node
    systemctl start ordm-explorer
    systemctl start ordm-monitor
    
    # Aguardar inicialização
    sleep 10
    
    # Verificar status
    systemctl status ordm-node --no-pager
    systemctl status ordm-explorer --no-pager
    systemctl status ordm-monitor --no-pager
}

# Mostrar informações finais
show_info() {
    log "Deploy concluído com sucesso!"
    echo
    echo "🌐 URLs da Testnet:"
    echo "  Explorer: https://explorer.ordm-testnet.com"
    echo "  Monitor:  https://monitor.ordm-testnet.com"
    echo "  API:      https://api.ordm-testnet.com"
    echo
    echo "🔧 Comandos úteis:"
    echo "  Status:   systemctl status ordm-*"
    echo "  Logs:     journalctl -u ordm-node -f"
    echo "  Restart:  systemctl restart ordm-node"
    echo
    echo "📊 Monitoramento:"
    echo "  Logs:     /var/log/ordm-monitor.log"
    echo "  Backup:   /var/log/ordm-backup.log"
    echo "  Dados:    /var/lib/ordm/"
    echo
    echo "🔐 Segurança:"
    echo "  Firewall: ufw status"
    echo "  SSL:      certbot certificates"
    echo
    echo "📝 Próximos passos:"
    echo "  1. Configurar DNS para os domínios"
    echo "  2. Atualizar TESTNET_README.md com as URLs"
    echo "  3. Testar todos os endpoints"
    echo "  4. Configurar alertas de monitoramento"
}

# Função principal
main() {
    log "Iniciando deploy público da ORDM Testnet..."
    
    check_root
    update_system
    install_go
    create_user
    setup_firewall
    setup_nginx
    setup_ssl
    build_project
    create_services
    setup_monitoring
    setup_backup
    start_services
    show_info
}

# Executar função principal
main "$@"
