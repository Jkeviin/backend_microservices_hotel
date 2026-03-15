package config

import "os"

// Config contiene la configuracion del servicio de inventario.
type Config struct {
	DBDsn     string
	JWTSecret string
	Port      string
	UseMockDB bool
}

// Load carga la configuracion desde variables de entorno con valores por defecto.
func Load() Config {
	return Config{
		DBDsn:     getEnv("DB_DSN", "root:root@tcp(127.0.0.1:3306)/hotel_reservas?parseTime=true"),
		JWTSecret: getEnv("JWT_SECRET", "dev-secret-change-me"),
		Port:      getEnv("PORT", "8082"),
		UseMockDB: getEnv("USE_MOCK_DB", "true") == "true",
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
