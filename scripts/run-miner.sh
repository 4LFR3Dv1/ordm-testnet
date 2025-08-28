#!/bin/bash

# ⛏️ Script para rodar ORDM Miner da Testnet
# Uso: ./scripts/run-miner.sh [opções]

set -e

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configurações padrão
RPC_URL="http://localhost:8081"
MINER_KEY=""
THREADS=1
DIFFICULTY=4
DATA_PATH="./data/miner"
MINER_NAME="ordm-miner"

# Função para exibir ajuda
show_help() {
    echo -e "${BLUE}⛏️ ORDM Miner - Script de Execução${NC}"
    echo ""
    echo "Uso: $0 [opções]"
    echo ""
    echo "Opções obrigatórias:"
    echo "  --miner-key KEY          Chave privada do minerador (obrigatório)"
    echo ""
    echo "Opções opcionais:"
    echo "  --rpc URL                URL do node RPC [padrão: http://localhost:8081]"
    echo "  --threads NUM            Número de threads de mineração [padrão: 1]"
    echo "  --difficulty NUM         Dificuldade de mineração [padrão: 4]"
    echo "  -d, --data PATH          Caminho para dados [padrão: ./data/miner]"
    echo "  --name NAME              Nome do minerador [padrão: ordm-miner]"
    echo "  -h, --help               Exibir esta ajuda"
    echo ""
    echo "Exemplos:"
    echo "  $0 --miner-key abc123                                    # Minerador básico"
    echo "  $0 --miner-key abc123 --threads 4 --rpc http://node:8081 # Minerador com 4 threads"
    echo "  $0 --miner-key abc123 --name my-miner                    # Minerador com nome personalizado"
    echo ""
    echo "Notas:"
    echo "  - Certifique-se de que o node RPC está rodando antes de iniciar o minerador"
    echo "  - A chave privada é usada para identificar o minerador na rede"
    echo "  - Use --threads para aumentar a potência de mineração (cuidado com o CPU)"
    echo ""
}

# Função para verificar dependências
check_dependencies() {
    echo -e "${BLUE}🔍 Verificando dependências...${NC}"
    
    # Verificar se Go está instalado
    if ! command -v go &> /dev/null; then
        echo -e "${RED}❌ Go não está instalado. Instale Go 1.19+ primeiro.${NC}"
        exit 1
    fi
    
    # Verificar se o binário existe
    if [ ! -f "./cmd/ordm-miner/main.go" ]; then
        echo -e "${RED}❌ Binário ordm-miner não encontrado. Execute 'go build ./cmd/ordm-miner' primeiro.${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ Dependências verificadas${NC}"
}

# Função para verificar conectividade com o node
check_node_connectivity() {
    echo -e "${BLUE}🔗 Verificando conectividade com o node...${NC}"
    
    # Extrair host e porta da URL RPC
    RPC_HOST=$(echo "$RPC_URL" | sed 's|http://||' | cut -d: -f1)
    RPC_PORT=$(echo "$RPC_URL" | sed 's|http://||' | cut -d: -f2)
    
    # Verificar se a porta está acessível
    if ! nc -z "$RPC_HOST" "$RPC_PORT" 2>/dev/null; then
        echo -e "${RED}❌ Não foi possível conectar ao node em $RPC_URL${NC}"
        echo -e "${YELLOW}   Certifique-se de que o node está rodando:${NC}"
        echo -e "${YELLOW}   ./scripts/run-node.sh${NC}"
        exit 1
    fi
    
    # Tentar fazer uma requisição HTTP
    if ! curl -s --max-time 5 "$RPC_URL/api/v1/blockchain/info" > /dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  Node está acessível mas API pode não estar respondendo${NC}"
        echo -e "${YELLOW}   Continuando mesmo assim...${NC}"
    else
        echo -e "${GREEN}✅ Conectividade com o node verificada${NC}"
    fi
}

# Função para criar diretórios necessários
create_directories() {
    echo -e "${BLUE}📁 Criando diretórios...${NC}"
    
    mkdir -p "$DATA_PATH"
    
    echo -e "${GREEN}✅ Diretórios criados${NC}"
}

# Função para gerar chave privada se não fornecida
generate_miner_key() {
    if [ -z "$MINER_KEY" ]; then
        echo -e "${BLUE}🔑 Gerando chave privada para o minerador...${NC}"
        
        # Gerar chave aleatória (32 bytes em hex)
        MINER_KEY=$(openssl rand -hex 32 2>/dev/null || echo "miner_key_$(date +%s)")
        
        echo -e "${GREEN}✅ Chave privada gerada: $MINER_KEY${NC}"
        echo -e "${YELLOW}   ⚠️  Guarde esta chave para usar novamente${NC}"
    fi
}

# Função para compilar o binário
build_binary() {
    echo -e "${BLUE}🔨 Compilando binário ordm-miner...${NC}"
    
    if ! go build -o ordm-miner ./cmd/ordm-miner; then
        echo -e "${RED}❌ Erro ao compilar binário${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ Binário compilado com sucesso${NC}"
}

# Função para executar o minerador
run_miner() {
    echo -e "${BLUE}⛏️ Iniciando ORDM Miner...${NC}"
    echo -e "${BLUE}   RPC URL: $RPC_URL${NC}"
    echo -e "${BLUE}   Miner: $MINER_NAME${NC}"
    echo -e "${BLUE}   Threads: $THREADS${NC}"
    echo -e "${BLUE}   Difficulty: $DIFFICULTY${NC}"
    echo -e "${BLUE}   Data Path: $DATA_PATH${NC}"
    echo ""
    
    # Construir comando
    CMD="./ordm-miner"
    CMD="$CMD --rpc $RPC_URL"
    CMD="$CMD --miner-key $MINER_KEY"
    CMD="$CMD --threads $THREADS"
    CMD="$CMD --difficulty $DIFFICULTY"
    CMD="$CMD --data $DATA_PATH"
    CMD="$CMD --name $MINER_NAME"
    
    echo -e "${GREEN}Executando: $CMD${NC}"
    echo ""
    
    # Executar o comando
    exec $CMD
}

# Função para limpeza
cleanup() {
    echo -e "${YELLOW}🛑 Parando minerador...${NC}"
    # O minerador já tem graceful shutdown implementado
}

# Configurar trap para cleanup
trap cleanup SIGINT SIGTERM

# Parse de argumentos
while [[ $# -gt 0 ]]; do
    case $1 in
        --rpc)
            RPC_URL="$2"
            shift 2
            ;;
        --miner-key)
            MINER_KEY="$2"
            shift 2
            ;;
        --threads)
            THREADS="$2"
            shift 2
            ;;
        --difficulty)
            DIFFICULTY="$2"
            shift 2
            ;;
        -d|--data)
            DATA_PATH="$2"
            shift 2
            ;;
        --name)
            MINER_NAME="$2"
            shift 2
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            echo -e "${RED}❌ Opção desconhecida: $1${NC}"
            show_help
            exit 1
            ;;
    esac
done

# Verificar se a chave do minerador foi fornecida
if [ -z "$MINER_KEY" ]; then
    echo -e "${RED}❌ --miner-key é obrigatório${NC}"
    echo ""
    show_help
    exit 1
fi

# Executar verificações e inicialização
echo -e "${BLUE}⛏️ ORDM Miner - Inicializando...${NC}"
echo ""

check_dependencies
check_node_connectivity
create_directories
build_binary

echo ""
echo -e "${GREEN}🎉 Tudo pronto! Iniciando minerador...${NC}"
echo ""

run_miner
