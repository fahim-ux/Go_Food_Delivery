package user

import (
	"Uber_Food_Delivery/pkg/handler"
	registration "Uber_Food_Delivery/pkg/handler/register"
	"Uber_Food_Delivery/pkg/tests"
	"encoding/json"
	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAddUser(t *testing.T) {
	testDB := tests.Setup()
	testServer := handler.NewServer(testDB)
	registration.NewRegister(testServer, "/register")

	type FakeUser struct {
		User     string `json:"user" faker:"name"`
		Email    string `json:"email" faker:"email"`
		Password string `json:"password" faker:"password"`
	}

	t.Run("should return 201 created", func(t *testing.T) {

		var customUser FakeUser
		_ = faker.FakeData(&customUser)
		payload, err := json.Marshal(&customUser)
		if err != nil {
			t.Fatal(err)
		}

		req, _ := http.NewRequest(http.MethodPost, "/register/user", strings.NewReader(string(payload)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		testServer.Gin().ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		tests.Teardown(testDB)
	})

}