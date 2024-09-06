package cart

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

func (s *CartHandler) PlaceNewOrder(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	userID := c.GetInt64("userID")
	cartInfo, err := s.service.GetCartId(ctx, userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	order, err := s.service.PlaceOrder(ctx, cartInfo.CartID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Println(order)
	c.JSON(http.StatusCreated, gin.H{"message": "Order placed!"})

}

func (s *CartHandler) getOrderList(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	userID := c.GetInt64("userID")
	orders, err := s.service.OrderList(ctx, userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
	return
}

func (s *CartHandler) getOrderItemsList(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	userID := c.GetInt64("userID")
	orderID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	orders, err := s.service.OrderItemsList(ctx, userID, orderID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
	return
}
