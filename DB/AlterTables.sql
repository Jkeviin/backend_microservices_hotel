-- =====================================================
-- ALTER TABLES - MEJORAS DE TRAZABILIDAD, CONCURRENCIA,
-- IDEMPOTENCIA Y OPTIMIZACION PARA hotel_reservas
-- =====================================================

USE hotel_reservas;

-- =====================================================
-- INDICES COMPUESTOS PARA FILTROS Y SOLAPAMIENTOS
-- =====================================================
ALTER TABLE Reservas
ADD INDEX idx_reservas_hab_fechas (id_habitacion, fecha_inicio, fecha_fin);

-- =====================================================
-- IDEMPOTENCIA EN OPERACIONES CRITICAS (en BD)
-- =====================================================
ALTER TABLE Reservas
ADD COLUMN idempotency_key VARCHAR(64) NULL,
ADD UNIQUE KEY uk_reservas_idem (idempotency_key);

ALTER TABLE Pagos
ADD COLUMN idempotency_key VARCHAR(64) NULL,
ADD UNIQUE KEY uk_pagos_idem (idempotency_key);

-- =====================================================
-- OPTIMISTIC LOCKING Y TIMESTAMPS PARA PARCHES SEGUROS
-- =====================================================
ALTER TABLE Reservas
ADD COLUMN version INT NOT NULL DEFAULT 1,
ADD COLUMN updated_at TIMESTAMP NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP;

ALTER TABLE Clientes
ADD COLUMN updated_at TIMESTAMP NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP;

ALTER TABLE Habitaciones
ADD COLUMN updated_at TIMESTAMP NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP;

ALTER TABLE Pagos
ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- =====================================================
-- ESTADO DE PAGO (MEJORA DE TRAZABILIDAD)
-- =====================================================
ALTER TABLE Pagos
ADD COLUMN estado_pago ENUM('pendiente','aprobado','rechazado') NOT NULL DEFAULT 'aprobado';

-- =====================================================
-- INTEGRIDAD DE FECHAS (MySQL >= 8.0.16)
-- =====================================================
ALTER TABLE Reservas
ADD CONSTRAINT chk_reservas_fechas CHECK (fecha_fin > fecha_inicio);
