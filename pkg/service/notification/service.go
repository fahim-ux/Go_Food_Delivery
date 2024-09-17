package notification

import (
	"Go_Food_Delivery/pkg/database"
	"Go_Food_Delivery/pkg/nats"
	"github.com/gorilla/websocket"
	"log/slog"
)

type NotificationService struct {
	db   database.Database
	env  string
	nats *nats.NATS
}

func NewNotificationService(db database.Database, env string, nats *nats.NATS) *NotificationService {
	return &NotificationService{db, env, nats}
}

func (s *NotificationService) SubscribeNewOrders(clients map[*websocket.Conn]bool) error {
	slog.Info("NotificationService::SubscribeNewOrders")

	err := s.nats.Sub("orders.new.*", clients)
	if err != nil {
		return err
	}
	return nil
}
