package main

import (
	"bytes"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegister(t *testing.T) {
	app := fiber.New()

	// Регистрация маршрута
	app.Post("/register", Register)

	// Пример данных для теста
	user := &User{
		Email:    "test12@example.com",
		Password: "password123",
	}

	// Преобразование данных в JSON
	body, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("could not marshal json: %v", err)
	}

	// Создаем запрос
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Создаем ответ
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("could not send request: %v", err)
	}

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status %v; got %v", http.StatusCreated, resp.StatusCode)
	}
}

func TestSignIn(t *testing.T) {
	app := fiber.New()
	app.Post("/signin", SignIn)

	// Подготовка тестового пользователя
	user := &User{
		Email:    "test@example.com",
		Password: "password123",
	}
	body, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("could not marshal json: %v", err)
	}

	// Создание HTTP-запроса
	req := httptest.NewRequest(http.MethodPost, "/signin", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Выполнение запроса
	resp, err := app.Test(req, -1) // Таймаут -1 означает ожидание завершения запроса без ограничения времени
	if err != nil {
		t.Fatalf("could not send request: %v", err)
	}
	defer resp.Body.Close()

	// Проверка статуса ответа
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %v; got %v", http.StatusOK, resp.StatusCode)
	}
}

func TestGetCats(t *testing.T) {
	app := fiber.New()
	app.Get("/cats", GetCats)
	jwtToken := CreateJwt("test@example.com")
	req := httptest.NewRequest(http.MethodGet, "/cats", nil)
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("could not send request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %v; got %v", http.StatusOK, resp.StatusCode)
	}
}

func TestGetCat(t *testing.T) {
	app := fiber.New()
	app.Get("/cat/:id", GetCat)
	jwtToken := CreateJwt("test@example.com")
	req := httptest.NewRequest(http.MethodGet, "/cat/", nil)
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("could not read response body: %v", err)
	}
	if err != nil {
		t.Fatalf("could not send request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %v; got %v", http.StatusOK, resp.StatusCode)
	}

}
