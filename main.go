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
	publicGroup := app.Group("")
	publicGroup.Get("/cats/", Get_cats)
	publicGroup.Get("/cat/:id", Get_cat)
	publicGroup.Post("/cats/", Create_cat)
	publicGroup.Delete("/cat/:id", Delete_cat)
	publicGroup.Put("/cat/:id", PutCat)
	publicGroup.Post("/register/", Register)
	publicGroup.Post("/signin/", sign_in)
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

type Registerform struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Register(c *fiber.Ctx) error {
	user := new(Registerform)
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

		// Возвращаем секретный ключ для проверки подписи
		return jwtSecretKey, nil
	})

	if err != nil {
		return nil, err
	}
	sub := token.Claims.(jwt.MapClaims)["sub"]
	_, exist := Users[sub.(string)]
	if !exist {
		return nil, fmt.Errorf("invalid token")
	}

	// Проверяем валидность токена
	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return sub.(string), nil
	} else {
		return nil, fmt.Errorf("invalid token")
	}
}

func create_jwt(email string) string {
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

func sign_in(c *fiber.Ctx) error {
	user := new(Registerform)
	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if Users[user.Email] != user.Password {
		fmt.Println(Users)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Не верый пароль.",
		})
	}
	token := create_jwt(user.Email)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error": "None",
		"token": token,
	})

}

func Get_cats(c *fiber.Ctx) error {
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

func Get_cat(c *fiber.Ctx) error {
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

func Delete_cat(c *fiber.Ctx) error {
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

func Create_cat(c *fiber.Ctx) error {
	//Обработка post запроса
	cat := new(Cat)
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
