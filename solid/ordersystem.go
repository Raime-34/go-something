package main

// RepositoryWriter — запись заказа.
type RepositoryWriter interface {
	CreateOrder(string, []string, float64) error
}

// SchemaInitializer — создание таблиц.
// Отделён от записи: OrderService этот метод не вызывает.
type SchemaInitializer interface {
	InitSchema() error
}

// Notifier — отправка уведомления клиенту.
type Notifier interface {
	Send(string)
}

type OrderService struct {
	db       RepositoryWriter
	notifier Notifier
}

func NewOrderService(db RepositoryWriter, notifier Notifier) *OrderService {
	return &OrderService{
		db:       db,
		notifier: notifier,
	}
}

func (s *OrderService) ProceedOrder(customer string, products []string, total float64) error {
	err := s.db.CreateOrder(customer, products, total)
	if err != nil {
		return err
	}

	s.notifier.Send(customer)

	return nil
}
