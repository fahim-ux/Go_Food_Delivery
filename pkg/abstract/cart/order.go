package cart

import (
	"Go_Food_Delivery/pkg/database/models/order"
	"context"
)

type Order interface {
	PlaceOrder(ctx context.Context, cartId int64) (*order.Order, error)
}