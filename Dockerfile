# 🚫 DOCKERFILE DESABILITADO - USAR BUILD NATIVO
# Este arquivo existe apenas para falhar e forçar o uso do render.yaml

FROM alpine:latest
RUN echo "🚫 DOCKERFILE DESABILITADO - USAR BUILD NATIVO NO render.yaml"
RUN exit 1
