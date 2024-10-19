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

var Cats = map[int]string{
	1: "Бенгал",
	2: "Британская",
	3: "Сиамская",
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

func validateToken(tokenString string) (jwt.Claims, error) {
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
	// Проверяем валидность токена
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
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
	claims, err := validateToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	sub := claims.(jwt.MapClaims)["sub"]
	_, exist := Users[sub.(string)]
	if !exist {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Неверные данные",
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
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
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
	Cats[cat.Id] = cat.Name
	return c.Status(fiber.StatusCreated).JSON(cat)
}

func PutCat(c *fiber.Ctx) error {
	//Обработка put запроса по id
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
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
	Cats[id] = cat.Name
	return c.Status(fiber.StatusCreated).JSON(cat)
}
