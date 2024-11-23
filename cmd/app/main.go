package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"strconv"
	"time"
)

var jwtSecretKey = []byte(os.Getenv("JWT_SECRET_KEY"))

type User struct {
	gorm.Model
	Email    string `gorm:"unique"`
	Password string
}

type CatsModel struct {
	gorm.Model
	Name   string
	Author uint
	User   User `gorm:"foreignKey:Author;references:ID"`
}

func main() {
	Migrate(ConnDB())
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

type PutCats struct {
	//Структура для поиска на put
	Name string `json:"name"`
}

func Migrate(db *gorm.DB) {
	db.AutoMigrate(&CatsModel{})
	db.AutoMigrate(&User{})
}

func ConnDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("cats.db"), &gorm.Config{})
	if err != nil {
		fmt.Println("failed to connect database:", err)
	}
	return db
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
	user := new(User)
	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	var existingUser User
	db := ConnDB()
	result := db.Where("email = ?", user.Email).First(&existingUser)
	fmt.Println(result.Error)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": result.Error,
		})
	}

	db.Create(&User{Email: user.Email, Password: user.Password})

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

	email, ok := token.Claims.(jwt.MapClaims)["sub"].(string)
	if !ok || email == "" {
		return nil, fmt.Errorf("invalid or missing email in token")
	}
	db := ConnDB()
	var userInfo User
	result := db.Where("email = ?", email).First(&userInfo)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
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
	fmt.Println(payload)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	t, err := token.SignedString(jwtSecretKey)
	if err != nil {
		return "error with crete jwt"
	}
	return t

}

func SignIn(c *fiber.Ctx) error {
	user := new(User)
	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	db := ConnDB()
	var userInfo User
	result := db.Where("email = ?", user.Email).First(&userInfo)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Не найден email",
		})
	}
	if userInfo.Password != user.Password {
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
	db := ConnDB()
	var cats []CatsModel
	result := db.Find(&cats)
	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": result.Error,
		})
	}

	return c.Status(fiber.StatusOK).JSON(cats)
}

func GetCat(c *fiber.Ctx) error {
	//Обработка get запроса по id
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}
	db := ConnDB()
	var cat CatsModel
	result := db.Where("id = ?", id).First(&cat)
	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": result.Error,
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"id":   id,
		"name": cat.Name,
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
	var cat CatsModel
	var user User
	db := ConnDB()
	resultCat := db.Where("id = ?", id).First(&cat)
	if resultCat.Error != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": resultCat.Error,
		})
	}
	resultUser := db.Where("id = ?", cat.Author).First(&user)
	if resultUser.Error != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": resultUser.Error,
		})
	}
	if user.Email != email {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "You are is not author",
			"author": user.Email,
		})
	}
	db.Delete(CatsModel{}, cat.ID)
	return c.SendStatus(fiber.StatusNoContent)
}

func CreateCat(c *fiber.Ctx) error {
	//Обработка post запроса
	cat := new(CatsModel)
	tokenString := c.Get("Authorization")[7:]
	email, err := validateToken(tokenString)
	db := ConnDB()
	var UserInfo User
	result := db.Where("email = ?", email).First(&UserInfo)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	cat.Author = UserInfo.ID
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
	db.Create(&cat)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"name":   cat.Name,
		"author": cat.Author,
	})
}

func PutCat(c *fiber.Ctx) error {
	// Обработка PUT-запроса для обновления данных кота по его ID
	tokenString := c.Get("Authorization")[7:]
	email, err := validateToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired token",
		})
	}

	// Получаем ID кота из параметров URL
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid cat ID",
		})
	}

	db := ConnDB()

	// Проверяем, существует ли пользователь
	var user User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Проверяем, существует ли кот и принадлежит ли он автору
	var cat CatsModel
	if err := db.First(&cat, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Cat with this ID does not exist",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	if cat.Author != user.ID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You are not the author of this cat",
		})
	}

	// Парсим входные данные
	updateData := new(PutCats)
	if err := c.BodyParser(updateData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input data",
		})
	}

	// Обновляем данные кота
	cat.Name = updateData.Name
	if err := db.Save(&cat).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update cat",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Cat updated successfully",
		"cat": fiber.Map{
			"id":     cat.ID,
			"name":   cat.Name,
			"author": cat.Author,
		},
	})
}
