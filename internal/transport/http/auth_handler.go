package http

import (
	"famloop/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authUC    *usecase.AuthUseCase
	avatarUC  *usecase.AvatarUseCase
	uploadDir string
}

func NewAuthHandler(authUC *usecase.AuthUseCase, avatarUC *usecase.AvatarUseCase, uploadDir string) *AuthHandler {
	return &AuthHandler{
		authUC:    authUC,
		avatarUC:  avatarUC,
		uploadDir: uploadDir,
	}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var input usecase.RegisterInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "Payload JSON pendaftaran tidak valid",
			},
		})
	}

	result, err := h.authUC.Register(c.Context(), input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": err.Error(),
			},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Pendaftaran berhasil",
		"data":    result,
	})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var input usecase.LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "Payload JSON login tidak valid",
			},
		})
	}

	result, err := h.authUC.Login(c.Context(), input)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusUnauthorized,
				"message": err.Error(),
			},
		})
	}

	return c.JSON(fiber.Map{
		"message": "Login berhasil",
		"data":    result,
	})
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{"code": fiber.StatusUnauthorized, "message": "Unauthorized"},
		})
	}

	user, err := h.authUC.GetUserProfile(c.Context(), userID)
	if err != nil || user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{"code": fiber.StatusNotFound, "message": "Pengguna tidak ditemukan"},
		})
	}

	return c.JSON(fiber.Map{
		"data": user,
	})
}

func (h *AuthHandler) UpdateProfile(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{"code": fiber.StatusUnauthorized, "message": "Unauthorized"},
		})
	}

	var payload struct {
		Username    string `json:"username"`
		Email       string `json:"email"`
		PhoneNumber string `json:"phone_number"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{"code": fiber.StatusBadRequest, "message": "Payload JSON tidak valid"},
		})
	}

	updated, err := h.authUC.UpdateProfile(c.Context(), userID, payload.Username, payload.Email, payload.PhoneNumber)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{"code": fiber.StatusBadRequest, "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"message": "Profil berhasil diperbarui",
		"data":    updated,
	})
}

func (h *AuthHandler) UploadAvatar(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{"code": fiber.StatusUnauthorized, "message": "Unauthorized"},
		})
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{"code": fiber.StatusBadRequest, "message": "File foto profil 'avatar' diperlukan"},
		})
	}

	src, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{"code": fiber.StatusInternalServerError, "message": "Gagal membuka file avatar"},
		})
	}
	defer src.Close()

	avatarURL, err := h.avatarUC.UploadUserAvatar(c.Context(), userID, src, file.Filename, h.uploadDir)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{"code": fiber.StatusBadRequest, "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"message": "Foto profil berhasil diperbarui",
		"data": fiber.Map{
			"avatar_url": avatarURL,
		},
	})
}
