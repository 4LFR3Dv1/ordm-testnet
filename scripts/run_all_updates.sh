#!/bin/bash

# 🚀 Script Principal - Execução Completa das Atualizações ORDM
# Baseado no PLANO_ATUALIZACOES.md

set -e

# Cores para output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
}

info() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] INFO: $1${NC}"
}

# Função para verificar se comando existe
check_command() {
    if ! command -v $1 &> /dev/null; then
        error "$1 não está instalado. Por favor, instale primeiro."
        exit 1
    fi
}

# Função para verificar pré-requisitos
check_prerequisites() {
    log "🔍 Verificando pré-requisitos..."
    
    check_command "go"
    check_command "git"
    check_command "bash"
    
    # Verificar versão do Go
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    REQUIRED_VERSION="1.25"
    
    if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
        error "Go $REQUIRED_VERSION+ é necessário. Versão atual: $GO_VERSION"
        exit 1
    fi
    
    log "✅ Pré-requisitos verificados"
}

# Função para executar script com tratamento de erro
run_script() {
    local script_name=$1
    local script_path="scripts/$script_name"
    
    if [ ! -f "$script_path" ]; then
        error "Script não encontrado: $script_path"
        return 1
    fi
    
    log "🔄 Executando $script_name..."
    
    if bash "$script_path"; then
        log "✅ $script_name concluído com sucesso"
        return 0
    else
        error "❌ $script_name falhou"
        return 1
    fi
}

# Função para mostrar menu
show_menu() {
    echo -e "${BLUE}"
    echo "=========================================="
    echo "🚀 ORDM Blockchain 2-Layer - Atualizações"
    echo "=========================================="
    echo -e "${NC}"
    echo "Escolha uma opção:"
    echo ""
    echo "1.  PARTE 1: Consolidação Arquitetural"
    echo "2.  PARTE 2A: Storage Offline"
    echo "3.  PARTE 2B: Storage Online"
    echo "4.  PARTE 2C: Protocolo de Sincronização"
    echo "5.  PARTE 3: Segurança"
    echo "6.  PARTE 4A: Auditoria de Dependências"
    echo "7.  PARTE 5A: Testes Unitários"
    echo ""
    echo "8.  Executar TODAS as partes (sequencial)"
    echo "9.  Executar TODAS as partes (paralelo)"
    echo "10. Verificar status das atualizações"
    echo "11. Sair"
    echo ""
}

# Função para executar parte específica
run_part() {
    case $1 in
        1)
            run_script "part1_consolidate_architecture.sh"
            ;;
        2)
            run_script "part2a_offline_storage.sh"
            ;;
        3)
            run_script "part2b_online_storage.sh"
            ;;
        4)
            run_script "part2c_sync_protocol.sh"
            ;;
        5)
            run_script "part3_security.sh"
            ;;
        6)
            run_script "part4a_dependencies.sh"
            ;;
        7)
            run_script "part5a_unit_tests.sh"
            ;;
        *)
            error "Opção inválida: $1"
            return 1
            ;;
    esac
}

# Função para executar todas as partes sequencialmente
run_all_sequential() {
    log "🚀 Executando TODAS as partes sequencialmente..."
    
    local scripts=(
        "part1_consolidate_architecture.sh"
        "part2a_offline_storage.sh"
        "part2b_online_storage.sh"
        "part2c_sync_protocol.sh"
        "part3_security.sh"
        "part4a_dependencies.sh"
        "part5a_unit_tests.sh"
    )
    
    local failed_scripts=()
    
    for script in "${scripts[@]}"; do
        if run_script "$script"; then
            log "✅ $script executado com sucesso"
        else
            error "❌ $script falhou"
            failed_scripts+=("$script")
        fi
    done
    
    if [ ${#failed_scripts[@]} -eq 0 ]; then
        log "🎉 TODAS as partes foram executadas com sucesso!"
    else
        error "❌ Alguns scripts falharam:"
        for script in "${failed_scripts[@]}"; do
            error "   - $script"
        done
        return 1
    fi
}

# Função para executar todas as partes em paralelo
run_all_parallel() {
    warn "⚠️  Execução paralela pode causar conflitos. Use com cuidado!"
    read -p "Continuar? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log "Execução paralela cancelada"
        return 0
    fi
    
    log "🚀 Executando TODAS as partes em paralelo..."
    
    local scripts=(
        "part1_consolidate_architecture.sh"
        "part2a_offline_storage.sh"
        "part2b_online_storage.sh"
        "part2c_sync_protocol.sh"
        "part3_security.sh"
        "part4a_dependencies.sh"
        "part5a_unit_tests.sh"
    )
    
    local pids=()
    
    for script in "${scripts[@]}"; do
        log "🔄 Iniciando $script em background..."
        bash "scripts/$script" &
        pids+=($!)
    done
    
    log "⏳ Aguardando conclusão de todos os scripts..."
    
    local failed_scripts=()
    for i in "${!scripts[@]}"; do
        if wait ${pids[$i]}; then
            log "✅ ${scripts[$i]} concluído com sucesso"
        else
            error "❌ ${scripts[$i]} falhou"
            failed_scripts+=("${scripts[$i]}")
        fi
    done
    
    if [ ${#failed_scripts[@]} -eq 0 ]; then
        log "🎉 TODAS as partes foram executadas com sucesso!"
    else
        error "❌ Alguns scripts falharam:"
        for script in "${failed_scripts[@]}"; do
            error "   - $script"
        done
        return 1
    fi
}

# Função para verificar status
check_status() {
    log "📊 Verificando status das atualizações..."
    
    local files=(
        "DECISIONS.md"
        "DEPENDENCIES.md"
        "FLOW_DIAGRAM.md"
        "API_CONTRACTS.md"
        "pkg/storage/offline_storage.go"
        "pkg/storage/render_storage.go"
        "pkg/sync/protocol.go"
        "pkg/auth/rate_limiter.go"
        "pkg/auth/pin_generator.go"
        "pkg/crypto/keystore.go"
        "pkg/logger/secure_logger.go"
        "pkg/blockchain/real_block_test.go"
        "pkg/wallet/secure_wallet_test.go"
        "pkg/auth/user_manager_test.go"
        "scripts/run_tests.sh"
    )
    
    local existing=0
    local missing=0
    
    for file in "${files[@]}"; do
        if [ -f "$file" ]; then
            echo -e "${GREEN}✅ $file${NC}"
            ((existing++))
        else
            echo -e "${RED}❌ $file${NC}"
            ((missing++))
        fi
    done
    
    echo ""
    echo "📈 Resumo:"
    echo "   - Arquivos existentes: $existing"
    echo "   - Arquivos faltando: $missing"
    echo "   - Total: ${#files[@]}"
    
    if [ $missing -eq 0 ]; then
        log "🎉 Todas as atualizações foram aplicadas!"
    else
        warn "⚠️  Algumas atualizações ainda precisam ser aplicadas"
    fi
}

# Função principal
main() {
    log "🚀 Iniciando sistema de atualizações ORDM"
    
    # Verificar pré-requisitos
    check_prerequisites
    
    # Loop principal
    while true; do
        show_menu
        read -p "Digite sua opção (1-11): " choice
        
        case $choice in
            1|2|3|4|5|6|7)
                run_part $choice
                ;;
            8)
                run_all_sequential
                ;;
            9)
                run_all_parallel
                ;;
            10)
                check_status
                ;;
            11)
                log "👋 Saindo..."
                exit 0
                ;;
            *)
                error "Opção inválida: $choice"
                ;;
        esac
        
        echo ""
        read -p "Pressione Enter para continuar..."
        echo ""
    done
}

# Executar função principal
main "$@"

