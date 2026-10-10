package http

import (
	"famloop/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

type GroupHandler struct {
	groupUC   *usecase.GroupUseCase
	avatarUC  *usecase.AvatarUseCase
	uploadDir string
}

func NewGroupHandler(groupUC *usecase.GroupUseCase, avatarUC *usecase.AvatarUseCase, uploadDir string) *GroupHandler {
	return &GroupHandler{
		groupUC:   groupUC,
		avatarUC:  avatarUC,
		uploadDir: uploadDir,
	}
}

func (h *GroupHandler) CreateGroup(c *fiber.Ctx) error {
	parentID := c.Locals("user_id").(string)

	var payload struct {
		Name string `json:"name"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{"code": fiber.StatusBadRequest, "message": "Payload JSON tidak valid"},
		})
	}

	group, err := h.groupUC.CreateGroup(c.Context(), parentID, payload.Name)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{"code": fiber.StatusBadRequest, "message": err.Error()},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Grup keluarga berhasil dibuat",
		"data":    group,
	})
}

func (h *GroupHandler) GetGroups(c *fiber.Ctx) error {
	parentID := c.Locals("user_id").(string)

	groups, err := h.groupUC.GetParentGroups(c.Context(), parentID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{"code": fiber.StatusInternalServerError, "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"data": groups,
	})
}

func (h *GroupHandler) GetGroup(c *fiber.Ctx) error {
	parentID := c.Locals("user_id").(string)
	groupID := c.Params("id")

	group, devices, err := h.groupUC.GetGroupDetails(c.Context(), parentID, groupID)
	if err != nil {
		if err == usecase.ErrGroupNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": fiber.Map{"code": fiber.StatusNotFound, "message": err.Error()},
			})
		}
		if err == usecase.ErrGroupUnauthorized {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": fiber.Map{"code": fiber.StatusForbidden, "message": err.Error()},
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{"code": fiber.StatusInternalServerError, "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"group":   group,
			"members": devices,
		},
	})
}

func (h *GroupHandler) PairDevice(c *fiber.Ctx) error {
	parentID := c.Locals("user_id").(string)
	groupID := c.Params("id")

	var payload usecase.QRScanPayload
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{"code": fiber.StatusBadRequest, "message": "Payload hasil scan QR tidak valid"},
		})
	}

	device, err := h.groupUC.PairDeviceFromQR(c.Context(), parentID, groupID, payload)
	if err != nil {
		status := fiber.StatusBadRequest
		if err == usecase.ErrGroupNotFound {
			status = fiber.StatusNotFound
		} else if err == usecase.ErrGroupUnauthorized {
			status = fiber.StatusForbidden
		}
		return c.Status(status).JSON(fiber.Map{
			"error": fiber.Map{"code": status, "message": err.Error()},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Perangkat anggota keluarga berhasil dihubungkan",
		"data":    device,
	})
}

func (h *GroupHandler) UploadDeviceAvatar(c *fiber.Ctx) error {
	deviceID := c.Params("id")
	groupID := c.Query("group_id")

	if deviceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{"code": fiber.StatusBadRequest, "message": "device id parameter diperlukan"},
		})
	}

	// Ambil groupID dari device jika tidak disertakan di query
	if groupID == "" {
		if dev, err := h.groupUC.GetDevice(c.Context(), deviceID); err == nil && dev != nil {
			groupID = dev.GroupID
		}
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

	avatarURL, err := h.avatarUC.UploadDeviceAvatar(c.Context(), deviceID, groupID, src, file.Filename, h.uploadDir)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{"code": fiber.StatusBadRequest, "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"message": "Foto profil anggota berhasil diperbarui",
		"data": fiber.Map{
			"device_id":  deviceID,
			"avatar_url": avatarURL,
		},
	})
}

func (h *GroupHandler) DeleteGroup(c *fiber.Ctx) error {
	parentID := c.Locals("user_id").(string)
	groupID := c.Params("id")

	if err := h.groupUC.DeleteGroup(c.Context(), parentID, groupID); err != nil {
		status := fiber.StatusBadRequest
		if err == usecase.ErrGroupNotFound {
			status = fiber.StatusNotFound
		} else if err == usecase.ErrGroupUnauthorized {
			status = fiber.StatusForbidden
		}
		return c.Status(status).JSON(fiber.Map{
			"error": fiber.Map{"code": status, "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"message": "Grup keluarga berhasil dihapus",
	})
}

func (h *GroupHandler) RemoveMember(c *fiber.Ctx) error {
	parentID := c.Locals("user_id").(string)
	groupID := c.Params("id")
	deviceID := c.Params("device_id")

	if err := h.groupUC.RemoveMemberDevice(c.Context(), parentID, groupID, deviceID); err != nil {
		status := fiber.StatusBadRequest
		if err == usecase.ErrGroupNotFound {
			status = fiber.StatusNotFound
		} else if err == usecase.ErrGroupUnauthorized {
			status = fiber.StatusForbidden
		}
		return c.Status(status).JSON(fiber.Map{
			"error": fiber.Map{"code": status, "message": err.Error()},
		})
	}

	return c.JSON(fiber.Map{
		"message": "Anggota keluarga berhasil dihapus dari grup",
	})
}

func (h *GroupHandler) GetDevice(c *fiber.Ctx) error {
	deviceID := c.Params("id")
	dev, err := h.groupUC.GetDevice(c.Context(), deviceID)
	if err != nil || dev == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{"code": fiber.StatusNotFound, "message": "Perangkat tidak ditemukan"},
		})
	}

	return c.JSON(fiber.Map{
		"data": dev,
	})
}
