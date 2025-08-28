#!/bin/bash

# 🎨 Script: Parte 3.1 - Design System Matrix
# Descrição: Implementa CSS variables, typography e animation system estilo matrix

set -e

echo "🎨 [$(date)] Iniciando Parte 3.1: Design System Matrix"
echo "======================================================"

# 3.1.1 - CSS Variables Matrix
echo "🎨 3.1.1 - Implementando CSS Variables Matrix..."

mkdir -p static/css
cat > static/css/matrix-theme.css << 'EOF'
/* 🎨 ORDM Blockchain 2-Layer - Matrix Theme */
/* Design System para Interface Terminal Matrix */

:root {
  /* 🎨 Cores Principais Matrix */
  --matrix-bg: #0a0a0a;
  --matrix-bg-secondary: #111111;
  --matrix-bg-tertiary: #1a1a1a;
  
  /* 🎨 Cores de Texto */
  --matrix-text: #00ff00;
  --matrix-text-secondary: #00cc00;
  --matrix-text-muted: #008800;
  --matrix-text-dim: #004400;
  
  /* 🎨 Cores de Destaque */
  --matrix-accent: #00ff00;
  --matrix-accent-secondary: #00cc00;
  --matrix-accent-tertiary: #009900;
  
  /* 🎨 Cores de Status */
  --matrix-success: #00ff00;
  --matrix-warning: #ffff00;
  --matrix-error: #ff0000;
  --matrix-info: #00ffff;
  
  /* 🎨 Cores de Interface */
  --matrix-border: #333333;
  --matrix-border-light: #555555;
  --matrix-border-dark: #222222;
  
  /* 🎨 Efeitos de Glow */
  --matrix-glow: rgba(0, 255, 0, 0.3);
  --matrix-glow-strong: rgba(0, 255, 0, 0.5);
  --matrix-glow-weak: rgba(0, 255, 0, 0.1);
  
  /* 🎨 Sombras */
  --matrix-shadow: 0 0 10px rgba(0, 255, 0, 0.2);
  --matrix-shadow-strong: 0 0 20px rgba(0, 255, 0, 0.4);
  --matrix-shadow-inset: inset 0 0 10px rgba(0, 255, 0, 0.1);
  
  /* 🎨 Gradientes */
  --matrix-gradient-primary: linear-gradient(135deg, #0a0a0a 0%, #1a1a1a 100%);
  --matrix-gradient-accent: linear-gradient(135deg, #00ff00 0%, #00cc00 100%);
  --matrix-gradient-glow: linear-gradient(135deg, rgba(0, 255, 0, 0.1) 0%, rgba(0, 255, 0, 0.05) 100%);
  
  /* 🎨 Tipografia */
  --matrix-font-family: 'Courier New', 'Monaco', 'Consolas', 'Lucida Console', monospace;
  --matrix-font-size-small: 12px;
  --matrix-font-size-base: 14px;
  --matrix-font-size-large: 16px;
  --matrix-font-size-xl: 18px;
  --matrix-font-size-xxl: 24px;
  --matrix-font-size-title: 32px;
  
  /* 🎨 Espaçamentos */
  --matrix-spacing-xs: 4px;
  --matrix-spacing-sm: 8px;
  --matrix-spacing-md: 16px;
  --matrix-spacing-lg: 24px;
  --matrix-spacing-xl: 32px;
  --matrix-spacing-xxl: 48px;
  
  /* 🎨 Bordas */
  --matrix-border-radius-sm: 2px;
  --matrix-border-radius-md: 4px;
  --matrix-border-radius-lg: 8px;
  --matrix-border-radius-xl: 12px;
  
  /* 🎨 Transições */
  --matrix-transition-fast: 0.15s ease-in-out;
  --matrix-transition-normal: 0.3s ease-in-out;
  --matrix-transition-slow: 0.5s ease-in-out;
  
  /* 🎨 Z-Index */
  --matrix-z-dropdown: 1000;
  --matrix-z-modal: 2000;
  --matrix-z-tooltip: 3000;
  --matrix-z-notification: 4000;
}

/* 🎨 Reset e Base */
* {
  box-sizing: border-box;
}

body {
  background: var(--matrix-bg);
  color: var(--matrix-text);
  font-family: var(--matrix-font-family);
  font-size: var(--matrix-font-size-base);
  line-height: 1.6;
  margin: 0;
  padding: 0;
  overflow-x: hidden;
}

/* 🎨 Scrollbar Matrix */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: var(--matrix-bg-secondary);
  border-radius: var(--matrix-border-radius-sm);
}

::-webkit-scrollbar-thumb {
  background: var(--matrix-border);
  border-radius: var(--matrix-border-radius-sm);
  box-shadow: var(--matrix-shadow);
}

::-webkit-scrollbar-thumb:hover {
  background: var(--matrix-border-light);
  box-shadow: var(--matrix-shadow-strong);
}

/* 🎨 Seleção de Texto */
::selection {
  background: var(--matrix-accent);
  color: var(--matrix-bg);
}

::-moz-selection {
  background: var(--matrix-accent);
  color: var(--matrix-bg);
}
EOF

# 3.1.2 - Typography Matrix
echo "📝 3.1.2 - Implementando Typography Matrix..."

cat > static/css/typography.css << 'EOF'
/* 📝 ORDM Matrix Typography System */

/* 🎨 Classes de Fonte */
.matrix-font {
  font-family: var(--matrix-font-family);
  font-weight: bold;
  text-shadow: 0 0 5px var(--matrix-glow);
  letter-spacing: 0.5px;
}

.matrix-font-light {
  font-family: var(--matrix-font-family);
  font-weight: normal;
  text-shadow: 0 0 3px var(--matrix-glow-weak);
  letter-spacing: 0.3px;
}

.matrix-font-heavy {
  font-family: var(--matrix-font-family);
  font-weight: 900;
  text-shadow: 0 0 8px var(--matrix-glow-strong);
  letter-spacing: 1px;
}

/* 🎨 Tamanhos de Fonte */
.matrix-text-xs {
  font-size: var(--matrix-font-size-small);
  line-height: 1.4;
}

.matrix-text-sm {
  font-size: var(--matrix-font-size-base);
  line-height: 1.5;
}

.matrix-text-md {
  font-size: var(--matrix-font-size-large);
  line-height: 1.6;
}

.matrix-text-lg {
  font-size: var(--matrix-font-size-xl);
  line-height: 1.5;
}

.matrix-text-xl {
  font-size: var(--matrix-font-size-xxl);
  line-height: 1.4;
}

.matrix-text-title {
  font-size: var(--matrix-font-size-title);
  line-height: 1.3;
  font-weight: 900;
}

/* 🎨 Cores de Texto */
.matrix-text-primary {
  color: var(--matrix-text);
}

.matrix-text-secondary {
  color: var(--matrix-text-secondary);
}

.matrix-text-muted {
  color: var(--matrix-text-muted);
}

.matrix-text-dim {
  color: var(--matrix-text-dim);
}

/* 🎨 Estados de Texto */
.matrix-text-success {
  color: var(--matrix-success);
  text-shadow: 0 0 5px var(--matrix-success);
}

.matrix-text-warning {
  color: var(--matrix-warning);
  text-shadow: 0 0 5px var(--matrix-warning);
}

.matrix-text-error {
  color: var(--matrix-error);
  text-shadow: 0 0 5px var(--matrix-error);
}

.matrix-text-info {
  color: var(--matrix-info);
  text-shadow: 0 0 5px var(--matrix-info);
}

/* 🎨 Efeitos de Texto */
.matrix-text-glow {
  text-shadow: 0 0 10px var(--matrix-glow-strong);
}

.matrix-text-pulse {
  animation: matrix-text-pulse 2s ease-in-out infinite;
}

.matrix-text-flicker {
  animation: matrix-text-flicker 3s ease-in-out infinite;
}

.matrix-text-typewriter {
  overflow: hidden;
  border-right: 2px solid var(--matrix-accent);
  white-space: nowrap;
  animation: matrix-typewriter 3s steps(40, end), matrix-blink-caret 0.75s step-end infinite;
}

/* 🎨 Títulos */
.matrix-title {
  font-size: var(--matrix-font-size-title);
  font-weight: 900;
  color: var(--matrix-accent);
  text-shadow: 0 0 15px var(--matrix-glow-strong);
  letter-spacing: 2px;
  text-transform: uppercase;
  margin: 0;
  padding: var(--matrix-spacing-md) 0;
}

.matrix-subtitle {
  font-size: var(--matrix-font-size-xl);
  font-weight: bold;
  color: var(--matrix-text-secondary);
  text-shadow: 0 0 8px var(--matrix-glow);
  letter-spacing: 1px;
  margin: 0;
  padding: var(--matrix-spacing-sm) 0;
}

.matrix-heading {
  font-size: var(--matrix-font-size-xxl);
  font-weight: bold;
  color: var(--matrix-text);
  text-shadow: 0 0 10px var(--matrix-glow);
  letter-spacing: 1px;
  margin: 0;
  padding: var(--matrix-spacing-sm) 0;
}

/* 🎨 Parágrafos */
.matrix-paragraph {
  font-size: var(--matrix-font-size-base);
  line-height: 1.6;
  color: var(--matrix-text-secondary);
  margin: var(--matrix-spacing-sm) 0;
  text-shadow: 0 0 3px var(--matrix-glow-weak);
}

.matrix-paragraph-small {
  font-size: var(--matrix-font-size-small);
  line-height: 1.5;
  color: var(--matrix-text-muted);
  margin: var(--matrix-spacing-xs) 0;
}

/* 🎨 Labels e Captions */
.matrix-label {
  font-size: var(--matrix-font-size-small);
  font-weight: bold;
  color: var(--matrix-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: var(--matrix-spacing-xs);
}

.matrix-caption {
  font-size: var(--matrix-font-size-small);
  color: var(--matrix-text-muted);
  font-style: italic;
  margin-top: var(--matrix-spacing-xs);
}

/* 🎨 Código e Terminal */
.matrix-code {
  font-family: var(--matrix-font-family);
  font-size: var(--matrix-font-size-small);
  background: var(--matrix-bg-secondary);
  color: var(--matrix-text);
  padding: var(--matrix-spacing-sm);
  border: 1px solid var(--matrix-border);
  border-radius: var(--matrix-border-radius-sm);
  box-shadow: var(--matrix-shadow-inset);
}

.matrix-terminal-text {
  font-family: var(--matrix-font-family);
  font-size: var(--matrix-font-size-base);
  color: var(--matrix-text);
  text-shadow: 0 0 5px var(--matrix-glow);
  line-height: 1.4;
  white-space: pre-wrap;
}

/* 🎨 Links */
.matrix-link {
  color: var(--matrix-accent);
  text-decoration: none;
  text-shadow: 0 0 3px var(--matrix-glow);
  transition: all var(--matrix-transition-fast);
}

.matrix-link:hover {
  color: var(--matrix-accent-secondary);
  text-shadow: 0 0 8px var(--matrix-glow-strong);
  text-decoration: underline;
}

.matrix-link:active {
  color: var(--matrix-accent-tertiary);
}

/* 🎨 Listas */
.matrix-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.matrix-list-item {
  padding: var(--matrix-spacing-xs) 0;
  color: var(--matrix-text-secondary);
  position: relative;
  padding-left: var(--matrix-spacing-md);
}

.matrix-list-item::before {
  content: ">";
  color: var(--matrix-accent);
  position: absolute;
  left: 0;
  text-shadow: 0 0 5px var(--matrix-glow);
}

.matrix-list-item-success::before {
  content: "✓";
  color: var(--matrix-success);
}

.matrix-list-item-error::before {
  content: "✗";
  color: var(--matrix-error);
}

.matrix-list-item-warning::before {
  content: "⚠";
  color: var(--matrix-warning);
}
EOF

# 3.1.3 - Animation System
echo "🎬 3.1.3 - Implementando Animation System..."

cat > static/css/animations.css << 'EOF'
/* 🎬 ORDM Matrix Animation System */

/* 🎨 Keyframes Principais */
@keyframes matrix-glow {
  0% { 
    box-shadow: 0 0 5px var(--matrix-glow);
    text-shadow: 0 0 5px var(--matrix-glow);
  }
  50% { 
    box-shadow: 0 0 20px var(--matrix-glow-strong);
    text-shadow: 0 0 15px var(--matrix-glow-strong);
  }
  100% { 
    box-shadow: 0 0 5px var(--matrix-glow);
    text-shadow: 0 0 5px var(--matrix-glow);
  }
}

@keyframes matrix-text-pulse {
  0% { 
    opacity: 0.8;
    text-shadow: 0 0 5px var(--matrix-glow);
  }
  50% { 
    opacity: 1;
    text-shadow: 0 0 15px var(--matrix-glow-strong);
  }
  100% { 
    opacity: 0.8;
    text-shadow: 0 0 5px var(--matrix-glow);
  }
}

@keyframes matrix-text-flicker {
  0%, 100% { 
    opacity: 1;
    text-shadow: 0 0 5px var(--matrix-glow);
  }
  10% { 
    opacity: 0.8;
    text-shadow: 0 0 3px var(--matrix-glow-weak);
  }
  20% { 
    opacity: 1;
    text-shadow: 0 0 8px var(--matrix-glow);
  }
  30% { 
    opacity: 0.9;
    text-shadow: 0 0 5px var(--matrix-glow);
  }
  40% { 
    opacity: 1;
    text-shadow: 0 0 12px var(--matrix-glow-strong);
  }
  50% { 
    opacity: 0.7;
    text-shadow: 0 0 2px var(--matrix-glow-weak);
  }
  60% { 
    opacity: 1;
    text-shadow: 0 0 10px var(--matrix-glow);
  }
  70% { 
    opacity: 0.9;
    text-shadow: 0 0 6px var(--matrix-glow);
  }
  80% { 
    opacity: 1;
    text-shadow: 0 0 14px var(--matrix-glow-strong);
  }
  90% { 
    opacity: 0.8;
    text-shadow: 0 0 4px var(--matrix-glow);
  }
}

@keyframes matrix-typewriter {
  from { width: 0; }
  to { width: 100%; }
}

@keyframes matrix-blink-caret {
  from, to { border-color: transparent; }
  50% { border-color: var(--matrix-accent); }
}

@keyframes matrix-fade-in {
  from { 
    opacity: 0;
    transform: translateY(20px);
  }
  to { 
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes matrix-fade-out {
  from { 
    opacity: 1;
    transform: translateY(0);
  }
  to { 
    opacity: 0;
    transform: translateY(-20px);
  }
}

@keyframes matrix-slide-in-left {
  from { 
    opacity: 0;
    transform: translateX(-50px);
  }
  to { 
    opacity: 1;
    transform: translateX(0);
  }
}

@keyframes matrix-slide-in-right {
  from { 
    opacity: 0;
    transform: translateX(50px);
  }
  to { 
    opacity: 1;
    transform: translateX(0);
  }
}

@keyframes matrix-scale-in {
  from { 
    opacity: 0;
    transform: scale(0.8);
  }
  to { 
    opacity: 1;
    transform: scale(1);
  }
}

@keyframes matrix-rotate-in {
  from { 
    opacity: 0;
    transform: rotate(-180deg) scale(0.5);
  }
  to { 
    opacity: 1;
    transform: rotate(0deg) scale(1);
  }
}

@keyframes matrix-bounce {
  0%, 20%, 53%, 80%, 100% {
    transform: translateY(0);
  }
  40%, 43% {
    transform: translateY(-10px);
  }
  70% {
    transform: translateY(-5px);
  }
  90% {
    transform: translateY(-2px);
  }
}

@keyframes matrix-shake {
  0%, 100% {
    transform: translateX(0);
  }
  10%, 30%, 50%, 70%, 90% {
    transform: translateX(-5px);
  }
  20%, 40%, 60%, 80% {
    transform: translateX(5px);
  }
}

@keyframes matrix-spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

@keyframes matrix-pulse {
  0% {
    transform: scale(1);
    box-shadow: 0 0 5px var(--matrix-glow);
  }
  50% {
    transform: scale(1.05);
    box-shadow: 0 0 20px var(--matrix-glow-strong);
  }
  100% {
    transform: scale(1);
    box-shadow: 0 0 5px var(--matrix-glow);
  }
}

/* 🎨 Classes de Animação */
.matrix-animate-glow {
  animation: matrix-glow 2s ease-in-out infinite;
}

.matrix-animate-pulse {
  animation: matrix-pulse 2s ease-in-out infinite;
}

.matrix-animate-bounce {
  animation: matrix-bounce 1s ease-in-out;
}

.matrix-animate-shake {
  animation: matrix-shake 0.5s ease-in-out;
}

.matrix-animate-spin {
  animation: matrix-spin 1s linear infinite;
}

.matrix-animate-fade-in {
  animation: matrix-fade-in 0.5s ease-out;
}

.matrix-animate-fade-out {
  animation: matrix-fade-out 0.5s ease-in;
}

.matrix-animate-slide-in-left {
  animation: matrix-slide-in-left 0.5s ease-out;
}

.matrix-animate-slide-in-right {
  animation: matrix-slide-in-right 0.5s ease-out;
}

.matrix-animate-scale-in {
  animation: matrix-scale-in 0.3s ease-out;
}

.matrix-animate-rotate-in {
  animation: matrix-rotate-in 0.6s ease-out;
}

/* 🎨 Delays de Animação */
.matrix-delay-1 { animation-delay: 0.1s; }
.matrix-delay-2 { animation-delay: 0.2s; }
.matrix-delay-3 { animation-delay: 0.3s; }
.matrix-delay-4 { animation-delay: 0.4s; }
.matrix-delay-5 { animation-delay: 0.5s; }

/* 🎨 Durações de Animação */
.matrix-duration-fast { animation-duration: 0.2s; }
.matrix-duration-normal { animation-duration: 0.5s; }
.matrix-duration-slow { animation-duration: 1s; }
.matrix-duration-slower { animation-duration: 2s; }

/* 🎨 Estados de Hover */
.matrix-hover-glow:hover {
  animation: matrix-glow 1s ease-in-out infinite;
}

.matrix-hover-pulse:hover {
  animation: matrix-pulse 1s ease-in-out infinite;
}

.matrix-hover-scale:hover {
  transform: scale(1.05);
  transition: transform var(--matrix-transition-fast);
}

.matrix-hover-lift:hover {
  transform: translateY(-2px);
  box-shadow: var(--matrix-shadow-strong);
  transition: all var(--matrix-transition-fast);
}

/* 🎨 Estados de Loading */
.matrix-loading {
  position: relative;
  overflow: hidden;
}

.matrix-loading::after {
  content: "";
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, var(--matrix-glow), transparent);
  animation: matrix-loading-shimmer 1.5s infinite;
}

@keyframes matrix-loading-shimmer {
  0% { left: -100%; }
  100% { left: 100%; }
}

/* 🎨 Estados de Focus */
.matrix-focus-glow:focus {
  outline: none;
  box-shadow: 0 0 15px var(--matrix-glow-strong);
  animation: matrix-glow 1s ease-in-out infinite;
}

/* 🎨 Estados de Active */
.matrix-active-press:active {
  transform: scale(0.95);
  transition: transform var(--matrix-transition-fast);
}

/* 🎨 Estados de Disabled */
.matrix-disabled {
  opacity: 0.5;
  pointer-events: none;
  filter: grayscale(100%);
}

/* 🎨 Transições Suaves */
.matrix-transition-all {
  transition: all var(--matrix-transition-normal);
}

.matrix-transition-color {
  transition: color var(--matrix-transition-fast), text-shadow var(--matrix-transition-fast);
}

.matrix-transition-transform {
  transition: transform var(--matrix-transition-fast);
}

.matrix-transition-opacity {
  transition: opacity var(--matrix-transition-fast);
}

.matrix-transition-shadow {
  transition: box-shadow var(--matrix-transition-fast);
}
EOF

echo "✅ [$(date)] Parte 3.1: Design System Matrix concluída!"
echo "📋 Implementações:"
echo "  ✅ 3.1.1 - CSS Variables Matrix criadas"
echo "  ✅ 3.1.2 - Typography Matrix implementada"
echo "  ✅ 3.1.3 - Animation System criado"
echo ""
echo "🚀 Próximo: Execute 'part3b_matrix_components.sh'"

