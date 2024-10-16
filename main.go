package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"strconv"
)

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

func Get_cats(c *fiber.Ctx) error {
	//Обработка get запроса
	return c.JSON(Cats)
}

func Get_cat(c *fiber.Ctx) error {
	//Обработка get запроса по id
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}
	return c.JSON(Cats[id])
}

func Delete_cat(c *fiber.Ctx) error {
	//Обработка delete запроса по id
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}
	delete(Cats, id)
	return c.JSON(Cats)
}

func Create_cat(c *fiber.Ctx) error {
	//Обработка post запроса по id
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
