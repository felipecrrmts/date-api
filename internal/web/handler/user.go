package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/muzzapp/date-api/internal/users"
)

type UserHandler struct {
	service Users
	secret  string
}

func NewUserHandler(secret string, service Users) *UserHandler {
	return &UserHandler{service: service, secret: secret}
}

func (h *UserHandler) Login() fiber.Handler {
	return func(c *fiber.Ctx) error {
		r := new(LoginRequest)
		if err := c.BodyParser(r); err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}
		if r.Email == "" || r.Password == "" {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		user, err := h.service.Login(c.Context(), r.Email, r.Password)
		switch {
		case errors.Is(err, users.ErrUserNotFound), errors.Is(err, users.ErrPasswordMismatch):
			return c.SendStatus(fiber.StatusBadRequest)
		case err != nil:
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		token, err := h.generateToken(user.ID, user.Name)
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		return c.JSON(fiber.Map{"token": token})
	}
}

func (h *UserHandler) generateToken(ID int32, name string) (string, error) {
	claims := jwt.MapClaims{
		"id":   ID,
		"name": name,
		"exp":  time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(h.secret))
}

func (h *UserHandler) CreateUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, err := h.service.CreateUser(c.Context())
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		return c.JSON(toCreateUserResponse(user), fiber.MIMEApplicationJSON)
	}
}

func (h *UserHandler) Discover() fiber.Handler {
	return func(c *fiber.Ctx) error {
		r := new(DiscoverRequest)
		if err := c.QueryParser(r); err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}
		r.validate()

		requesterID := userIDFromToken(c)
		profiles, err := h.service.Discover(c.Context(), requesterID, r.MinAge, r.MaxAge, r.Gender, r.Ranked)
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		return c.JSON(toDiscoverResponse(profiles))
	}
}

func (h *UserHandler) Swipe() fiber.Handler {
	return func(c *fiber.Ctx) error {
		r := new(SwipeRequest)
		if err := c.BodyParser(r); err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}
		requesterID := userIDFromToken(c)
		ok, err := h.service.Swipe(c.Context(), requesterID, r.SwipedID, r.Ok)
		switch {
		case errors.Is(err, users.ErrUserNotFound):
			return c.SendStatus(fiber.StatusBadRequest)
		case err != nil:
			return c.SendStatus(fiber.StatusInternalServerError)
		case ok:
			return c.JSON(&SwipeResponse{Swipe: Swipe{
				Matched:   true,
				MatchedID: r.SwipedID,
			}})
		default:
			return c.JSON(&SwipeResponse{Swipe: Swipe{
				Matched: false,
			}})
		}
	}
}

func userIDFromToken(c *fiber.Ctx) int32 {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	return int32(claims["id"].(float64))
}
