package services

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/habbazettt/mahad-service-go/dto"
	"github.com/habbazettt/mahad-service-go/models"
	"github.com/habbazettt/mahad-service-go/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type MahasantriService struct {
	DB *gorm.DB
}

// GetAllMahasantri - Mengambil semua mahasantri (Hanya untuk mentor)
func (s *MahasantriService) GetAllMahasantri(c *fiber.Ctx) error {
	var mahasantri []models.Mahasantri

	// Preload Mentor jika ingin menampilkan informasi mentornya juga
	if err := s.DB.Preload("Mentor").Find(&mahasantri).Error; err != nil {
		logrus.WithError(err).Error("Failed to fetch mahasantri")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to fetch mahasantri", err.Error())
	}

	response := make([]dto.MahasantriResponse, len(mahasantri))
	for i, m := range mahasantri {
		response[i] = dto.MahasantriResponse{
			ID:       m.ID,
			Nama:     m.Nama,
			NIM:      m.NIM,
			Jurusan:  m.Jurusan,
			Gender:   m.Gender,
			MentorID: m.MentorID,
		}
	}

	logrus.Info("Mahasantri retrieved successfully")
	return utils.SuccessResponse(c, fiber.StatusOK, "Mahasantri retrieved successfully", response)
}

// GetMahasantriByID - Mengambil mahasantri berdasarkan ID
func (s *MahasantriService) GetMahasantriByID(c *fiber.Ctx) error {
	id := c.Params("id")

	// Cek apakah ID valid (harus angka)
	if _, err := strconv.Atoi(id); err != nil {
		return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid ID format", nil)
	}

	var mahasantri models.Mahasantri
	if err := s.DB.Preload("Mentor").First(&mahasantri, id).Error; err != nil {
		logrus.WithError(err).Warn("Mahasantri not found")
		return utils.ResponseError(c, fiber.StatusNotFound, "Mahasantri not found", nil)
	}

	response := dto.MahasantriResponse{
		ID:       mahasantri.ID,
		Nama:     mahasantri.Nama,
		NIM:      mahasantri.NIM,
		Jurusan:  mahasantri.Jurusan,
		Gender:   mahasantri.Gender,
		MentorID: mahasantri.MentorID,
	}

	logrus.WithFields(logrus.Fields{
		"mahasantri_id": mahasantri.ID,
	}).Info("Mahasantri retrieved successfully")

	return utils.SuccessResponse(c, fiber.StatusOK, "Mahasantri found", response)
}

// GetMahasantriByMentorID - Mengambil semua mahasantri berdasarkan mentor_id (Hanya untuk mentor)
func (s *MahasantriService) GetMahasantriByMentorID(c *fiber.Ctx) error {
	mentorID := c.Params("mentor_id")

	// Cek apakah mentor_id valid
	if _, err := strconv.Atoi(mentorID); err != nil {
		return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid mentor ID format", nil)
	}

	var mahasantri []models.Mahasantri
	if err := s.DB.Where("mentor_id = ?", mentorID).Find(&mahasantri).Error; err != nil {
		logrus.WithError(err).Error("Failed to fetch mahasantri for mentor")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to fetch mahasantri for mentor", err.Error())
	}

	response := make([]dto.MahasantriResponse, len(mahasantri))
	for i, m := range mahasantri {
		response[i] = dto.MahasantriResponse{
			ID:       m.ID,
			Nama:     m.Nama,
			NIM:      m.NIM,
			Jurusan:  m.Jurusan,
			Gender:   m.Gender,
			MentorID: m.MentorID,
		}
	}

	logrus.WithFields(logrus.Fields{
		"mentor_id": mentorID,
	}).Info("Mahasantri retrieved successfully for mentor")

	return utils.SuccessResponse(c, fiber.StatusOK, "Mahasantri retrieved successfully", response)
}

// UpdateMahasantri - Memperbarui data mahasantri berdasarkan ID (Hanya untuk mentor)
func (s *MahasantriService) UpdateMahasantri(c *fiber.Ctx) error {
	id := c.Params("id")
	var mahasantri models.Mahasantri

	if err := s.DB.First(&mahasantri, id).Error; err != nil {
		logrus.WithError(err).Warn("Mahasantri not found")
		return utils.ResponseError(c, fiber.StatusNotFound, "Mahasantri not found", nil)
	}

	// Bind request body ke DTO
	var updateRequest dto.UpdateMahasantriRequest
	if err := c.BodyParser(&updateRequest); err != nil {
		logrus.WithError(err).Error("Failed to parse request body")
		return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	updated := false
	if updateRequest.Nama != nil && *updateRequest.Nama != mahasantri.Nama {
		mahasantri.Nama = *updateRequest.Nama
		updated = true
	}
	if updateRequest.NIM != nil && *updateRequest.NIM != mahasantri.NIM {
		mahasantri.NIM = *updateRequest.NIM
		updated = true
	}
	if updateRequest.Jurusan != nil && *updateRequest.Jurusan != mahasantri.Jurusan {
		mahasantri.Jurusan = *updateRequest.Jurusan
		updated = true
	}
	if updateRequest.Gender != nil && *updateRequest.Gender != mahasantri.Gender {
		mahasantri.Gender = *updateRequest.Gender
		updated = true
	}

	// Jika tidak ada perubahan, langsung return
	if !updated {
		return utils.ResponseError(c, fiber.StatusBadRequest, "No changes detected", nil)
	}

	// Simpan perubahan ke database
	if err := s.DB.Save(&mahasantri).Error; err != nil {
		logrus.WithError(err).Error("Failed to update mahasantri")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to update mahasantri", err.Error())
	}

	logrus.WithFields(logrus.Fields{
		"mahasantri_id": mahasantri.ID,
	}).Info("Mahasantri updated successfully")

	// Gunakan DTO agar field `mentor` tidak ditampilkan
	response := dto.MahasantriResponse{
		ID:       mahasantri.ID,
		Nama:     mahasantri.Nama,
		NIM:      mahasantri.NIM,
		Jurusan:  mahasantri.Jurusan,
		Gender:   mahasantri.Gender,
		MentorID: mahasantri.MentorID,
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Mahasantri updated successfully", response)
}

// DeleteMahasantri - Menghapus mahasantri berdasarkan ID (Hanya untuk mentor)
func (s *MahasantriService) DeleteMahasantri(c *fiber.Ctx) error {
	id := c.Params("id")
	var mahasantri models.Mahasantri

	if err := s.DB.First(&mahasantri, id).Error; err != nil {
		logrus.WithError(err).Warn("Mahasantri not found")
		return utils.ResponseError(c, fiber.StatusNotFound, "Mahasantri not found", nil)
	}

	s.DB.Delete(&mahasantri)
	logrus.WithFields(logrus.Fields{
		"mahasantri_id": mahasantri.ID,
	}).Info("Mahasantri deleted successfully")

	return utils.SuccessResponse(c, fiber.StatusOK, "Mahasantri deleted successfully", nil)
}
