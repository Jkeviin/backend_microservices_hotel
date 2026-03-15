#!/bin/bash
# ============================================
# Setup Hotel Reservas - Full Stack
# ============================================
# Levanta MySQL + 4 microservicios + frontend
# Funciona desde un PC limpio con Docker instalado.
# Se puede ejecutar multiples veces sin problema.
#
# Uso: ./setup.sh
# ============================================

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

FRONTEND_DIR="$SCRIPT_DIR/Aca-Final-BD-main"
FRONTEND_REPO="https://github.com/JuanFeUsme/Aca-Final-BD.git"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Hotel Reservas - Setup Completo${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# =============================================
# 1. PREREQUISITOS: Docker, docker-compose, Node
# =============================================
echo -e "${YELLOW}[1/6] Verificando prerequisitos...${NC}"

# Docker instalado?
if ! command -v docker &> /dev/null; then
    echo -e "${RED}ERROR: Docker no esta instalado.${NC}"
    echo ""
    echo "  Opciones de instalacion:"
    echo "    - Docker Desktop: https://www.docker.com/products/docker-desktop"
    echo "    - Colima (macOS): brew install colima docker docker-compose"
    echo "    - Linux:          sudo apt install docker.io docker-compose"
    exit 1
fi

# Docker corriendo?
if ! docker info &> /dev/null; then
    echo -e "${YELLOW}  Docker no esta corriendo. Intentando iniciar...${NC}"
    if command -v colima &> /dev/null; then
        echo -e "${YELLOW}  Iniciando Colima...${NC}"
        colima start
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        echo -e "${YELLOW}  Intentando abrir Docker Desktop...${NC}"
        open -a Docker 2>/dev/null || true
        echo "  Esperando a que Docker Desktop inicie..."
        for i in $(seq 1 30); do
            if docker info &> /dev/null; then break; fi
            sleep 2
        done
    fi
    # Verificar de nuevo
    if ! docker info &> /dev/null; then
        echo -e "${RED}ERROR: No se pudo iniciar Docker.${NC}"
        echo "  Inicialo manualmente y volve a ejecutar este script."
        exit 1
    fi
fi
echo -e "${GREEN}  Docker OK${NC}"

# docker-compose disponible?
COMPOSE_CMD=""
if command -v docker-compose &> /dev/null; then
    COMPOSE_CMD="docker-compose"
elif docker compose version &> /dev/null; then
    COMPOSE_CMD="docker compose"
else
    echo -e "${RED}ERROR: docker-compose no esta instalado.${NC}"
    echo "  Instalar: brew install docker-compose (macOS)"
    echo "            sudo apt install docker-compose (Linux)"
    exit 1
fi
echo -e "${GREEN}  $COMPOSE_CMD OK${NC}"

# Node/npm (para frontend)
if command -v node &> /dev/null; then
    echo -e "${GREEN}  Node $(node -v) OK${NC}"
else
    echo -e "${YELLOW}  Node.js no encontrado. El frontend no se podra ejecutar.${NC}"
    echo "  Instalar: https://nodejs.org/ o 'brew install node'"
fi

# =============================================
# 2. LIMPIAR: contenedores, volumenes, imagenes previas
# =============================================
echo ""
echo -e "${YELLOW}[2/6] Limpiando ejecuciones previas...${NC}"

# Si hay contenedores del proyecto corriendo, pararlos
RUNNING=$(docker ps -q --filter "name=hotel_" 2>/dev/null)
if [ -n "$RUNNING" ]; then
    echo -e "${YELLOW}  Deteniendo contenedores hotel_* en ejecucion...${NC}"
    docker stop $RUNNING 2>/dev/null || true
fi

# Limpiar con compose (borra contenedores + volumenes + red)
$COMPOSE_CMD down -v 2>/dev/null || true

# Limpiar contenedores huerfanos por nombre
for NAME in hotel_mysql hotel_clientes hotel_inventario hotel_reservas_svc hotel_pagos; do
    if docker ps -a --format '{{.Names}}' | grep -q "^${NAME}$"; then
        echo -e "${YELLOW}  Eliminando contenedor huerfano: $NAME${NC}"
        docker rm -f "$NAME" 2>/dev/null || true
    fi
done

echo -e "${GREEN}  Limpio${NC}"

# =============================================
# 3. LIBERAR PUERTO 3306 (MySQL local)
# =============================================
echo ""
echo -e "${YELLOW}[3/6] Verificando puerto 3306...${NC}"

if lsof -i :3306 &> /dev/null; then
    echo -e "${YELLOW}  Puerto 3306 ocupado. Intentando liberar...${NC}"

    # macOS con Homebrew
    if command -v brew &> /dev/null; then
        brew services stop mysql 2>/dev/null || true
        brew services stop mysql@5.7 2>/dev/null || true
        brew services stop mysql@8.0 2>/dev/null || true
        brew services stop mariadb 2>/dev/null || true
    fi

    # Linux con systemctl
    if command -v systemctl &> /dev/null; then
        sudo systemctl stop mysql 2>/dev/null || true
        sudo systemctl stop mysqld 2>/dev/null || true
        sudo systemctl stop mariadb 2>/dev/null || true
    fi

    sleep 2

    # Verificar de nuevo
    if lsof -i :3306 &> /dev/null; then
        echo -e "${RED}  Puerto 3306 sigue ocupado:${NC}"
        lsof -i :3306 | head -5
        echo ""
        echo -e "${RED}  Detene el servicio manualmente y volve a ejecutar.${NC}"
        exit 1
    fi
fi
echo -e "${GREEN}  Puerto 3306 libre${NC}"

# Verificar otros puertos necesarios
for PORT in 8081 8082 8083 8084; do
    if lsof -i :$PORT &> /dev/null; then
        PID=$(lsof -ti :$PORT 2>/dev/null | head -1)
        echo -e "${YELLOW}  Puerto $PORT ocupado (PID: $PID). Liberando...${NC}"
        kill $PID 2>/dev/null || true
        sleep 1
    fi
done

# =============================================
# 4. CONSTRUIR Y LEVANTAR
# =============================================
echo ""
echo -e "${YELLOW}[4/6] Construyendo y levantando servicios...${NC}"
echo -e "${CYAN}  (primera vez puede tardar varios minutos descargando imagenes)${NC}"

$COMPOSE_CMD up --build -d

# Esperar MySQL healthy
echo -e "${YELLOW}  Esperando a que MySQL este listo...${NC}"
for i in $(seq 1 60); do
    STATUS=$(docker inspect --format='{{.State.Health.Status}}' hotel_mysql 2>/dev/null || echo "starting")
    if [ "$STATUS" = "healthy" ]; then
        break
    fi
    if [ "$i" -eq 30 ]; then
        echo -e "${YELLOW}  Aun esperando MySQL...${NC}"
    fi
    sleep 2
done

if [ "$STATUS" != "healthy" ]; then
    echo -e "${RED}  ERROR: MySQL no arranco en 2 minutos.${NC}"
    echo -e "${RED}  Ultimos logs:${NC}"
    docker logs hotel_mysql --tail 30
    exit 1
fi
echo -e "${GREEN}  MySQL healthy${NC}"

# Esperar microservicios
echo -e "${YELLOW}  Verificando microservicios...${NC}"
sleep 3

ALL_OK=true
SERVICES=("clientes:8081" "inventario:8082" "reservas:8083" "pagos:8084")

for SVC in "${SERVICES[@]}"; do
    NAME="${SVC%%:*}"
    PORT="${SVC##*:}"

    # Reintentar hasta 3 veces
    for RETRY in 1 2 3; do
        HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:${PORT}/api/${NAME}" 2>/dev/null || echo "000")
        if [ "$HTTP_CODE" != "000" ]; then
            echo -e "${GREEN}  $NAME (puerto $PORT): OK${NC}"
            break
        fi
        if [ "$RETRY" -lt 3 ]; then sleep 2; fi
    done

    if [ "$HTTP_CODE" = "000" ]; then
        echo -e "${RED}  $NAME (puerto $PORT): NO RESPONDE${NC}"
        echo -e "${RED}    Logs: docker logs hotel_${NAME} --tail 10${NC}"
        ALL_OK=false
    fi
done

if [ "$ALL_OK" = false ]; then
    echo ""
    echo -e "${RED}  Algunos servicios no respondieron. Revisa los logs.${NC}"
fi

# =============================================
# 5. FRONTEND
# =============================================
echo ""
echo -e "${YELLOW}[5/6] Configurando frontend...${NC}"

# Clonar si no existe
if [ ! -d "$FRONTEND_DIR" ]; then
    if command -v git &> /dev/null; then
        echo -e "${YELLOW}  Clonando repositorio del frontend...${NC}"
        git clone "$FRONTEND_REPO" "$FRONTEND_DIR" 2>/dev/null || {
            echo -e "${YELLOW}  No se pudo clonar. Descargalo manualmente:${NC}"
            echo "  $FRONTEND_REPO"
            echo "  y colocalo en: $FRONTEND_DIR"
        }
    else
        echo -e "${YELLOW}  Git no instalado. Descarga el frontend manualmente:${NC}"
        echo "  $FRONTEND_REPO"
        echo "  y colocalo en: $FRONTEND_DIR"
    fi
fi

# Instalar dependencias si existe
if [ -d "$FRONTEND_DIR" ]; then
    if [ ! -d "$FRONTEND_DIR/node_modules" ]; then
        if command -v npm &> /dev/null; then
            echo -e "${YELLOW}  Instalando dependencias npm...${NC}"
            cd "$FRONTEND_DIR"
            npm install --registry https://registry.npmjs.org 2>&1 | tail -3
            cd "$SCRIPT_DIR"
        else
            echo -e "${YELLOW}  npm no disponible. Instala Node.js para usar el frontend.${NC}"
        fi
    fi
    echo -e "${GREEN}  Frontend listo${NC}"
else
    echo -e "${YELLOW}  Frontend no configurado (opcional)${NC}"
fi

# =============================================
# 6. VERIFICACION FINAL
# =============================================
echo ""
echo -e "${YELLOW}[6/6] Verificacion final...${NC}"

echo -e "${GREEN}  Contenedores:${NC}"
docker ps --filter "name=hotel_" --format "    {{.Names}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null

# =============================================
# RESUMEN
# =============================================
echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  SETUP COMPLETO${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "  ${CYAN}--- Base de datos ---${NC}"
echo -e "  Host:      ${GREEN}127.0.0.1${NC}"
echo -e "  Puerto:    ${GREEN}3306${NC}"
echo -e "  Database:  ${GREEN}hotel_reservas${NC}"
echo -e "  Usuario:   ${GREEN}hotel_user${NC}"
echo -e "  Password:  ${GREEN}hotel_pass${NC}"
echo -e "  Root:      ${GREEN}root / rootpass${NC}"
echo ""
echo -e "  ${CYAN}--- Microservicios ---${NC}"
echo -e "  Clientes:    ${GREEN}http://localhost:8081${NC}"
echo -e "  Inventario:  ${GREEN}http://localhost:8082${NC}"
echo -e "  Reservas:    ${GREEN}http://localhost:8083${NC}"
echo -e "  Pagos:       ${GREEN}http://localhost:8084${NC}"
echo ""
if [ -d "$FRONTEND_DIR" ]; then
    echo -e "  ${CYAN}--- Frontend ---${NC}"
    echo -e "  Iniciar con: ${YELLOW}cd $FRONTEND_DIR && npm run dev${NC}"
    echo -e "  Abrir:       ${GREEN}http://localhost:5173${NC}"
    echo ""
fi
echo -e "  ${CYAN}--- Comandos utiles ---${NC}"
echo -e "  Detener todo:          ${YELLOW}$COMPOSE_CMD down${NC}"
echo -e "  Detener y borrar data: ${YELLOW}$COMPOSE_CMD down -v${NC}"
echo -e "  Ver logs:              ${YELLOW}$COMPOSE_CMD logs -f${NC}"
echo -e "  Reiniciar:             ${YELLOW}./setup.sh${NC}"
echo ""
