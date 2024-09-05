package cart

import (
	"Go_Food_Delivery/pkg/database"
	"Go_Food_Delivery/pkg/database/models/cart"
	"Go_Food_Delivery/pkg/database/models/order"
	"context"
	"errors"
)

func (cartSrv *CartService) PlaceOrder(ctx context.Context, cartId int64) (*order.Order, error) {
	var cartItems []cart.CartItems
	var newOrder order.Order
	var newOrderItems []order.OrderItems
	var orderTotal float64 = 0.0
	var relatedFields = []string{"MenuItem"}
	err := cartSrv.db.SelectWithRelation(ctx, &cartItems, relatedFields, "cart_items.cart_id", cartId)
	if err != nil {
		return nil, err
	}

	if len(cartItems) == 0 {
		return nil, errors.New("no items in cart")
	}

	// Creating a new order.
	newOrder.UserID = 1
	newOrder.OrderStatus = "pending"
	newOrder.TotalAmount = orderTotal
	newOrder.DeliveryAddress = "New Delhi"

	_, err = cartSrv.db.Insert(ctx, &newOrder)
	if err != nil {
		return nil, err
	}

	newOrderItems = make([]order.OrderItems, len(cartItems))
	for i, cartItem := range cartItems {
		newOrderItems[i].OrderID = newOrder.OrderID
		newOrderItems[i].ItemID = cartItem.ItemID
		newOrderItems[i].RestaurantID = cartItem.RestaurantID
		newOrderItems[i].Quantity = cartItem.Quantity
		newOrderItems[i].Price = cartItem.MenuItem.Price * float64(cartItem.Quantity)
		_, err = cartSrv.db.Insert(ctx, &newOrderItems[i])
		if err != nil {
			return nil, err
		}
		orderTotal += newOrderItems[i].Price
	}

	_, err = cartSrv.db.Update(ctx, "orders", database.Filter{"total_amount": orderTotal, "order_status": "in_progress"},
		database.Filter{"order_id": newOrder.OrderID})
	if err != nil {
		return nil, err
	}

	//remove all items from the cart.
	filter := database.Filter{"cart_id": cartId}

	_, err = cartSrv.db.Delete(ctx, "cart_items", filter)
	if err != nil {
		return nil, errors.New("failed to delete cart items")
	}
	return nil, err
}
