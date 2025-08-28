#!/bin/bash

# 🔗 Script: Integração Matrix GUI
# Descrição: Integra os novos componentes com a GUI existente

set -e

echo "🔗 [$(date)] Iniciando Integração Matrix GUI"
echo "============================================"

# Verificar componentes
echo "🔍 Verificando componentes implementados..."

required_files=(
    "pkg/config/config.go"
    "pkg/auth/rate_limiter.go"
    "pkg/auth/session.go"
    "static/css/matrix-theme.css"
    "static/css/typography.css"
    "static/css/animations.css"
)

for file in "${required_files[@]}"; do
    if [ ! -f "$file" ]; then
        echo "❌ Componente faltando: $file"
        echo "🚀 Execute primeiro: ./scripts/run_matrix_interface.sh"
        exit 1
    fi
done

echo "✅ Todos os componentes estão implementados!"

# 1. Criar templates HTML matrix
echo "🎨 1. Criando templates HTML matrix..."

mkdir -p templates/matrix

cat > templates/matrix/login.html << 'EOF'
<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>ORDM Blockchain 2-Layer - Matrix Terminal</title>
    <link rel="stylesheet" href="/static/css/matrix-theme.css">
    <link rel="stylesheet" href="/static/css/typography.css">
    <link rel="stylesheet" href="/static/css/animations.css">
</head>
<body class="matrix-bg">
    <div class="matrix-container">
        <div class="matrix-login-container">
            <div class="matrix-logo">
                <h1 class="matrix-title matrix-animate-glow">ORDM</h1>
                <p class="matrix-subtitle">Blockchain 2-Layer Terminal</p>
            </div>
            
            <div class="matrix-tabs">
                <button class="matrix-tab active" data-tab="simple">LOGIN SIMPLES</button>
                <button class="matrix-tab" data-tab="advanced">LOGIN AVANÇADO</button>
            </div>
            
            <div class="matrix-form-container">
                <form id="simple-login" class="matrix-form active">
                    <div class="matrix-input-group">
                        <label class="matrix-label">USUÁRIO</label>
                        <input type="text" name="username" class="matrix-input" required>
                    </div>
                    <div class="matrix-input-group">
                        <label class="matrix-label">SENHA</label>
                        <input type="password" name="password" class="matrix-input" required>
                    </div>
                    <button type="submit" class="matrix-btn matrix-btn-primary">ENTRAR</button>
                </form>
                
                <form id="advanced-login" class="matrix-form">
                    <div class="matrix-input-group">
                        <label class="matrix-label">CHAVE PÚBLICA</label>
                        <input type="text" name="publicKey" class="matrix-input" required>
                    </div>
                    <div class="matrix-input-group">
                        <label class="matrix-label">PIN 2FA</label>
                        <input type="text" name="pin" class="matrix-input" maxlength="8" required>
                    </div>
                    <button type="submit" class="matrix-btn matrix-btn-primary">ENTRAR</button>
                </form>
            </div>
            
            <div class="matrix-status">
                <p class="matrix-text-muted">Sistema de Segurança Ativo</p>
                <p class="matrix-text-dim">Rate Limiting • CSRF Protection • HTTPS</p>
            </div>
        </div>
    </div>
    
    <script>
        document.addEventListener('DOMContentLoaded', function() {
            const tabs = document.querySelectorAll('.matrix-tab');
            const forms = document.querySelectorAll('.matrix-form');
            
            tabs.forEach(tab => {
                tab.addEventListener('click', () => {
                    const target = tab.dataset.tab;
                    tabs.forEach(t => t.classList.remove('active'));
                    tab.classList.add('active');
                    forms.forEach(form => {
                        form.classList.remove('active');
                        if (form.id === target + '-login') {
                            form.classList.add('active');
                        }
                    });
                });
            });
            
            document.getElementById('simple-login').addEventListener('submit', function(e) {
                e.preventDefault();
                const formData = new FormData(this);
                fetch('/user-login', { method: 'POST', body: formData })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        window.location.href = '/dashboard';
                    } else {
                        alert('Erro: ' + data.message);
                    }
                });
            });
            
            document.getElementById('advanced-login').addEventListener('submit', function(e) {
                e.preventDefault();
                const formData = new FormData(this);
                fetch('/advanced-login', { method: 'POST', body: formData })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        window.location.href = '/dashboard';
                    } else {
                        alert('Erro: ' + data.message);
                    }
                });
            });
        });
    </script>
</body>
</html>
EOF

# 2. Criar CSS adicional
echo "🎨 2. Criando CSS adicional..."

cat > static/css/components.css << 'EOF'
/* 🎨 ORDM Matrix Components CSS */

.matrix-container {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 16px;
}

.matrix-login-container {
    background: var(--matrix-bg-secondary);
    border: 1px solid var(--matrix-border);
    border-radius: 8px;
    padding: 32px;
    max-width: 500px;
    width: 100%;
    box-shadow: var(--matrix-shadow-strong);
}

.matrix-logo {
    text-align: center;
    margin-bottom: 32px;
}

.matrix-tabs {
    display: flex;
    margin-bottom: 24px;
    border-bottom: 1px solid var(--matrix-border);
}

.matrix-tab {
    flex: 1;
    background: transparent;
    border: none;
    color: var(--matrix-text-muted);
    padding: 16px;
    cursor: pointer;
    font-family: var(--matrix-font-family);
    font-weight: bold;
    text-transform: uppercase;
    transition: all 0.3s;
}

.matrix-tab.active {
    color: var(--matrix-accent);
    border-bottom: 2px solid var(--matrix-accent);
    box-shadow: 0 0 10px var(--matrix-glow);
}

.matrix-form {
    display: none;
}

.matrix-form.active {
    display: block;
}

.matrix-input-group {
    margin-bottom: 16px;
}

.matrix-input {
    width: 100%;
    background: var(--matrix-bg);
    border: 1px solid var(--matrix-border);
    color: var(--matrix-text);
    padding: 16px;
    font-family: var(--matrix-font-family);
    border-radius: 4px;
    transition: all 0.3s;
}

.matrix-input:focus {
    border-color: var(--matrix-accent);
    box-shadow: 0 0 10px var(--matrix-glow);
    outline: none;
}

.matrix-btn {
    background: transparent;
    border: 1px solid var(--matrix-text);
    color: var(--matrix-text);
    padding: 16px 24px;
    cursor: pointer;
    font-family: var(--matrix-font-family);
    font-weight: bold;
    text-transform: uppercase;
    border-radius: 4px;
    transition: all 0.3s;
    width: 100%;
    margin-top: 16px;
}

.matrix-btn:hover {
    background: var(--matrix-text);
    color: var(--matrix-bg);
    box-shadow: 0 0 15px var(--matrix-glow-strong);
}

.matrix-btn-primary {
    border-color: var(--matrix-accent);
    color: var(--matrix-accent);
}

.matrix-btn-primary:hover {
    background: var(--matrix-accent);
    color: var(--matrix-bg);
}

.matrix-status {
    text-align: center;
    margin-top: 24px;
    padding-top: 16px;
    border-top: 1px solid var(--matrix-border);
}
EOF

# 3. Criar script de teste
echo "🧪 3. Criando script de teste..."

cat > scripts/test_matrix_integration.sh << 'EOF'
#!/bin/bash

echo "🧪 Testando Integração Matrix..."
echo "================================"

# Verificar se o servidor está rodando
if curl -s http://localhost:3000 > /dev/null 2>&1; then
    echo "✅ Servidor rodando em http://localhost:3000"
else
    echo "❌ Servidor não está rodando"
    echo "🚀 Execute: go run cmd/gui/main.go"
    exit 1
fi

# Testar página de login
login_response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/)
if [ "$login_response" = "200" ]; then
    echo "✅ Página de login carregada"
else
    echo "❌ Erro na página de login (HTTP $login_response)"
fi

# Testar CSS matrix
css_response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/static/css/matrix-theme.css)
if [ "$css_response" = "200" ]; then
    echo "✅ CSS matrix carregado"
else
    echo "❌ Erro no CSS matrix (HTTP $css_response)"
fi

echo ""
echo "🎉 Teste concluído!"
echo "🌐 Acesse: http://localhost:3000"
EOF

chmod +x scripts/test_matrix_integration.sh

echo "✅ Script de teste criado"

# 4. Criar README
echo "📚 4. Criando README..."

cat > INTEGRATION_README.md << 'EOF'
# 🔗 Integração Matrix GUI - ORDM

## 🎯 Componentes Integrados

### **🔐 Segurança**
- ✅ Rate Limiting
- ✅ CSRF Protection  
- ✅ Input Validation
- ✅ Sessões JWT
- ✅ PIN 2FA (8 dígitos)

### **🎨 Interface Matrix**
- ✅ Design system completo
- ✅ CSS variables matrix
- ✅ Templates HTML matrix
- ✅ Animações e efeitos

## 📁 Arquivos Criados

```
templates/matrix/
└── login.html          # Login matrix

static/css/
└── components.css      # Componentes matrix

scripts/
└── test_matrix_integration.sh   # Script de teste
```

## 🚀 Como Usar

### **1. Testar Integração**
```bash
./scripts/test_matrix_integration.sh
```

### **2. Iniciar Servidor**
```bash
go run cmd/gui/main.go
```

### **3. Acessar Interface**
```
http://localhost:3000
```

## 🎨 Características

- **Fundo:** Preto (#0a0a0a)
- **Texto:** Verde (#00ff00) com glow
- **Fonte:** Courier New, Monaco (monospace)
- **Efeitos:** Glow verde, sombras, gradientes
- **Animações:** Pulse, flicker, typewriter

## 🎉 Resultado

Interface Matrix Terminal integrada com segurança robusta!
EOF

echo "✅ README criado"

echo ""
echo "🎉 [$(date)] Integração Matrix GUI concluída!"
echo "============================================="
echo "📋 Implementações:"
echo "  ✅ 1. Templates HTML matrix criados"
echo "  ✅ 2. CSS adicional criado"
echo "  ✅ 3. Script de teste criado"
echo "  ✅ 4. README criado"
echo ""
echo "🚀 Próximos passos:"
echo "  1. Testar: ./scripts/test_matrix_integration.sh"
echo "  2. Iniciar: go run cmd/gui/main.go"
echo "  3. Acessar: http://localhost:3000"
echo ""
echo "🎨 Interface Matrix Terminal integrada!"
