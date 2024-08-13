package review

import (
	reviewModel "Go_Food_Delivery/pkg/database/models/review"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

func (s *ReviewProtectedHandler) addReview(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	userID := c.GetInt64("userID")
	restaurantId, err := strconv.ParseInt(c.Param("restaurant_id"), 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Invalid RestaurantID"})
		return
	}

	var reviewParam reviewModel.ReviewParams
	var review reviewModel.Review
	if err := c.BindJSON(&reviewParam); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := s.validate.Struct(reviewParam); err != nil {
		validationError := reviewModel.ReviewValidationError(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": validationError})
		return
	}

	comment := reviewParam.Comment
	rating := reviewParam.Rating
	review.UserID = userID
	review.RestaurantID = restaurantId
	review.Rating = rating
	review.Comment = comment
	_, err = s.service.Add(ctx, &review)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Review Added!"})

}