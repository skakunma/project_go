package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"time"
)

var jwtSecretKey = []byte(os.Getenv("JWT_SECRET_KEY"))

var Users = map[string]string{
	"admin@mail.ru": "admin",
}

var Cats = map[int]map[string]string{
	1: {"name": "Бенгал", "author": "admin@mail.ru"},
	2: {"name": "Британская", "author": "admin@mail.ru"},
	3: {"name": "Сиамская", "author": "admin@mail.ru"},
}

func main() {
	app := fiber.New()
	app.Use(JwtMiddleware)
	publicGroup := app.Group("")
	publicGroup.Get("/cats/", GetCats)
	publicGroup.Get("/cat/:id", GetCat)
	publicGroup.Post("/cats/", CreateCat)
	publicGroup.Delete("/cat/:id", DeleteCat)
	publicGroup.Put("/cat/:id", PutCat)
	publicGroup.Post("/register/", Register)
	publicGroup.Post("/signin/", SignIn)
	logrus.Fatal(app.Listen(":8000"))
}

type Cat struct {
	//Структура для поиска на post
	Name string `json:"name"`
	Id   int    `json:"id"`
}

type PutCats struct {
	//Структура для поиска на put
	Name string `json:"name"`
}

type RegisterForm struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func JwtMiddleware(c *fiber.Ctx) error {
	if c.Path() == "/register/" || c.Path() == "/signin/" {
		return c.Next()
	}
	// Получаем заголовок Authorization
	tokenString := c.Get("Authorization")

	// Проверяем, что заголовок начинается с "Bearer "
	if len(tokenString) < 8 || tokenString[:7] != "Bearer " {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Необходимо предоставить валидный JWT-токен",
		})
	}

	// Извлекаем сам токен
	tokenString = tokenString[7:]

	// Парсим и проверяем токен
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем, что алгоритм токена соответствует ожидаемому
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.NewError(fiber.StatusUnauthorized, "Неправильный метод подписи")
		}
		return jwtSecretKey, nil
	})

	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Неправильный или просроченный JWT-токен",
		})
	}

	return c.Next() // Передаем управление дальше, если токен валиден
}

func Register(c *fiber.Ctx) error {
	user := new(RegisterForm)
	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if _, exist := Users[user.Email]; exist {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Такой email уже зареган",
		})
	}

	Users[user.Email] = user.Password

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"error": "None",
	})
}

func validateToken(tokenString string) (interface{}, error) {
	// Парсинг и проверка подписи
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return jwtSecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	// Получаем email из токена
	email, ok := token.Claims.(jwt.MapClaims)["email"].(string)
	if !ok || email == "" {
		return nil, fmt.Errorf("invalid or missing email in token")
	}

	_, exist := Users[email]
	if !exist {
		return nil, fmt.Errorf("invalid token")
	}

	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return email, nil
	} else {
		return nil, fmt.Errorf("invalid token")
	}
}

func CreateJwt(email string) string {
	payload := jwt.MapClaims{
		"sub": email,
		"exp": time.Now().Add(time.Second * 360).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	t, err := token.SignedString(jwtSecretKey)
	if err != nil {
		return "error with crete jwt"
	}
	return t

}

func SignIn(c *fiber.Ctx) error {
	user := new(RegisterForm)
	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if Users[user.Email] != user.Password {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Не верый пароль.",
		})
	}
	token := CreateJwt(user.Email)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error": "None",
		"token": token,
	})

}

func GetCats(c *fiber.Ctx) error {
	//Обработка get запроса
	tokenString := c.Get("Authorization")[7:]
	_, err := validateToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(Cats)
}

func GetCat(c *fiber.Ctx) error {
	//Обработка get запроса по id
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"id":   id,
		"name": Cats[id],
	})
}

func DeleteCat(c *fiber.Ctx) error {
	//Обработка delete запроса по id
	tokenString := c.Get("Authorization")[7:]
	email, err := validateToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}
	if Cats[id]["author"] != email {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "You are is not author",
		})
	}
	delete(Cats, id)
	return c.SendStatus(fiber.StatusNoContent)
}

func CreateCat(c *fiber.Ctx) error {
	//Обработка post запроса
	cat := new(Cat)
	tokenString := c.Get("Authorization")[7:]
	email, err := validateToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if err := c.BodyParser(cat); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}
	if _, exists := Cats[cat.Id]; exists {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Cat with this ID already exists",
		})
	}
	if Cats[cat.Id] == nil {
		Cats[cat.Id] = make(map[string]string)
	}

	if email != nil {
		Cats[cat.Id]["author"] = email.(string)
	}
	Cats[cat.Id]["name"] = cat.Name
	return c.Status(fiber.StatusCreated).JSON(cat)
}

func PutCat(c *fiber.Ctx) error {
	//Обработка put запроса по id
	tokenString := c.Get("Authorization")[7:]
	email, err := validateToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}
	if Cats[id]["author"] != email {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "You are is not author",
		})
	}
	cat := new(PutCats)
	if err := c.BodyParser(cat); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}
	if _, exists := Cats[id]; !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Cat with this ID does not exists",
		})
	}
	Cats[id]["name"] = cat.Name
	return c.Status(fiber.StatusCreated).JSON(cat)
}
