# Hotel Reservas - Sistema de Gestion de Reservas

MVP de sistema de reservas de hotel con arquitectura de microservicios.

## Arquitectura

4 microservicios independientes en Go con DDD (Domain-Driven Design):

| Servicio | Puerto | Responsabilidad |
|----------|--------|----------------|
| Clientes | 8081 | CRUD de huespedes |
| Inventario | 8082 | Habitaciones y tipos |
| Reservas | 8083 | Ciclo de vida de reservas |
| Pagos | 8084 | Registro de pagos |

**Stack tecnologico:**
- Backend: Go 1.22, chi router, JWT (HS256), sqlx
- Frontend: React + Vite (repo separado)
- Base de datos: MySQL 8.0
- Infraestructura: Docker, Docker Compose

### Decisiones arquitectonicas

- **DDD**: Value Objects, Aggregates, Domain Services, errores tipados
- **Servicios autocontenidos**: cada servicio tiene su propio go.mod, middleware, config. Se pueden separar en repos independientes
- **Idempotencia en BD**: columnas `idempotency_key` con UNIQUE KEY en Reservas y Pagos
- **Optimistic locking**: campo `version` en Reservas para control de concurrencia
- **Dual mode**: `USE_MOCK_DB=true` (memoria, sin MySQL) o `false` (MySQL real)

### Estructura del proyecto

```
SQLScripts/
├── DB/
│   ├── DataBaseCreationScript.sql    # Crea BD y tablas
│   ├── AlterTables.sql               # Indices, idempotencia, timestamps
│   ├── InserUserData.sql             # Datos de ejemplo
│   └── ERRDiagram.mwb                # Diagrama ER (MySQL Workbench)
├── services/
│   ├── clientes/                     # :8081
│   ├── inventario/                   # :8082
│   ├── reservas/                     # :8083
│   └── pagos/                        # :8084
├── Aca-Final-BD-main/                # Frontend React (opcional)
├── docker-compose.yml
├── setup.sh                          # Script de setup automatico
├── Hotel_Reservas.postman_collection.json
├── API_ENDPOINTS.md                  # Referencia detallada de API
└── README.md
```

## Inicio rapido

### Opcion 1: Setup automatico (recomendado)

```bash
chmod +x setup.sh
./setup.sh
```

El script automaticamente:
- Verifica/inicia Docker
- Para MySQL local si ocupa el puerto
- Construye y levanta MySQL + 4 microservicios
- Clona e instala el frontend
- Muestra credenciales y URLs

### Opcion 2: Docker Compose manual

```bash
docker-compose up --build -d
```

### Opcion 3: Sin base de datos (modo mock)

```bash
# Compilar
cd services/clientes && go build -o /tmp/hotel-clientes ./cmd/ && cd ../..
cd services/inventario && go build -o /tmp/hotel-inventario ./cmd/ && cd ../..
cd services/reservas && go build -o /tmp/hotel-reservas ./cmd/ && cd ../..
cd services/pagos && go build -o /tmp/hotel-pagos ./cmd/ && cd ../..

# Ejecutar (4 terminales o con &)
JWT_SECRET=dev-secret-change-me PORT=8081 USE_MOCK_DB=true /tmp/hotel-clientes &
JWT_SECRET=dev-secret-change-me PORT=8082 USE_MOCK_DB=true /tmp/hotel-inventario &
JWT_SECRET=dev-secret-change-me PORT=8083 USE_MOCK_DB=true /tmp/hotel-reservas &
JWT_SECRET=dev-secret-change-me PORT=8084 USE_MOCK_DB=true /tmp/hotel-pagos &
```

## Frontend

El frontend React se encuentra en `Aca-Final-BD-main/` o se puede clonar:

```bash
git clone https://github.com/JuanFeUsme/Aca-Final-BD.git Aca-Final-BD-main
cd Aca-Final-BD-main
npm install
npm run dev
```

Abrir http://localhost:5173 - Login con cualquier email y rol `admin`.

## Credenciales

### Base de datos (Docker)

| Campo | Valor |
|-------|-------|
| Host | 127.0.0.1 |
| Puerto | 3306 |
| Database | hotel_reservas |
| Usuario | hotel_user |
| Password | hotel_pass |
| Root | root / rootpass |

### JWT

Los servicios usan `JWT_SECRET=dev-secret-change-me` en modo desarrollo. El frontend genera tokens fake que los servicios aceptan en este modo.

## API

Referencia completa en [API_ENDPOINTS.md](API_ENDPOINTS.md).

Coleccion Postman: [Hotel_Reservas.postman_collection.json](Hotel_Reservas.postman_collection.json)

### Resumen de endpoints

| Metodo | Endpoint | Puerto | Descripcion |
|--------|----------|--------|-------------|
| POST | /api/clientes | 8081 | Crear cliente |
| GET | /api/clientes/{id} | 8081 | Obtener cliente |
| PATCH | /api/clientes/{id} | 8081 | Actualizar cliente |
| GET | /api/habitaciones | 8082 | Listar habitaciones (con filtros) |
| GET | /api/habitaciones/{id} | 8082 | Obtener habitacion |
| PATCH | /api/habitaciones/{id} | 8082 | Actualizar habitacion (solo admin) |
| GET | /api/tipos-habitacion | 8082 | Catalogo de tipos |
| POST | /api/reservas | 8083 | Crear reserva (idempotente) |
| GET | /api/reservas/{id} | 8083 | Obtener reserva |
| PATCH | /api/reservas/{id} | 8083 | Cancelar/reprogramar reserva |
| GET | /api/reservas/estados | 8083 | Catalogo de estados |
| POST | /api/pagos | 8084 | Registrar pago (idempotente) |
| GET | /api/pagos/{id} | 8084 | Obtener pago |

### Flujo tipico

```
1. GET  /api/habitaciones?estado=disponible     → Ver habitaciones libres
2. POST /api/clientes                           → Crear/buscar cliente
3. POST /api/reservas + Idempotency-Key         → Crear reserva (pendiente_pago)
4. POST /api/pagos + Idempotency-Key            → Pagar (reserva → confirmada)
5. PATCH /api/reservas/{id}                     → Cancelar o reprogramar
```

## Variables de entorno

| Variable | Default | Descripcion |
|----------|---------|-------------|
| PORT | 8081-8084 | Puerto HTTP del servicio |
| JWT_SECRET | - | Clave secreta JWT (misma en los 4 servicios) |
| USE_MOCK_DB | true | true = memoria, false = MySQL |
| DB_DSN | - | Connection string MySQL (solo si USE_MOCK_DB=false) |

## Tests

### Unitarios (128 tests, ~91% coverage)

```bash
cd services/clientes && go test ./... -v -count=1      # 34 tests
cd services/inventario && go test ./... -v -count=1     # 37 tests
cd services/reservas && go test ./... -v -count=1       # 38 tests
cd services/pagos && go test ./... -v -count=1          # 19 tests
```

### Integracion (70 tests - curl contra servicios corriendo)

Ver scripts de test en la documentacion.

## Comandos utiles

```bash
# Levantar todo
./setup.sh

# Ver logs
docker-compose logs -f

# Logs de un servicio
docker-compose logs -f reservas

# Detener
docker-compose down

# Detener y borrar datos
docker-compose down -v

# Reconstruir despues de cambios
docker-compose up --build -d

# Conectar a MySQL
docker exec -it hotel_mysql mysql -u hotel_user -photel_pass hotel_reservas
```

## Troubleshooting

| Problema | Solucion |
|----------|----------|
| 401 "token requerido" | Verificar header `Authorization: Bearer <token>` |
| 403 "permisos insuficientes" | El rol del token no tiene acceso. Usar `admin` o `recepcion` |
| "connection refused" | El servicio no esta corriendo. Verificar con `docker-compose ps` |
| MySQL "Access denied" | Verificar credenciales. Con Docker usar `127.0.0.1` (no `localhost`) |
| Puerto 3306 ocupado | Parar MySQL local: `brew services stop mysql` |
| Datos se pierden al reiniciar | Normal en modo mock. Usar Docker para persistencia |
