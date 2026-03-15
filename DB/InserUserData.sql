SET NAMES utf8mb4;
USE hotel_reservas;

-- 5 Estados de Reserva (alineados con los microservicios)
INSERT INTO Estados_Reserva (nombre) VALUES
('pendiente_pago'), ('confirmada'), ('check_in'), ('check_out'), ('cancelada');

-- 3 Tipos de Habitación
INSERT INTO Tipos_Habitacion (nombre, precio_base, capacidad) VALUES 
('Estándar', 150.00, 2),
('Deluxe', 250.00, 3),
('Suite', 400.00, 4);

-- 5 Habitaciones (distribuidas en tipos)
INSERT INTO Habitaciones (numero, id_tipo_habitacion, estado) VALUES 
('101', 1, 'disponible'),
('102', 1, 'disponible'),
('201', 2, 'mantenimiento'),
('202', 2, 'disponible'),
('301', 3, 'disponible');

-- 5 Clientes
INSERT INTO Clientes (nombre, email, telefono) VALUES 
('Juan Pérez', 'juan.perez@email.com', '3001234567'),
('María García', 'maria.garcia@email.com', '3007654321'),
('Carlos López', 'carlos.lopez@email.com', '3011111111'),
('Ana Rodríguez', 'ana.rodriguez@email.com', '3022222222'),
('Luis Martínez', 'luis.martinez@email.com', '3033333333');

-- 6 Reservas (sin id_tipo_habitacion, usando habitaciones existentes)
INSERT INTO Reservas (id_cliente, id_habitacion, id_estado, fecha_inicio, fecha_fin, total) VALUES 
(1, 1, 2, '2026-03-15', '2026-03-18', 450.00),  -- Confirmada, Juan en 101
(2, 2, 2, '2026-03-20', '2026-03-25', 750.00),  -- Confirmada, María en 102
(3, 1, 1, '2026-03-10', '2026-03-12', 300.00),  -- Pendiente, Carlos en 101
(4, 3, 3, '2026-03-05', '2026-03-07', 500.00),  -- Check-in, Ana en 201
(5, 4, 5, '2026-03-01', '2026-03-03', 0.00),    -- Cancelada, Luis en 202
(1, 5, 2, '2026-04-01', '2026-04-05', 1600.00); -- Confirmada, Juan en 301

-- 3 Pagos de ejemplo
INSERT INTO Pagos (id_reserva, monto, fecha_pago) VALUES 
(1, 450.00, '2026-03-14'),  -- Pago reserva 1
(2, 750.00, '2026-03-19'),  -- Pago reserva 2
(6, 1600.00, '2026-03-31'); -- Pago reserva 6
