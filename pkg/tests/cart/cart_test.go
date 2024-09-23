package cart

import (
	"Go_Food_Delivery/cmd/api/middleware"
	restroTypes "Go_Food_Delivery/pkg/database/models/restaurant"
	userModel "Go_Food_Delivery/pkg/database/models/user"
	"Go_Food_Delivery/pkg/handler"
	crt "Go_Food_Delivery/pkg/handler/cart"
	"Go_Food_Delivery/pkg/handler/restaurant"
	"Go_Food_Delivery/pkg/handler/user"
	natsPkg "Go_Food_Delivery/pkg/nats"
	"Go_Food_Delivery/pkg/service/cart_order"
	restro "Go_Food_Delivery/pkg/service/restaurant"
	usr "Go_Food_Delivery/pkg/service/user"
	"Go_Food_Delivery/pkg/tests"
	common "Go_Food_Delivery/pkg/tests/restaurant"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-faker/faker/v4"
	"github.com/go-playground/validator/v10"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type TestNATS struct {
	Conn *nats.Conn
}

func NewTestNATS(url string) (*TestNATS, error) {
	nc, err := nats.Connect(url, nats.Name("food-delivery-test-nats"))
	if err != nil {
		log.Fatalf("Error connecting to NATS:: %s", err)
	}
	return &TestNATS{Conn: nc}, err
}

func TestCart(t *testing.T) {
	t.Setenv("APP_ENV", "TEST")
	t.Setenv("STORAGE_TYPE", "local")
	t.Setenv("STORAGE_DIRECTORY", "uploads")
	t.Setenv("LOCAL_STORAGE_PATH", "./tmp")
	testDB := tests.Setup()
	AppEnv := os.Getenv("APP_ENV")
	testServer := handler.NewServer(testDB)
	validate := validator.New()

	ctx := context.Background()

	// Define NATS container request
	req := testcontainers.ContainerRequest{
		Image:        "nats:2.10.20",
		ExposedPorts: []string{"4222/tcp"},
		WaitingFor:   wait.ForListeningPort("4222/tcp"),
	}

	// NATS container
	natsContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("failed to start container: %v", err)
	}
	defer func() {
		_ = natsContainer.Terminate(ctx)
	}()

	host, err := natsContainer.Host(ctx)
	if err != nil {
		log.Fatalf("failed to get host: %v", err)
	}

	port, err := natsContainer.MappedPort(ctx, "4222")
	if err != nil {
		log.Fatalf("failed to get mapped port: %v", err)
	}
	natsURL := fmt.Sprintf("nats://%s:%s", host, port.Port())

	// Connect NATS
	natTestServer, err := NewTestNATS(natsURL)

	middlewares := []gin.HandlerFunc{middleware.AuthMiddleware()}

	// User
	userService := usr.NewUserService(testDB, AppEnv)
	user.NewUserHandler(testServer, "/user", userService, validate)

	// Restaurant
	restaurantService := restro.NewRestaurantService(testDB, AppEnv)
	restaurant.NewRestaurantHandler(testServer, "/restaurant", restaurantService)

	// Cart
	cartService := cart_order.NewCartService(testDB, AppEnv, (*natsPkg.NATS)(natTestServer))
	crt.NewCartHandler(testServer, "/cart", cartService, middlewares, validate)

	var RestaurantResponseID int64
	var RestaurantMenuID int64
	name := faker.Name()
	file := []byte{10, 10, 10, 10, 10} // fake image bytes
	description := faker.Paragraph()
	address := faker.Word()
	city := faker.Word()
	state := faker.Word()

	type FakeRestaurantMenu struct {
		RestaurantID int64   `json:"restaurant_id"`
		Name         string  `json:"name"`
		Description  string  `json:"description"`
		Price        float64 `json:"price"`
		Category     string  `json:"category"`
		Available    bool    `json:"available"`
	}

	type CartParams struct {
		ItemID       int64 `json:"item_id"`
		RestaurantID int64 `json:"restaurant_id"`
		Quantity     int64 `json:"quantity"`
	}

	form := common.FakeRestaurant{
		Name:        name,
		File:        file,
		Description: description,
		Address:     address,
		City:        city,
		State:       state,
	}

	body, contentType, err := common.GenerateData(form)
	if err != nil {
		t.Fatalf("Error generating form-data: %v", err)
	}

	type FakeUser struct {
		User     string `json:"user" faker:"name"`
		Email    string `json:"email" faker:"email"`
		Password string `json:"password" faker:"password"`
	}

	var customUser FakeUser
	var userInfo userModel.User
	_ = faker.FakeData(&customUser)
	userInfo.Email = customUser.Email
	userInfo.Password = customUser.Password

	_, err = userService.Add(ctx, &userInfo)
	if err != nil {
		t.Error(err)
	}

	loginToken, err := userService.Login(ctx, userInfo.ID, "Food Delivery")
	if err != nil {
		t.Fatal(err)
	}

	Token := fmt.Sprintf("Bearer %s", loginToken)

	t.Run("Restaurant::Create", func(t *testing.T) {

		req, _ := http.NewRequest(http.MethodPost, "/restaurant/", body)
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()
		testServer.Gin.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

	})

	t.Run("Restaurant::Listing", func(t *testing.T) {

		type RestaurantResponse struct {
			RestaurantID int64  `json:"restaurant_id"`
			Name         string `json:"name"`
			StoreImage   string `json:"store_image"`
			Description  string `json:"description"`
			Address      string `json:"address"`
			City         string `json:"city"`
			State        string `json:"state"`
			CreatedAt    string `json:"CreatedAt"`
			UpdatedAt    string `json:"UpdatedAt"`
		}

		req, _ := http.NewRequest(http.MethodGet, "/restaurant/", nil)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		testServer.Gin.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var restaurants []RestaurantResponse
		err := json.Unmarshal(w.Body.Bytes(), &restaurants)
		if err != nil {
			t.Fatalf("Failed to decode response body: %v", err)
		}

		// set the restaurantID
		RestaurantResponseID = restaurants[0].RestaurantID

	})

	t.Run("RestaurantMenu::Create", func(t *testing.T) {
		var customMenu FakeRestaurantMenu

		customMenu.Available = true
		customMenu.Price = 40.35
		customMenu.Name = "burger"
		customMenu.Description = "burger"
		customMenu.Category = "FAST_FOODS"
		customMenu.RestaurantID = RestaurantResponseID
		payload, err := json.Marshal(&customMenu)
		if err != nil {
			t.Fatal("Error::", err)
		}
		req, _ := http.NewRequest(http.MethodPost, "/restaurant/menu", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		testServer.Gin.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

	})

	t.Run("RestaurantMenu::List", func(t *testing.T) {
		url := fmt.Sprintf("%s%d", "/restaurant/menu/", RestaurantResponseID)
		req, _ := http.NewRequest(http.MethodGet, url, nil)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		testServer.Gin.ServeHTTP(w, req)

		var menuItems []restroTypes.MenuItem
		err := json.Unmarshal(w.Body.Bytes(), &menuItems)
		if err != nil {
			fmt.Println("Error unmarshalling JSON:", err)
			return
		}

		RestaurantMenuID = menuItems[0].MenuID

		assert.Equal(t, http.StatusOK, w.Code)

	})

	t.Run("Cart::AddItemToCart", func(t *testing.T) {
		var cartParams CartParams
		cartParams.ItemID = RestaurantMenuID
		cartParams.RestaurantID = RestaurantResponseID
		cartParams.Quantity = 1
		payload, err := json.Marshal(&cartParams)
		if err != nil {
			t.Fatal("Error::", err)
		}
		req, _ := http.NewRequest(http.MethodPost, "/cart/add", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", Token)
		w := httptest.NewRecorder()
		fmt.Println(w.Body.String())

		testServer.Gin.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

	})

	// Cleanup
	tests.Teardown(testDB)

}