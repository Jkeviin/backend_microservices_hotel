-- =====================================================
-- BASE DE DATOS HOTEL_RESERVAS
-- =====================================================

CREATE DATABASE IF NOT EXISTS hotel_reservas 
    DEFAULT CHARACTER SET utf8mb4 
    COLLATE utf8mb4_unicode_ci;

USE hotel_reservas;

-- =====================================================
-- TABLA 1: ESTADOS DE RESERVA
-- =====================================================
CREATE TABLE Estados_Reserva (
    id_estado INT NOT NULL AUTO_INCREMENT,
    nombre VARCHAR(50) NOT NULL,
    PRIMARY KEY (id_estado),
    UNIQUE INDEX nombre (nombre)
);

-- =====================================================
-- TABLA 2: CLIENTES
-- =====================================================
CREATE TABLE Clientes (
    id_cliente INT NOT NULL AUTO_INCREMENT,
    nombre VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL,
    telefono VARCHAR(20) NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id_cliente),
    UNIQUE INDEX email (email)
);

-- =====================================================
-- TABLA 3: TIPOS DE HABITACION
-- =====================================================
CREATE TABLE Tipos_Habitacion (
    id_tipo INT NOT NULL AUTO_INCREMENT,
    nombre VARCHAR(50) NOT NULL,
    precio_base DECIMAL(10,2) NOT NULL,
    capacidad INT NOT NULL,
    PRIMARY KEY (id_tipo)
);

-- =====================================================
-- TABLA 4: HABITACIONES
-- =====================================================
CREATE TABLE Habitaciones (
    id_habitacion INT NOT NULL AUTO_INCREMENT,
    numero VARCHAR(10) NOT NULL,
    id_tipo_habitacion INT NOT NULL,
    estado ENUM('disponible', 'mantenimiento', 'ocupada') NOT NULL DEFAULT 'disponible',
    PRIMARY KEY (id_habitacion),
    UNIQUE INDEX numero (numero),
    INDEX id_tipo_habitacion (id_tipo_habitacion),
    CONSTRAINT FK_id_tipo_Tipos_Habitacion
        FOREIGN KEY (id_tipo_habitacion)
        REFERENCES Tipos_Habitacion (id_tipo)
);

-- =====================================================
-- TABLA 5: RESERVAS
-- =====================================================
CREATE TABLE Reservas (
    id_reserva INT NOT NULL AUTO_INCREMENT,
    id_cliente INT NOT NULL,
    id_habitacion INT NOT NULL,
    id_estado INT NOT NULL DEFAULT 1,
    fecha_inicio DATE NOT NULL,
    fecha_fin DATE NOT NULL,
    total DECIMAL(10,2) NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id_reserva),
    INDEX id_cliente (id_cliente),
    INDEX id_habitacion (id_habitacion),
    INDEX id_estado (id_estado),
    CONSTRAINT FK_id_cliente_Clientes
        FOREIGN KEY (id_cliente)
        REFERENCES Clientes (id_cliente),
    CONSTRAINT FK_id_habitacion_Habitaciones
        FOREIGN KEY (id_habitacion)
        REFERENCES Habitaciones (id_habitacion),
    CONSTRAINT FK_id_estado_Estados_Reserva
        FOREIGN KEY (id_estado)
        REFERENCES Estados_Reserva (id_estado)
);

-- =====================================================
-- TABLA 6: PAGOS
-- =====================================================
CREATE TABLE Pagos (
    id_pago INT NOT NULL AUTO_INCREMENT,
    id_reserva INT NOT NULL,
    monto DECIMAL(10,2) NOT NULL,
    fecha_pago DATE NOT NULL,
    PRIMARY KEY (id_pago),
    INDEX id_reserva (id_reserva),
    CONSTRAINT FK_id_reserva_reservas
        FOREIGN KEY (id_reserva)
        REFERENCES Reservas (id_reserva)
);
