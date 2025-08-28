#!/bin/bash

# 🚀 Script Principal: Interface Matrix Terminal ORDM
# Descrição: Executa todas as partes da implementação da interface matrix

set -e

echo "🚀 [$(date)] Iniciando Implementação da Interface Matrix Terminal ORDM"
echo "====================================================================="
echo ""

# Verificar pré-requisitos
if ! command -v go &> /dev/null; then
    echo "❌ Go não encontrado. Instale o Go 1.25+ primeiro."
    exit 1
fi

# Função para executar script com verificação
run_script() {
    local script_name=$1
    local description=$2
    
    echo "🔄 Executando: $description"
    echo "   Script: $script_name"
    echo "   ──────────────────────────────────────────────────────────────"
    
    if [ -f "scripts/$script_name" ]; then
        chmod +x "scripts/$script_name"
        if ./scripts/$script_name; then
            echo "✅ $description - CONCLUÍDO"
        else
            echo "❌ $description - FALHOU"
            exit 1
        fi
    else
        echo "❌ Script não encontrado: scripts/$script_name"
        exit 1
    fi
    
    echo ""
}

# Menu interativo
show_menu() {
    echo "🎯 Escolha uma opção:"
    echo ""
    echo "1️⃣  Executar TODAS as partes (recomendado)"
    echo "2️⃣  Executar apenas Segurança Crítica (Parte 1)"
    echo "3️⃣  Executar apenas Arquitetura Limpa (Parte 2)"
    echo "4️⃣  Executar apenas Interface Matrix (Parte 3)"
    echo "5️⃣  Executar partes individuais"
    echo "6️⃣  Verificar status das implementações"
    echo "0️⃣  Sair"
    echo ""
    read -p "Digite sua escolha (0-6): " choice
}

# Executar todas as partes
run_all_parts() {
    echo "🚀 Executando TODAS as partes da implementação..."
    echo "=================================================="
    echo ""
    
    # Parte 1: Segurança Crítica
    echo "🔐 PARTE 1: Segurança Crítica"
    echo "─────────────────────────────"
    run_script "part1a_auth_robust.sh" "1.1 - Autenticação Robusta"
    run_script "part1b_crypto_data.sh" "1.2 - Criptografia de Dados"
    run_script "part1c_attack_protection.sh" "1.3 - Proteção contra Ataques"
    
    # Parte 2: Arquitetura Limpa
    echo "🏗️ PARTE 2: Arquitetura Limpa"
    echo "─────────────────────────────"
    run_script "part2a_clean_architecture.sh" "2.1 - Separação Frontend/Backend"
    # run_script "part2b_thread_safe.sh" "2.2 - Thread-Safe State Management"
    # run_script "part2c_database_layer.sh" "2.3 - Database Layer"
    
    # Parte 3: Interface Matrix
    echo "🎨 PARTE 3: Interface Matrix Terminal"
    echo "─────────────────────────────────────"
    run_script "part3a_matrix_design.sh" "3.1 - Design System Matrix"
    # run_script "part3b_matrix_components.sh" "3.2 - Componentes Matrix"
    # run_script "part3c_matrix_layouts.sh" "3.3 - Layouts Específicos"
    
    echo "🎉 TODAS as partes foram executadas com sucesso!"
}

# Executar apenas Parte 1
run_security_only() {
    echo "🔐 Executando apenas Segurança Crítica..."
    echo "=========================================="
    echo ""
    
    run_script "part1a_auth_robust.sh" "1.1 - Autenticação Robusta"
    run_script "part1b_crypto_data.sh" "1.2 - Criptografia de Dados"
    run_script "part1c_attack_protection.sh" "1.3 - Proteção contra Ataques"
    
    echo "✅ Segurança Crítica implementada!"
}

# Executar apenas Parte 2
run_architecture_only() {
    echo "🏗️ Executando apenas Arquitetura Limpa..."
    echo "=========================================="
    echo ""
    
    run_script "part2a_clean_architecture.sh" "2.1 - Separação Frontend/Backend"
    # run_script "part2b_thread_safe.sh" "2.2 - Thread-Safe State Management"
    # run_script "part2c_database_layer.sh" "2.3 - Database Layer"
    
    echo "✅ Arquitetura Limpa implementada!"
}

# Executar apenas Parte 3
run_interface_only() {
    echo "🎨 Executando apenas Interface Matrix..."
    echo "========================================"
    echo ""
    
    run_script "part3a_matrix_design.sh" "3.1 - Design System Matrix"
    # run_script "part3b_matrix_components.sh" "3.2 - Componentes Matrix"
    # run_script "part3c_matrix_layouts.sh" "3.3 - Layouts Específicos"
    
    echo "✅ Interface Matrix implementada!"
}

# Executar partes individuais
run_individual_parts() {
    echo "🎯 Executar partes individuais"
    echo "=============================="
    echo ""
    echo "1️⃣  part1a_auth_robust.sh - Autenticação Robusta"
    echo "2️⃣  part1b_crypto_data.sh - Criptografia de Dados"
    echo "3️⃣  part1c_attack_protection.sh - Proteção contra Ataques"
    echo "4️⃣  part2a_clean_architecture.sh - Separação Frontend/Backend"
    echo "5️⃣  part3a_matrix_design.sh - Design System Matrix"
    echo "0️⃣  Voltar ao menu principal"
    echo ""
    read -p "Digite sua escolha (0-5): " subchoice
    
    case $subchoice in
        1) run_script "part1a_auth_robust.sh" "1.1 - Autenticação Robusta" ;;
        2) run_script "part1b_crypto_data.sh" "1.2 - Criptografia de Dados" ;;
        3) run_script "part1c_attack_protection.sh" "1.3 - Proteção contra Ataques" ;;
        4) run_script "part2a_clean_architecture.sh" "2.1 - Separação Frontend/Backend" ;;
        5) run_script "part3a_matrix_design.sh" "3.1 - Design System Matrix" ;;
        0) return ;;
        *) echo "❌ Opção inválida" ;;
    esac
}

# Verificar status
check_status() {
    echo "📊 Status das Implementações"
    echo "============================"
    echo ""
    
    # Verificar arquivos criados
    local files=(
        "pkg/config/config.go:Configuração"
        "pkg/auth/rate_limiter.go:Rate Limiter"
        "pkg/auth/session.go:Sessões JWT"
        "pkg/crypto/wallet_encryption.go:Criptografia"
        "pkg/auth/password.go:Hash de Senhas"
        "pkg/auth/pin_generator.go:PIN 2FA"
        "pkg/middleware/csrf.go:CSRF Protection"
        "pkg/validation/input.go:Input Validation"
        "pkg/server/https.go:HTTPS"
        "pkg/api/rest.go:API REST"
        "pkg/middleware/chain.go:Middleware Chain"
        "pkg/services/mining_service.go:Mining Service"
        "pkg/services/wallet_service.go:Wallet Service"
        "static/css/matrix-theme.css:Matrix Theme"
        "static/css/typography.css:Typography"
        "static/css/animations.css:Animations"
    )
    
    local total=0
    local created=0
    
    for file_info in "${files[@]}"; do
        IFS=':' read -r file description <<< "$file_info"
        total=$((total + 1))
        
        if [ -f "$file" ]; then
            echo "✅ $description"
            created=$((created + 1))
        else
            echo "❌ $description"
        fi
    done
    
    echo ""
    echo "📈 Progresso: $created/$total arquivos criados"
    echo "📊 Percentual: $((created * 100 / total))%"
    echo ""
    
    if [ $created -eq $total ]; then
        echo "🎉 TODAS as implementações foram concluídas!"
    else
        echo "⚠️  Algumas implementações ainda estão pendentes."
    fi
}

# Loop principal
while true; do
    show_menu
    
    case $choice in
        1) run_all_parts ;;
        2) run_security_only ;;
        3) run_architecture_only ;;
        4) run_interface_only ;;
        5) run_individual_parts ;;
        6) check_status ;;
        0) 
            echo "👋 Saindo..."
            exit 0
            ;;
        *) 
            echo "❌ Opção inválida. Tente novamente."
            echo ""
            ;;
    esac
    
    if [ "$choice" != "6" ]; then
        echo ""
        echo "🔄 Deseja continuar?"
        read -p "Pressione Enter para voltar ao menu ou 'q' para sair: " continue_choice
        if [ "$continue_choice" = "q" ]; then
            echo "👋 Saindo..."
            exit 0
        fi
    fi
    
    echo ""
done

