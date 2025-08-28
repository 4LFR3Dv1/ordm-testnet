#!/bin/bash

# 🚀 Script para rodar ORDM Node da Testnet
# Uso: ./scripts/run-node.sh [opções]

set -e

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configurações padrão
NETWORK="testnet"
PORT="8080"
P2P_PORT="3000"
RPC_PORT="8081"
DATA_PATH="./data/testnet"
CONFIG_FILE="config/testnet.json"
GENESIS_FILE="genesis/testnet.json"
MAX_PEERS=50
BLOCK_TIME="0"
DIFFICULTY=4

# Função para exibir ajuda
show_help() {
    echo -e "${BLUE}🚀 ORDM Node/Miner - Script de Execução${NC}"
    echo ""
    echo "Uso: $0 [opções]"
    echo ""
    echo "Opções:"
    echo "  -n, --network NETWORK    Rede (testnet/mainnet) [padrão: testnet]"
    echo "  -p, --port PORT          Porta HTTP [padrão: 8080]"
    echo "  --p2p-port PORT          Porta P2P [padrão: 3000]"
    echo "  --rpc-port PORT          Porta RPC [padrão: 8081]"
    echo "  -d, --data PATH          Caminho para dados [padrão: ./data/testnet]"
    echo "  -c, --config FILE        Arquivo de configuração"
    echo "  -g, --genesis FILE       Arquivo do bloco genesis"
    echo "  --max-peers NUM          Máximo de peers [padrão: 50]"
    echo "  --block-time DURATION    Tempo entre blocos (0 = sem mineração automática)"
    echo "  --difficulty NUM         Dificuldade de mineração [padrão: 4]"
    echo "  --mode MODE              Modo de operação (node/miner/both) [padrão: node]"
    echo "  --mining                 Habilitar mineração"
    echo "  --miner-key KEY          Chave privada do minerador"
    echo "  --miner-threads NUM      Número de threads de mineração [padrão: 1]"
    echo "  --miner-name NAME        Nome do minerador [padrão: ordm-node]"
    echo "  -h, --help               Exibir esta ajuda"
    echo ""
    echo "Exemplos:"
    echo "  $0                                    # Rodar node da testnet com configurações padrão"
    echo "  $0 --mode miner --miner-key abc123   # Rodar apenas como minerador"
    echo "  $0 --mode both --miner-threads 4     # Rodar node + minerador com 4 threads"
    echo "  $0 --mining --miner-name my-miner    # Rodar node com mineração habilitada"
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
    if [ ! -f "./cmd/ordmd/main.go" ]; then
        echo -e "${RED}❌ Binário ordmd não encontrado. Execute 'go build ./cmd/ordmd' primeiro.${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ Dependências verificadas${NC}"
}

# Função para criar diretórios necessários
create_directories() {
    echo -e "${BLUE}📁 Criando diretórios...${NC}"
    
    mkdir -p "$DATA_PATH"
    mkdir -p "$(dirname "$CONFIG_FILE")"
    mkdir -p "$(dirname "$GENESIS_FILE")"
    
    echo -e "${GREEN}✅ Diretórios criados${NC}"
}

# Função para verificar arquivos de configuração
check_config_files() {
    echo -e "${BLUE}📋 Verificando arquivos de configuração...${NC}"
    
    if [ ! -f "$CONFIG_FILE" ]; then
        echo -e "${YELLOW}⚠️  Arquivo de configuração não encontrado: $CONFIG_FILE${NC}"
        echo -e "${YELLOW}   Serão usadas configurações padrão${NC}"
    else
        echo -e "${GREEN}✅ Arquivo de configuração encontrado${NC}"
    fi
    
    if [ ! -f "$GENESIS_FILE" ]; then
        echo -e "${YELLOW}⚠️  Arquivo genesis não encontrado: $GENESIS_FILE${NC}"
        echo -e "${YELLOW}   Será criado bloco genesis padrão${NC}"
    else
        echo -e "${GREEN}✅ Arquivo genesis encontrado${NC}"
    fi
}

# Função para compilar o binário
build_binary() {
    echo -e "${BLUE}🔨 Compilando binário ordmd...${NC}"
    
    if ! go build -o ordmd ./cmd/ordmd; then
        echo -e "${RED}❌ Erro ao compilar binário${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ Binário compilado com sucesso${NC}"
}

# Função para executar o node
run_node() {
    echo -e "${BLUE}🚀 Iniciando ORDM Node/Miner...${NC}"
    echo -e "${BLUE}   Network: $NETWORK${NC}"
    echo -e "${BLUE}   Port: $PORT${NC}"
    echo -e "${BLUE}   P2P Port: $P2P_PORT${NC}"
    echo -e "${BLUE}   RPC Port: $RPC_PORT${NC}"
    echo -e "${BLUE}   Data Path: $DATA_PATH${NC}"
    echo -e "${BLUE}   MachineID: Será gerado automaticamente na primeira execução${NC}"
    echo ""
    
    # Construir comando
    CMD="./ordmd"
    CMD="$CMD --network $NETWORK"
    CMD="$CMD --port $PORT"
    CMD="$CMD --p2p-port $P2P_PORT"
    CMD="$CMD --rpc-port $RPC_PORT"
    CMD="$CMD --data $DATA_PATH"
    CMD="$CMD --max-peers $MAX_PEERS"
    CMD="$CMD --difficulty $DIFFICULTY"
    
    if [ "$CONFIG_FILE" != "" ]; then
        CMD="$CMD --config $CONFIG_FILE"
    fi
    
    if [ "$GENESIS_FILE" != "" ]; then
        CMD="$CMD --genesis $GENESIS_FILE"
    fi
    
    if [ "$BLOCK_TIME" != "0" ]; then
        CMD="$CMD --block-time $BLOCK_TIME"
    fi
    
    echo -e "${GREEN}Executando: $CMD${NC}"
    echo ""
    
    # Executar o comando
    exec $CMD
}

# Função para limpeza
cleanup() {
    echo -e "${YELLOW}🛑 Parando node...${NC}"
    # O node já tem graceful shutdown implementado
}

# Configurar trap para cleanup
trap cleanup SIGINT SIGTERM

# Parse de argumentos
while [[ $# -gt 0 ]]; do
    case $1 in
        -n|--network)
            NETWORK="$2"
            shift 2
            ;;
        -p|--port)
            PORT="$2"
            shift 2
            ;;
        --p2p-port)
            P2P_PORT="$2"
            shift 2
            ;;
        --rpc-port)
            RPC_PORT="$2"
            shift 2
            ;;
        -d|--data)
            DATA_PATH="$2"
            shift 2
            ;;
        -c|--config)
            CONFIG_FILE="$2"
            shift 2
            ;;
        -g|--genesis)
            GENESIS_FILE="$2"
            shift 2
            ;;
        --max-peers)
            MAX_PEERS="$2"
            shift 2
            ;;
        --block-time)
            BLOCK_TIME="$2"
            shift 2
            ;;
        --difficulty)
            DIFFICULTY="$2"
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

# Executar verificações e inicialização
echo -e "${BLUE}🚀 ORDM Node - Inicializando...${NC}"
echo ""

check_dependencies
create_directories
check_config_files
build_binary

echo ""
echo -e "${GREEN}🎉 Tudo pronto! Iniciando node...${NC}"
echo ""

run_node
