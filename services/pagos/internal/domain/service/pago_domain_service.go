package service

import "pagos/internal/domain/model"

// PagoDomainService contiene reglas de negocio del dominio de pagos.
type PagoDomainService struct{}

// NewPagoDomainService crea una nueva instancia del servicio de dominio.
func NewPagoDomainService() *PagoDomainService {
	return &PagoDomainService{}
}

// ValidarMonto delega al value object la validacion del monto.
func (s *PagoDomainService) ValidarMonto(monto float64) error {
	return model.ValidarMonto(monto)
}
