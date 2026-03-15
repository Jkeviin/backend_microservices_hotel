# API Endpoints - Hotel Reservas

> Documentacion completa para el equipo de Frontend.
> Todos los endpoints requieren autenticacion JWT via header `Authorization: Bearer <token>`.

## Base URLs por Servicio

| Servicio | Puerto | Base URL |
|----------|--------|----------|
| Clientes | 8081 | `http://localhost:8081` |
| Inventario | 8082 | `http://localhost:8082` |
| Reservas | 8083 | `http://localhost:8083` |
| Pagos | 8084 | `http://localhost:8084` |

---

## Autenticacion y Autorizacion

### JWT Token

Todos los requests deben incluir el header:

```
Authorization: Bearer <jwt_token>
```

El JWT se firma con **HS256** y debe contener los siguientes claims:

| Claim | Tipo | Descripcion | Ejemplo |
|-------|------|-------------|---------|
| `sub` | string | Identificador unico del usuario | `"user-1"` |
| `role` | string | Rol del usuario en el sistema | `"admin"` |
| `exp` | int (unix timestamp) | Fecha de expiracion del token | `1773109653` |

### Roles del Sistema

| Rol | Descripcion | Permisos |
|-----|-------------|----------|
| `admin` | Administrador del hotel | Lectura + escritura en todos los servicios. Unico que puede cambiar estado de habitaciones |
| `recepcion` | Personal de recepcion | Lectura + escritura en clientes, reservas, pagos. Lectura en inventario |
| `lectura` | Solo lectura | **No tiene acceso** a ningun endpoint actualmente. Todos devuelven 403 |

> **Nota para Frontend**: El rol `lectura` fue disenado para futuras ampliaciones. Actualmente todos los endpoints requieren `recepcion` o `admin`.

### Ejemplo: Guardar Token para Curls

```bash
# Guardar token en variable de entorno (reemplazar con token real generado)
export TOKEN="eyJhbGciOiJIUzI1NiIs..."

# Luego usar en cualquier curl
curl http://localhost:8081/api/clientes/1 -H "Authorization: Bearer $TOKEN"
```

---

## Formato de Respuestas

### Respuestas exitosas

Todas las respuestas exitosas devuelven JSON con `Content-Type: application/json`.

- **GET**: Retorna el recurso o array de recursos con status `200`
- **POST**: Retorna el recurso creado con status `201`
- **PATCH**: Retorna el recurso actualizado con status `200`

### Respuestas de error

Todos los errores devuelven JSON con un campo `error`:

```json
{
  "error": "descripcion del error en espanol"
}
```

---

## Codigos de Error Comunes (todos los servicios)

| Status | Significado | Cuando ocurre | Accion sugerida en Frontend |
|--------|-------------|---------------|----------------------------|
| `400` | Bad Request | Datos de entrada invalidos: campos faltantes, formato incorrecto, validaciones fallidas | Mostrar mensaje de validacion al usuario. Revisar campos enviados |
| `401` | Unauthorized | Token JWT faltante, expirado o malformado | Redirigir al login. Renovar token si es posible |
| `403` | Forbidden | Token valido pero el rol no tiene permisos para esta operacion | Mostrar "No tienes permisos para esta accion" |
| `404` | Not Found | El recurso solicitado no existe | Mostrar "No encontrado" o redirigir |
| `409` | Conflict | Conflicto de datos: email duplicado, solapamiento de fechas, version desactualizada | Depende del caso: mostrar "ya existe", recargar datos y reintentar, etc |
| `422` | Unprocessable Entity | Los datos son validos pero las reglas de negocio no permiten la operacion | Mostrar el mensaje de error al usuario (ej: "habitacion en mantenimiento") |
| `429` | Too Many Requests | Rate limit excedido (maximo 30 req/min en endpoints de escritura por IP) | Esperar y reintentar. Mostrar "Demasiadas solicitudes, intenta en un momento" |
| `500` | Internal Server Error | Error inesperado del servidor | Mostrar error generico. Reintentar una vez. Si persiste, reportar |

---

## Rate Limiting

Los endpoints de **escritura** (POST, PATCH) tienen rate limiting de **30 requests por minuto por IP**.

- Header de respuesta cuando se excede: `429 Too Many Requests`
- Los endpoints GET (lectura) **no** tienen rate limit

---

## 1. Clientes (puerto 8081)

### POST /api/clientes

Crea un nuevo cliente en el sistema.

**Roles permitidos**: `recepcion`, `admin`

**Request Body**:

| Campo | Tipo | Requerido | Validaciones | Ejemplo |
|-------|------|-----------|--------------|---------|
| `nombre` | string | SI | No puede estar vacio. Se aplica `trim()` automaticamente (se eliminan espacios al inicio y final) | `"Carlos Martinez"` |
| `email` | string | SI | Debe contener `@` con texto a ambos lados, el dominio debe tener al menos un punto (ej: `user@domain.com`). **Debe ser unico** en todo el sistema | `"carlos@email.com"` |
| `telefono` | string | NO | Puede ser `null`, string vacio, o no enviarse. No tiene validacion de formato | `"3001234567"` |

**Ejemplo curl - Caso exitoso**:

```bash
curl -X POST http://localhost:8081/api/clientes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "nombre": "Carlos Martinez",
    "email": "carlos.martinez@email.com",
    "telefono": "3001234567"
  }'
```

**Response 201 Created**:

```json
{
  "id": 6,
  "nombre": "Carlos Martinez",
  "email": "carlos.martinez@email.com",
  "telefono": "3001234567",
  "created_at": "2026-03-09T15:30:00Z"
}
```

| Campo respuesta | Tipo | Descripcion |
|-----------------|------|-------------|
| `id` | int | ID autogenerado del cliente |
| `nombre` | string | Nombre ya con trim aplicado |
| `email` | string | Email registrado |
| `telefono` | string o null | Telefono si fue proporcionado |
| `created_at` | string (ISO 8601) | Fecha de creacion en UTC |

**Ejemplos de error**:

```bash
# Email duplicado → 409
curl -X POST http://localhost:8081/api/clientes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"nombre": "Otro Usuario", "email": "carlos.martinez@email.com", "telefono": "300"}'
# Response: {"error":"el email ya esta registrado"}

# Email invalido → 400
curl -X POST http://localhost:8081/api/clientes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"nombre": "Test", "email": "noesvalido", "telefono": "300"}'
# Response: {"error":"email invalido"}

# Nombre vacio → 400
curl -X POST http://localhost:8081/api/clientes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"nombre": "", "email": "nuevo@email.com"}'
# Response: {"error":"nombre es requerido"}

# Sin autenticacion → 401
curl -X POST http://localhost:8081/api/clientes \
  -H "Content-Type: application/json" \
  -d '{"nombre": "Test", "email": "t@e.com"}'
# Response: {"error":"token requerido"}

# Rol no permitido → 403
curl -X POST http://localhost:8081/api/clientes \
  -H "Authorization: Bearer $TOKEN_HUESPED" \
  -H "Content-Type: application/json" \
  -d '{"nombre": "Test", "email": "t@e.com"}'
# Response: {"error":"permisos insuficientes"}
```

**Tabla de errores completa**:

| Status | Causa | Mensaje de error |
|--------|-------|------------------|
| 400 | Nombre vacio o solo espacios | `"nombre es requerido"` |
| 400 | Email con formato invalido | `"email invalido"` |
| 400 | Body JSON mal formado | `"body JSON invalido"` |
| 401 | Sin header Authorization | `"token requerido"` |
| 401 | Token expirado o invalido | `"token invalido"` |
| 403 | Rol no autorizado | `"permisos insuficientes"` |
| 409 | Email ya registrado | `"el email ya esta registrado"` |
| 429 | Rate limit excedido | HTTP 429 |

---

### GET /api/clientes/{id}

Obtiene un cliente por su ID numerico.

**Roles permitidos**: `recepcion`, `admin`

**Parametros de URL**:

| Parametro | Tipo | Descripcion |
|-----------|------|-------------|
| `id` | int | ID numerico del cliente. Debe ser un entero positivo |

**Ejemplo curl**:

```bash
curl http://localhost:8081/api/clientes/1 \
  -H "Authorization: Bearer $TOKEN"
```

**Response 200 OK**:

```json
{
  "id": 1,
  "nombre": "Juan Perez",
  "email": "juan.perez@email.com",
  "telefono": "3001234567",
  "created_at": "2026-01-15T10:00:00Z",
  "updated_at": "2026-02-20T14:30:00Z"
}
```

| Campo respuesta | Tipo | Descripcion |
|-----------------|------|-------------|
| `id` | int | ID del cliente |
| `nombre` | string | Nombre completo |
| `email` | string | Email unico |
| `telefono` | string o null | Telefono (puede ser null si no se registro) |
| `created_at` | string (ISO 8601) | Fecha de creacion |
| `updated_at` | string (ISO 8601) o null | Fecha de ultima modificacion. `null` si nunca fue modificado |

**Errores**:

| Status | Causa | Ejemplo |
|--------|-------|---------|
| 400 | ID no es un entero valido | `GET /api/clientes/abc` |
| 404 | Cliente no encontrado | `GET /api/clientes/999` |

---

### PATCH /api/clientes/{id}

Actualiza parcialmente un cliente. **Solo se envian los campos que se quieren modificar**.

**Roles permitidos**: `recepcion`, `admin`

**Request Body** (todos los campos son opcionales, pero al menos uno debe enviarse):

| Campo | Tipo | Validaciones |
|-------|------|--------------|
| `nombre` | string | No vacio si se envia. Se aplica trim |
| `email` | string | Mismo formato que en POST. Debe ser unico (excluye al propio cliente) |
| `telefono` | string | Sin validacion especial |

**Ejemplo curl - Cambiar solo el telefono**:

```bash
curl -X PATCH http://localhost:8081/api/clientes/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"telefono": "3009999999"}'
```

**Ejemplo curl - Cambiar nombre y email**:

```bash
curl -X PATCH http://localhost:8081/api/clientes/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "nombre": "Juan Carlos Perez",
    "email": "jc.perez@email.com"
  }'
```

**Response 200 OK**: Mismo formato que GET (retorna el cliente completo actualizado).

**Errores**:

| Status | Causa |
|--------|-------|
| 400 | Email invalido o nombre vacio |
| 404 | Cliente no encontrado |
| 409 | Email ya registrado por otro cliente |

---

## 2. Inventario (puerto 8082)

### GET /api/habitaciones

Lista habitaciones con filtros opcionales. Retorna un array (puede estar vacio si no hay resultados).

**Roles permitidos**: `recepcion`, `admin`

**Query Params**:

| Parametro | Tipo | Requerido | Validaciones | Ejemplo |
|-----------|------|-----------|--------------|---------|
| `tipo` | int | NO | Debe ser un ID entero valido de tipo de habitacion | `?tipo=1` |
| `estado` | string | NO | Uno de: `disponible`, `mantenimiento`, `ocupada` | `?estado=disponible` |
| `disponible_desde` | string | NO* | Formato `YYYY-MM-DD` | `?disponible_desde=2026-04-01` |
| `disponible_hasta` | string | NO* | Formato `YYYY-MM-DD`. Debe ser posterior a `disponible_desde` | `?disponible_hasta=2026-04-05` |

> **IMPORTANTE**: Si se envia un parametro de fecha, **DEBEN enviarse ambos** (`disponible_desde` Y `disponible_hasta`). Enviar solo uno causa error 400.

> Los filtros son **acumulativos** (AND). Ejemplo: `?tipo=1&estado=disponible` filtra habitaciones de tipo 1 QUE ADEMAS esten disponibles.

**Ejemplo curl - Todas las habitaciones**:

```bash
curl "http://localhost:8082/api/habitaciones" \
  -H "Authorization: Bearer $TOKEN"
```

**Ejemplo curl - Filtrar por tipo**:

```bash
curl "http://localhost:8082/api/habitaciones?tipo=2" \
  -H "Authorization: Bearer $TOKEN"
```

**Ejemplo curl - Habitaciones disponibles en un rango de fechas**:

```bash
curl "http://localhost:8082/api/habitaciones?estado=disponible&disponible_desde=2026-04-01&disponible_hasta=2026-04-05" \
  -H "Authorization: Bearer $TOKEN"
```

**Ejemplo curl - Solo habitaciones en mantenimiento**:

```bash
curl "http://localhost:8082/api/habitaciones?estado=mantenimiento" \
  -H "Authorization: Bearer $TOKEN"
```

**Response 200 OK**:

```json
[
  {
    "id": 1,
    "numero": "101",
    "id_tipo_habitacion": 1,
    "estado": "disponible"
  },
  {
    "id": 2,
    "numero": "102",
    "id_tipo_habitacion": 1,
    "estado": "disponible"
  }
]
```

| Campo | Tipo | Descripcion |
|-------|------|-------------|
| `id` | int | ID interno de la habitacion |
| `numero` | string | Numero visible de la habitacion (ej: "101", "202") |
| `id_tipo_habitacion` | int | FK al tipo de habitacion (ver GET /api/tipos-habitacion para el catalogo) |
| `estado` | string | Estado actual: `disponible`, `mantenimiento`, `ocupada` |

> **Nota**: Si no hay resultados, retorna un array vacio `[]` con status 200 (no 404).

**Errores**:

| Status | Causa | Ejemplo |
|--------|-------|---------|
| 400 | `tipo` no es entero | `?tipo=abc` |
| 400 | `estado` invalido | `?estado=roto` |
| 400 | Solo una fecha enviada | `?disponible_desde=2026-04-01` (falta `disponible_hasta`) |
| 400 | `disponible_hasta` <= `disponible_desde` | Fechas invertidas |
| 400 | Formato de fecha invalido | `?disponible_desde=01-04-2026` |

---

### GET /api/habitaciones/{id}

Obtiene una habitacion por ID.

**Roles permitidos**: `recepcion`, `admin`

**Ejemplo curl**:

```bash
curl http://localhost:8082/api/habitaciones/1 \
  -H "Authorization: Bearer $TOKEN"
```

**Response 200 OK**:

```json
{
  "id": 1,
  "numero": "101",
  "id_tipo_habitacion": 1,
  "estado": "disponible"
}
```

**Errores**:

| Status | Causa |
|--------|-------|
| 400 | ID no es entero valido |
| 404 | Habitacion no encontrada |

---

### PATCH /api/habitaciones/{id}

Actualiza estado o tipo de una habitacion.

**Roles permitidos**: solo `admin` (recepcion recibe 403)

**Request Body** (campos opcionales):

| Campo | Tipo | Validaciones |
|-------|------|--------------|
| `estado` | string | Uno de: `disponible`, `mantenimiento`, `ocupada`. **Restriccion**: NO se puede cambiar a `mantenimiento` si esta `ocupada` |
| `id_tipo_habitacion` | int | ID de un tipo de habitacion existente |

**Ejemplo curl - Poner en mantenimiento**:

```bash
curl -X PATCH http://localhost:8082/api/habitaciones/3 \
  -H "Authorization: Bearer $TOKEN_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{"estado": "mantenimiento"}'
```

**Response 200 OK**: Mismo formato que GET.

**Ejemplos de error**:

```bash
# Recepcion intenta cambiar estado → 403
curl -X PATCH http://localhost:8082/api/habitaciones/1 \
  -H "Authorization: Bearer $TOKEN_RECEPCION" \
  -H "Content-Type: application/json" \
  -d '{"estado": "mantenimiento"}'
# Response: {"error":"permisos insuficientes"}

# Habitacion ocupada a mantenimiento → 400
curl -X PATCH http://localhost:8082/api/habitaciones/4 \
  -H "Authorization: Bearer $TOKEN_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{"estado": "mantenimiento"}'
# Response: {"error":"no se puede poner en mantenimiento una habitacion ocupada"}

# Estado invalido → 400
curl -X PATCH http://localhost:8082/api/habitaciones/1 \
  -H "Authorization: Bearer $TOKEN_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{"estado": "roto"}'
# Response: {"error":"estado invalido: roto"}
```

**Errores**:

| Status | Causa |
|--------|-------|
| 400 | Estado invalido o transicion no permitida (ocupada → mantenimiento) |
| 403 | Rol no es `admin` |
| 404 | Habitacion no encontrada |

---

### GET /api/tipos-habitacion

Lista el catalogo de tipos de habitacion. Datos de solo lectura (no hay POST/PATCH).

**Roles permitidos**: `recepcion`, `admin`

**Ejemplo curl**:

```bash
curl http://localhost:8082/api/tipos-habitacion \
  -H "Authorization: Bearer $TOKEN"
```

**Response 200 OK**:

```json
[
  { "id": 1, "nombre": "Simple", "precio_base": 80.00, "capacidad": 2 },
  { "id": 2, "nombre": "Doble", "precio_base": 120.00, "capacidad": 3 },
  { "id": 3, "nombre": "Suite", "precio_base": 250.00, "capacidad": 4 }
]
```

| Campo | Tipo | Descripcion |
|-------|------|-------------|
| `id` | int | ID del tipo (usar este valor en `id_tipo_habitacion` de habitaciones) |
| `nombre` | string | Nombre legible del tipo |
| `precio_base` | float | Precio por noche en la moneda local |
| `capacidad` | int | Numero maximo de huespedes |

---

## 3. Reservas (puerto 8083)

### POST /api/reservas

Crea una nueva reserva. Esta operacion es **idempotente**: si se reenvia el mismo `Idempotency-Key`, retorna la reserva original sin crear duplicados.

**Roles permitidos**: `recepcion`, `admin`

**Headers requeridos**:

| Header | Requerido | Descripcion |
|--------|-----------|-------------|
| `Idempotency-Key` | **SI** | Clave unica por operacion. Sugerencia de formato: `reserva-{clienteId}-hab{habId}-{fecha}`. Maximo 64 caracteres |

> **Para Frontend**: Generar la key antes de enviar el request. Si el request falla por timeout, reenviar con la **misma** key para evitar duplicados. El backend detecta la key duplicada y retorna la reserva existente.

**Request Body**:

| Campo | Tipo | Requerido | Validaciones | Ejemplo |
|-------|------|-----------|--------------|---------|
| `id_cliente` | int | SI | Debe ser un ID de cliente existente | `1` |
| `id_habitacion` | int | SI | Debe ser un ID de habitacion existente. La habitacion **no** debe estar en `mantenimiento` | `3` |
| `fecha_inicio` | string | SI | Formato `YYYY-MM-DD`. No puede ser una fecha pasada | `"2026-04-01"` |
| `fecha_fin` | string | SI | Formato `YYYY-MM-DD`. Debe ser **estrictamente posterior** a `fecha_inicio` | `"2026-04-05"` |

**Reglas de negocio**:

1. Se valida que **no exista otra reserva activa** (estado: `pendiente_pago`, `confirmada`, `completada`) en la misma habitacion que se solape con las fechas
2. La habitacion **no debe estar** en estado `mantenimiento`
3. La reserva se crea con estado `pendiente_pago` y `version: 1`
4. La `idempotency_key` se almacena en BD con UNIQUE KEY para garantizar unicidad incluso con MySQL

**Ejemplo curl - Crear reserva**:

```bash
curl -X POST http://localhost:8083/api/reservas \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: reserva-cliente1-hab3-20260401" \
  -d '{
    "id_cliente": 1,
    "id_habitacion": 3,
    "fecha_inicio": "2026-04-01",
    "fecha_fin": "2026-04-05"
  }'
```

**Response 201 Created**:

```json
{
  "id": 4,
  "id_cliente": 1,
  "id_habitacion": 3,
  "estado": "pendiente_pago",
  "fecha_inicio": "2026-04-01",
  "fecha_fin": "2026-04-05",
  "version": 1,
  "created_at": "2026-03-09T15:30:00Z"
}
```

| Campo | Tipo | Descripcion |
|-------|------|-------------|
| `id` | int | ID autogenerado de la reserva |
| `id_cliente` | int | ID del cliente asociado |
| `id_habitacion` | int | ID de la habitacion reservada |
| `estado` | string | Estado actual de la reserva (ver seccion Estados de Reserva) |
| `fecha_inicio` | string | Fecha de check-in (formato `YYYY-MM-DD`) |
| `fecha_fin` | string | Fecha de check-out (formato `YYYY-MM-DD`) |
| `total` | float o null | Total calculado (null si aun no se ha calculado) |
| `version` | int | **Importante**: Numero de version para optimistic locking. Enviar este valor al hacer PATCH |
| `created_at` | string (ISO 8601) | Fecha de creacion |
| `updated_at` | string (ISO 8601) o null | Fecha de ultima modificacion |

**Ejemplo - Idempotency replay (misma key, misma respuesta)**:

```bash
# Segundo envio con misma key → retorna la reserva original (sin duplicar)
curl -X POST http://localhost:8083/api/reservas \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: reserva-cliente1-hab3-20260401" \
  -d '{
    "id_cliente": 1,
    "id_habitacion": 3,
    "fecha_inicio": "2026-04-01",
    "fecha_fin": "2026-04-05"
  }'
# Response: 201 con la misma reserva (id: 4)
```

**Ejemplos de error**:

```bash
# Sin header Idempotency-Key → 400
curl -X POST http://localhost:8083/api/reservas \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"id_cliente":1,"id_habitacion":3,"fecha_inicio":"2026-04-01","fecha_fin":"2026-04-05"}'
# Response: {"error":"header Idempotency-Key es requerido para POST"}

# Solapamiento de fechas → 409
curl -X POST http://localhost:8083/api/reservas \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: otra-key" \
  -d '{"id_cliente":2,"id_habitacion":3,"fecha_inicio":"2026-04-03","fecha_fin":"2026-04-08"}'
# Response: {"error":"conflicto de fechas: la habitacion ya tiene una reserva activa en ese rango"}

# Habitacion en mantenimiento → 422
curl -X POST http://localhost:8083/api/reservas \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: key-mantenimiento" \
  -d '{"id_cliente":1,"id_habitacion":6,"fecha_inicio":"2026-04-01","fecha_fin":"2026-04-05"}'
# Response: {"error":"la habitacion no esta disponible"}

# Fechas invertidas → 400
curl -X POST http://localhost:8083/api/reservas \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: key-fechas" \
  -d '{"id_cliente":1,"id_habitacion":5,"fecha_inicio":"2026-04-05","fecha_fin":"2026-04-01"}'
# Response: {"error":"fechas invalidas: fecha_fin debe ser posterior a fecha_inicio"}
```

**Tabla de errores completa**:

| Status | Causa | Mensaje |
|--------|-------|---------|
| 400 | Falta header `Idempotency-Key` | `"header Idempotency-Key es requerido para POST"` |
| 400 | Fechas con formato invalido | `"fechas invalidas..."` |
| 400 | `fecha_fin` <= `fecha_inicio` | `"fechas invalidas: fecha_fin debe ser posterior a fecha_inicio"` |
| 400 | Body JSON mal formado | `"body JSON invalido"` |
| 400 | Campos requeridos faltantes | Depende del campo faltante |
| 409 | Solapamiento con otra reserva | `"conflicto de fechas: la habitacion ya tiene una reserva activa en ese rango"` |
| 422 | Habitacion en mantenimiento | `"la habitacion no esta disponible"` |
| 429 | Rate limit excedido | HTTP 429 |

---

### GET /api/reservas/{id}

Obtiene una reserva por ID.

**Roles permitidos**: `recepcion`, `admin`

**Ejemplo curl**:

```bash
curl http://localhost:8083/api/reservas/1 \
  -H "Authorization: Bearer $TOKEN"
```

**Response 200 OK**:

```json
{
  "id": 1,
  "id_cliente": 1,
  "id_habitacion": 1,
  "estado": "confirmada",
  "fecha_inicio": "2026-03-15",
  "fecha_fin": "2026-03-18",
  "total": 240.00,
  "version": 1,
  "created_at": "2026-03-10T10:00:00Z",
  "updated_at": "2026-03-10T10:05:00Z"
}
```

**Errores**:

| Status | Causa |
|--------|-------|
| 400 | ID no es entero |
| 404 | Reserva no encontrada |

---

### PATCH /api/reservas/{id}

Cancela o reprograma una reserva. Usa **optimistic locking**: debes enviar el campo `version` actual de la reserva. Si alguien mas la modifico entre que la leiste y la envias, recibiras un 409.

**Roles permitidos**: `recepcion`, `admin`

> **Para Frontend**: Siempre obtener la reserva con GET antes de hacer PATCH para tener la `version` actual.

#### Opcion A: Cancelar

| Campo | Tipo | Requerido | Descripcion |
|-------|------|-----------|-------------|
| `accion` | string | SI | Debe ser exactamente `"cancelar"` |
| `version` | int | SI | Version actual de la reserva (obtener del GET previo) |

**Ejemplo curl**:

```bash
# 1. Primero obtener la reserva para conocer la version
curl http://localhost:8083/api/reservas/2 -H "Authorization: Bearer $TOKEN"
# Response: {..., "version": 1, ...}

# 2. Cancelar usando la version obtenida
curl -X PATCH http://localhost:8083/api/reservas/2 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "accion": "cancelar",
    "version": 1
  }'
```

**Response 200 OK**:

```json
{
  "id": 2,
  "id_cliente": 2,
  "id_habitacion": 3,
  "estado": "cancelada",
  "fecha_inicio": "2026-04-01",
  "fecha_fin": "2026-04-05",
  "version": 2,
  "created_at": "2026-03-01T10:00:00Z",
  "updated_at": "2026-03-09T20:41:53Z"
}
```

> Notar que `version` paso de 1 a 2 y `estado` cambio a `"cancelada"`.

**Reglas**:
- No se puede cancelar si ya esta `cancelada` o `completada`
- Si la `version` no coincide → 409 (alguien mas la modifico)

#### Opcion B: Reprogramar

| Campo | Tipo | Requerido | Descripcion |
|-------|------|-----------|-------------|
| `fecha_inicio` | string | SI | Nueva fecha inicio (`YYYY-MM-DD`). Debe ser al menos manana |
| `fecha_fin` | string | SI | Nueva fecha fin. Posterior a `fecha_inicio` |
| `version` | int | SI | Version actual de la reserva |

**Ejemplo curl**:

```bash
curl -X PATCH http://localhost:8083/api/reservas/2 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "fecha_inicio": "2026-05-10",
    "fecha_fin": "2026-05-15",
    "version": 1
  }'
```

**Reglas**:
- No se puede reprogramar si ya inicio (`fecha_inicio` <= hoy)
- No se puede reprogramar si esta `cancelada` o `completada`
- Se valida solapamiento con otras reservas en la misma habitacion
- Si la `version` no coincide → 409

**Errores PATCH (ambas opciones)**:

| Status | Causa | Mensaje ejemplo |
|--------|-------|-----------------|
| 400 | Falta `version` | `"version es requerida"` |
| 400 | Body invalido | `"debe indicar accion='cancelar' o fecha_inicio y fecha_fin para reprogramar"` |
| 400 | Fechas invalidas | `"fechas invalidas..."` |
| 404 | Reserva no encontrada | `"reserva no encontrada"` |
| 409 | Version desactualizada | `"conflicto de version: la reserva fue modificada por otro proceso"` |
| 409 | Estado no permite la accion | `"la reserva ya esta cancelada"` o `"la reserva ya inicio, no se puede reprogramar"` |
| 409 | Solapamiento al reprogramar | `"conflicto de fechas..."` |

---

### GET /api/reservas/estados

Lista el catalogo de estados de reserva (datos fijos, solo lectura).

**Roles permitidos**: `recepcion`, `admin`

**Ejemplo curl**:

```bash
curl http://localhost:8083/api/reservas/estados \
  -H "Authorization: Bearer $TOKEN"
```

**Response 200 OK**:

```json
[
  { "id": 1, "nombre": "pendiente_pago" },
  { "id": 2, "nombre": "confirmada" },
  { "id": 3, "nombre": "cancelada" },
  { "id": 4, "nombre": "completada" }
]
```

---

## 4. Pagos (puerto 8084)

### POST /api/pagos

Registra un pago para una reserva. Operacion **idempotente** (misma logica que reservas).

**Roles permitidos**: `recepcion`, `admin`

**Headers requeridos**:

| Header | Requerido | Descripcion |
|--------|-----------|-------------|
| `Idempotency-Key` | **SI** | Clave unica. Sugerencia: `pago-reserva{id}-{fecha}`. Maximo 64 caracteres |

**Request Body**:

| Campo | Tipo | Requerido | Validaciones | Ejemplo |
|-------|------|-----------|--------------|---------|
| `id_reserva` | int | SI | Debe existir y estar en estado `pendiente_pago` o `confirmada`. No puede estar `cancelada` ni `completada` | `4` |
| `monto` | float | SI | Debe ser **estrictamente mayor a 0**. No se acepta 0 ni negativos | `480.00` |
| `fecha_pago` | string | SI | Formato `YYYY-MM-DD` | `"2026-04-01"` |

**Reglas de negocio**:
1. Al registrar el pago exitosamente, la reserva asociada se actualiza automaticamente a estado `confirmada`
2. El pago se crea siempre con estado `aprobado`
3. La `idempotency_key` se almacena en BD con UNIQUE KEY

**Ejemplo curl - Crear pago**:

```bash
curl -X POST http://localhost:8084/api/pagos \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: pago-reserva4-20260401" \
  -d '{
    "id_reserva": 4,
    "monto": 480.00,
    "fecha_pago": "2026-04-01"
  }'
```

**Response 201 Created**:

```json
{
  "id": 3,
  "id_reserva": 4,
  "monto": 480.00,
  "fecha_pago": "2026-04-01",
  "estado_pago": "aprobado",
  "created_at": "2026-03-09T15:45:00Z"
}
```

| Campo | Tipo | Descripcion |
|-------|------|-------------|
| `id` | int | ID autogenerado del pago |
| `id_reserva` | int | Reserva asociada al pago |
| `monto` | float | Monto pagado |
| `fecha_pago` | string | Fecha del pago (`YYYY-MM-DD`) |
| `estado_pago` | string | Siempre `"aprobado"` al crear |
| `created_at` | string (ISO 8601) | Fecha de creacion |

**Ejemplos de error**:

```bash
# Sin Idempotency-Key → 400
curl -X POST http://localhost:8084/api/pagos \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"id_reserva":4,"monto":480,"fecha_pago":"2026-04-01"}'
# Response: {"error":"header Idempotency-Key es requerido para POST"}

# Monto 0 → 400
curl -X POST http://localhost:8084/api/pagos \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: key1" \
  -d '{"id_reserva":4,"monto":0,"fecha_pago":"2026-04-01"}'
# Response: {"error":"monto debe ser mayor a 0"}

# Monto negativo → 400
curl -X POST http://localhost:8084/api/pagos \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: key-neg" \
  -d '{"id_reserva":4,"monto":-50,"fecha_pago":"2026-04-01"}'
# Response: {"error":"monto debe ser mayor a 0"}

# Reserva cancelada → 422
curl -X POST http://localhost:8084/api/pagos \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: key2" \
  -d '{"id_reserva":3,"monto":100,"fecha_pago":"2026-04-01"}'
# Response: {"error":"la reserva no permite pago (estado actual: cancelada)"}
```

**Tabla de errores completa**:

| Status | Causa |
|--------|-------|
| 400 | Falta header `Idempotency-Key` |
| 400 | Monto <= 0 |
| 400 | Body JSON invalido o campos faltantes |
| 409 | Idempotency key usada con datos diferentes |
| 422 | Reserva no existe o estado no permite pago (`cancelada`, `completada`) |
| 429 | Rate limit excedido |

---

### GET /api/pagos/{id}

Obtiene un pago por ID.

**Roles permitidos**: `recepcion`, `admin`

**Ejemplo curl**:

```bash
curl http://localhost:8084/api/pagos/1 \
  -H "Authorization: Bearer $TOKEN"
```

**Response 200 OK**:

```json
{
  "id": 1,
  "id_reserva": 1,
  "monto": 240.00,
  "fecha_pago": "2026-03-15",
  "estado_pago": "aprobado",
  "created_at": "2026-03-15T10:00:00Z"
}
```

**Errores**:

| Status | Causa |
|--------|-------|
| 404 | Pago no encontrado |

---

## Referencia Rapida de Endpoints

| Recurso | Metodo | Endpoint | Puerto | Roles | Idempotente | Rate Limit |
|---------|--------|----------|--------|-------|-------------|------------|
| Clientes | POST | `/api/clientes` | 8081 | recepcion, admin | No | 30/min |
| Clientes | GET | `/api/clientes/{id}` | 8081 | recepcion, admin | - | No |
| Clientes | PATCH | `/api/clientes/{id}` | 8081 | recepcion, admin | No | 30/min |
| Habitaciones | GET | `/api/habitaciones` | 8082 | recepcion, admin | - | No |
| Habitaciones | GET | `/api/habitaciones/{id}` | 8082 | recepcion, admin | - | No |
| Habitaciones | PATCH | `/api/habitaciones/{id}` | 8082 | **solo admin** | No | 30/min |
| Tipos Hab. | GET | `/api/tipos-habitacion` | 8082 | recepcion, admin | - | No |
| Reservas | POST | `/api/reservas` | 8083 | recepcion, admin | **SI** | 30/min |
| Reservas | GET | `/api/reservas/{id}` | 8083 | recepcion, admin | - | No |
| Reservas | PATCH | `/api/reservas/{id}` | 8083 | recepcion, admin | No | 30/min |
| Estados | GET | `/api/reservas/estados` | 8083 | recepcion, admin | - | No |
| Pagos | POST | `/api/pagos` | 8084 | recepcion, admin | **SI** | 30/min |
| Pagos | GET | `/api/pagos/{id}` | 8084 | recepcion, admin | - | No |

---

## Catalogos (datos de referencia)

### Estados de Reserva

| ID | Estado | Descripcion | Terminal | Permite cancelar | Permite reprogramar | Permite pago |
|----|--------|-------------|----------|------------------|---------------------|--------------|
| 1 | `pendiente_pago` | Reserva creada, esperando pago | No | Si | Si | Si |
| 2 | `confirmada` | Pago recibido, reserva activa | No | Si | Si (si no inicio) | Si |
| 3 | `cancelada` | Reserva cancelada por usuario o admin | **Si** | No | No | No |
| 4 | `completada` | Estadia finalizada | **Si** | No | No | No |

### Estados de Habitacion

| Estado | Descripcion | Puede recibir reservas |
|--------|-------------|------------------------|
| `disponible` | Lista para reservar | Si |
| `mantenimiento` | No disponible temporalmente | **No** |
| `ocupada` | Actualmente en uso | Si (para fechas futuras) |

### Estados de Pago

| Estado | Descripcion |
|--------|-------------|
| `pendiente` | Pago registrado pero no procesado |
| `aprobado` | Pago confirmado (estado por defecto al crear) |
| `rechazado` | Pago rechazado |

---

## Flujo Tipico de una Reserva (para Frontend)

```
1. Buscar habitaciones disponibles
   GET /api/habitaciones?estado=disponible&disponible_desde=2026-04-01&disponible_hasta=2026-04-05

2. Buscar o crear cliente
   GET /api/clientes/{id}
   o POST /api/clientes

3. Crear reserva
   POST /api/reservas (con Idempotency-Key)
   → Estado inicial: pendiente_pago, version: 1

4. Registrar pago
   POST /api/pagos (con Idempotency-Key)
   → La reserva pasa automaticamente a "confirmada"

5. (Opcional) Cancelar reserva
   GET /api/reservas/{id}  → obtener version actual
   PATCH /api/reservas/{id} con {"accion":"cancelar","version":N}

6. (Opcional) Reprogramar reserva
   GET /api/reservas/{id}  → obtener version actual
   PATCH /api/reservas/{id} con {"fecha_inicio":"...","fecha_fin":"...","version":N}
```

---

## Datos de Prueba Pre-cargados (modo mock)

Cuando los servicios corren con `USE_MOCK_DB=true` (default), tienen estos datos en memoria:

### Clientes (5 registros)

| ID | Nombre | Email |
|----|--------|-------|
| 1 | Juan Perez | juan.perez@email.com |
| 2 | Maria Garcia | maria.garcia@email.com |
| 3 | Carlos Lopez | carlos.lopez@email.com |
| 4 | Ana Rodriguez | ana.rodriguez@email.com |
| 5 | Luis Martinez | luis.martinez@email.com |

### Habitaciones (9 registros)

| ID | Numero | Tipo | Estado |
|----|--------|------|--------|
| 1 | 101 | 1 (Simple) | disponible |
| 2 | 102 | 1 (Simple) | disponible |
| 3 | 201 | 2 (Doble) | disponible |
| 4 | 202 | 2 (Doble) | ocupada |
| 5 | 301 | 3 (Suite) | disponible |
| 6 | 302 | 3 (Suite) | mantenimiento |
| 7 | 401 | 2 (Doble) | disponible |
| 8 | 402 | 1 (Simple) | disponible |
| 9 | 501 | 3 (Suite) | disponible |

### Reservas (3 registros)

| ID | Cliente | Habitacion | Estado | Fechas |
|----|---------|------------|--------|--------|
| 1 | 1 | 1 | confirmada | 2025-03-15 → 2025-03-18 |
| 2 | 2 | 3 | pendiente_pago | 2025-04-01 → 2025-04-05 |
| 3 | 3 | 2 | cancelada | 2025-02-10 → 2025-02-12 |

### Pagos (2 registros)

| ID | Reserva | Monto | Estado |
|----|---------|-------|--------|
| 1 | 1 | 240.00 | aprobado |
| 2 | 2 | 300.00 | aprobado |
