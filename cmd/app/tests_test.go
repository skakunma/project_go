package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegister(t *testing.T) {
	app := fiber.New()

	// Регистрация маршрута
	app.Post("/register", Register)

	// Пример данных для теста
	user := &RegisterForm{
		Email:    "test@example.com",
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
	user := &RegisterForm{
		Email:    "admin@mail.ru",
		Password: "admin",
	}
	body, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("could not marshal json: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/signin", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("could not send request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %v; got %v", http.StatusOK, resp.StatusCode)
	}
}
