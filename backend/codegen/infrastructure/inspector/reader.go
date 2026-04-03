package inspector

import (
	"gorm.io/gorm"

	"goadmin/codegen/driver/db"
)

// Reader is the higher-level abstraction used by the codegen inspector layer.
// It keeps the read-only schema contract but allows contextual rebinding.
type Reader interface {
	db.Inspector
	WithContext(database string, schema string) Reader
}

// Factory creates contextual readers from a database handle.
type Factory interface {
	New(db *gorm.DB) Reader
}

// FactoryFunc adapts a function to the Factory interface.
type FactoryFunc func(db *gorm.DB) Reader

// New adapts a function to the Factory interface.
func (f FactoryFunc) New(db *gorm.DB) Reader {
	if f == nil {
		return nil
	}
	return f(db)
}

// DefaultFactory builds GORM-backed readers.
type DefaultFactory struct{}

// New returns a reader built from the provided GORM database handle.
func (DefaultFactory) New(db *gorm.DB) Reader {
	return NewGormInspector(db)
}

// Service wraps a Factory so callers can inject a custom reader builder.
type Service struct {
	factory Factory
}

// NewService creates a service with a default factory when none is provided.
func NewService(factory Factory) *Service {
	if factory == nil {
		factory = DefaultFactory{}
	}
	return &Service{factory: factory}
}

// Open creates a reader and applies optional context labels.
func (s *Service) Open(db *gorm.DB, database string, schema string) Reader {
	if s == nil {
		return nil
	}
	reader := s.factory.New(db)
	if reader == nil {
		return nil
	}
	return reader.WithContext(database, schema)
}
