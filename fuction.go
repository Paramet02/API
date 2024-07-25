package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"strconv"
	"time"
)

// Handlers
func getMenus(c *fiber.Ctx) error {
	// Retrieve user data from the context
	user := c.Locals(userContextKey).(*UserData)

	// Use the user data (e.g., for authorization, custom responses, etc.)
	fmt.Printf("User Email: %s, Role: %s\n", user.Email, user.Role)

	return c.JSON(Menus)
}

func getMenu(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id")) // Receive the value to wipe the value to see if it's err or not.
	if err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	for _, Menu := range Menus {
		if Menu.MenuId == id {
			return c.JSON(Menu)
		}
	}

	return c.SendStatus(fiber.StatusNotFound)
}

func createMenu(c *fiber.Ctx) error {
	Menu := new(Menu) // Reserve the value in the address

	if err := c.BodyParser(Menu); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	// add len + 1
	Menu.MenuId = len(Menus) + 1
	// add value in address that has already been reserved
	Menus = append(Menus, *Menu)

	return c.JSON(Menu)
}

func updateMenu(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id")) // Receive the value to wipe the value to see if it's err or not.
	if err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	menuUpdate := new(Menu) // Reserve the value in the address
	if err := c.BodyParser(menuUpdate); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	for index, Menu := range Menus {
		if Menu.MenuId == id {
			Menu.Name = menuUpdate.Name
			Menu.Price = menuUpdate.Price
			Menu.Category = menuUpdate.Category
			Menus[index] = Menu
			return c.JSON(Menu)
		}
	}

	return c.SendStatus(fiber.StatusNotFound)
}

func deleteMenu(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id")) // Receive the value to wipe the value to see if it's err or not.

	if err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	for index, Menu := range Menus {
		if Menu.MenuId == id {
			// [1,2,3,4,5] = [1,2] + [4,5] = [1,2,4,5]
			Menus = append(Menus[:index], Menus[index+1:]...)
			return c.SendStatus(fiber.StatusNoContent)
		}
	}

	return c.SendStatus(fiber.StatusNotFound)
}

// Order

func getOrders(c *fiber.Ctx) error {
	return c.JSON(Orders)
}

func updateOrder(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))

	if err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	orderUpdate := new(Order)
	if err := c.BodyParser(orderUpdate); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	for index, Order := range Orders {
		if Order.OrderId == id {
			Order.Quantity = orderUpdate.Quantity
			Order.Status = orderUpdate.Status
			Orders[index] = Order
			return c.JSON(Order)
		}
	}
	return c.SendStatus(fiber.StatusNotFound)
}

func uploadImage(c *fiber.Ctx) error {
	// Read file from request
	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	// Save the file to the server
	err = c.SaveFile(file, "./uploads/"+file.Filename)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.SendString("File uploaded successfully: " + file.Filename)
}

func login(secretKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		type LoginRequest struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		var request LoginRequest
		if err := c.BodyParser(&request); err != nil {
			return err
		}

		// Check credentials - In real world, you should check against a database
		if request.Email != user.Email || request.Password != user.Password {
			return fiber.ErrUnauthorized
		}

		// Create token
		token := jwt.New(jwt.SigningMethodHS256)

		// Set claims
		claims := token.Claims.(jwt.MapClaims)
		claims["email"] = user.Email
		claims["role"] = "admin" // example role
		claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

		// Generate encoded token
		t, err := token.SignedString([]byte(secretKey))
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		return c.JSON(fiber.Map{"token": t})
	}
}

// extractUserFromJWT is a middleware that extracts user data from the JWT token
func extractUserFromJWT(c *fiber.Ctx) error {
	user := &UserData{}

	// Extract the token from the Fiber context (inserted by the JWT middleware)
	token := c.Locals("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)

	fmt.Println(claims)

	user.Email = claims["email"].(string)
	user.Role = claims["role"].(string)

	// Store the user data in the Fiber context
	c.Locals(userContextKey, user)

	return c.Next()
}

// isAdmin checks if the user is an admin
func isAdmin(c *fiber.Ctx) error {
	user := c.Locals(userContextKey).(*UserData)
  
	if user.Role != "admin" {
	  return fiber.ErrUnauthorized
	}
  
	return c.Next()
  }
