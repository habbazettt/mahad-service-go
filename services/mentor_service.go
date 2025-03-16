package services

import (
	"github.com/gofiber/fiber/v2"
	"github.com/habbazettt/mahad-service-go/dto"
	"github.com/habbazettt/mahad-service-go/models"
	"github.com/habbazettt/mahad-service-go/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type MentorService struct {
	DB *gorm.DB
}

// GetAllMentors - Mengambil semua mentor (Hanya untuk mentor)
func (s *MentorService) GetAllMentors(c *fiber.Ctx) error {
	var mentors []models.Mentor

	if err := s.DB.Preload("Mahasantri").Find(&mentors).Error; err != nil {
		logrus.WithError(err).Error("Failed to fetch mentors")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to fetch mentors", err.Error())
	}

	response := make([]dto.MentorResponse, len(mentors))
	for i, mentor := range mentors {
		// Konversi Mahasantri ke DTO
		mahasantriList := make([]dto.MahasantriResponse, len(mentor.Mahasantri))
		for j, m := range mentor.Mahasantri {
			mahasantriList[j] = dto.MahasantriResponse{
				ID:       m.ID,
				Nama:     m.Nama,
				NIM:      m.NIM,
				Jurusan:  m.Jurusan,
				Gender:   m.Gender,
				MentorID: m.MentorID,
			}
		}

		response[i] = dto.MentorResponse{
			ID:              mentor.ID,
			Nama:            mentor.Nama,
			Email:           mentor.Email,
			Gender:          mentor.Gender,
			MahasantriCount: len(mentor.Mahasantri),
			Mahasantri:      mahasantriList, // 🔥 Tambahkan ini
		}
	}

	logrus.Info("Mentors retrieved successfully")
	return utils.SuccessResponse(c, fiber.StatusOK, "All mentors retrieved successfully", response)
}

// GetMentorByID - Mengambil mentor berdasarkan ID
func (s *MentorService) GetMentorByID(c *fiber.Ctx) error {
	id := c.Params("id")
	var mentor models.Mentor

	// 🔥 Pastikan kita memuat Mahasantri terkait dengan Preload
	if err := s.DB.Preload("Mahasantri").First(&mentor, id).Error; err != nil {
		logrus.WithError(err).Warn("Mentor not found")
		return utils.ResponseError(c, fiber.StatusNotFound, "Mentor not found", nil)
	}

	// Konversi Mahasantri ke DTO
	mahasantriList := make([]dto.MahasantriResponse, len(mentor.Mahasantri))
	for j, m := range mentor.Mahasantri {
		mahasantriList[j] = dto.MahasantriResponse{
			ID:       m.ID,
			Nama:     m.Nama,
			NIM:      m.NIM,
			Jurusan:  m.Jurusan,
			Gender:   m.Gender,
			MentorID: m.MentorID,
		}
	}

	// Mapping ke DTO Response
	response := dto.MentorResponse{
		ID:              mentor.ID,
		Nama:            mentor.Nama,
		Email:           mentor.Email,
		Gender:          mentor.Gender,
		MahasantriCount: len(mentor.Mahasantri),
		Mahasantri:      mahasantriList, // 🔥 Tambahkan ini
	}

	logrus.WithFields(logrus.Fields{
		"mentor_id": mentor.ID,
	}).Info("Mentor retrieved successfully")

	return utils.SuccessResponse(c, fiber.StatusOK, "Successfully retrieved mentor by ID", response)
}

// UpdateMentor - Memperbarui data mentor berdasarkan ID (Hanya untuk mentor)
func (s *MentorService) UpdateMentor(c *fiber.Ctx) error {
	id := c.Params("id")

	// Ambil data mentor dari database
	var mentor models.Mentor
	if err := s.DB.Preload("Mahasantri").First(&mentor, id).Error; err != nil {
		logrus.WithError(err).Warn("Mentor not found")
		return utils.ResponseError(c, fiber.StatusNotFound, "Mentor not found", nil)
	}

	// Bind request body ke DTO
	var updateRequest dto.UpdateMentorRequest
	if err := c.BodyParser(&updateRequest); err != nil {
		logrus.WithError(err).Error("Failed to parse request body")
		return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	// Validasi email unik jika diupdate
	if updateRequest.Email != nil {
		var existingMentor models.Mentor
		if err := s.DB.Where("email = ? AND id != ?", *updateRequest.Email, id).First(&existingMentor).Error; err == nil {
			logrus.Warn("Email already in use by another mentor")
			return utils.ResponseError(c, fiber.StatusConflict, "Email already in use", nil)
		}

		// Validasi format email
		if !utils.IsValidEmail(*updateRequest.Email) {
			return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid email format", nil)
		}
	}

	// Cek apakah ada perubahan data
	updated := false
	if updateRequest.Nama != nil && mentor.Nama != *updateRequest.Nama {
		mentor.Nama = *updateRequest.Nama
		updated = true
	}
	if updateRequest.Email != nil && mentor.Email != *updateRequest.Email {
		mentor.Email = *updateRequest.Email
		updated = true
	}
	if updateRequest.Gender != nil && mentor.Gender != *updateRequest.Gender {
		mentor.Gender = *updateRequest.Gender
		updated = true
	}

	// Jika tidak ada perubahan, return langsung
	if !updated {
		return utils.ResponseError(c, fiber.StatusBadRequest, "No changes detected", nil)
	}

	// Simpan perubahan ke database
	if err := s.DB.Save(&mentor).Error; err != nil {
		logrus.WithError(err).Error("Failed to update mentor")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to update mentor", err.Error())
	}

	if err := s.DB.Preload("Mahasantri").First(&mentor, id).Error; err != nil {
		logrus.WithError(err).Error("Failed to fetch updated mentor")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to fetch updated mentor", err.Error())
	}

	// Konversi Mahasantri ke DTO
	mahasantriList := make([]dto.MahasantriResponse, len(mentor.Mahasantri))
	for j, m := range mentor.Mahasantri {
		mahasantriList[j] = dto.MahasantriResponse{
			ID:       m.ID,
			Nama:     m.Nama,
			NIM:      m.NIM,
			Jurusan:  m.Jurusan,
			Gender:   m.Gender,
			MentorID: m.MentorID,
		}
	}

	// Buat Response DTO
	response := dto.MentorResponse{
		ID:              mentor.ID,
		Nama:            mentor.Nama,
		Email:           mentor.Email,
		Gender:          mentor.Gender,
		MahasantriCount: len(mentor.Mahasantri),
		Mahasantri:      mahasantriList,
	}

	logrus.WithFields(logrus.Fields{
		"mentor_id": mentor.ID,
		"updated":   updated,
	}).Info("Mentor updated successfully")

	return utils.SuccessResponse(c, fiber.StatusOK, "Mentor updated successfully", response)
}

// DeleteMentor - Menghapus mentor berdasarkan ID (Hanya untuk mentor)
func (s *MentorService) DeleteMentor(c *fiber.Ctx) error {
	id := c.Params("id")
	var mentor models.Mentor

	if err := s.DB.First(&mentor, id).Error; err != nil {
		logrus.WithError(err).Warn("Mentor not found")
		return utils.ResponseError(c, fiber.StatusNotFound, "Mentor not found", nil)
	}

	s.DB.Delete(&mentor)
	logrus.WithFields(logrus.Fields{
		"mentor_id": mentor.ID,
	}).Info("Mentor deleted successfully")

	return utils.SuccessResponse(c, fiber.StatusOK, "Mentor deleted successfully", nil)
}
