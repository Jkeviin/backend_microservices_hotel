package config

import "os"

// Config contiene la configuracion del servicio de reservas.
type Config struct {
	DBDsn     string
	JWTSecret string
	Port      string
	UseMockDB bool
}

// LoadConfig lee la configuracion desde variables de entorno.
func LoadConfig() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	useMock := os.Getenv("USE_MOCK_DB")
	if useMock == "" {
		useMock = "true"
	}

	return Config{
		DBDsn:     os.Getenv("DB_DSN"),
		JWTSecret: os.Getenv("JWT_SECRET"),
		Port:      port,
		UseMockDB: useMock == "true",
	}
}
