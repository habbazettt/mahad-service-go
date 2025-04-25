package services

import (
	"math"
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

// GetAllMahasantri - Mengambil semua mahasantri dengan pagination (Hanya untuk mentor)
// @Summary Get All Mahasantri
// @Description Get a list of all Mahasantri with pagination, only accessible by mentor
// @Tags Mahasantri
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Limit per page" default(10)
// @Param name query string false "Search by name"
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=[]dto.MahasantriResponse,pagination=utils.Pagination} "Mahasantri retrieved successfully"
// @Failure 400 {object} utils.Response "Invalid request"
// @Failure 500 {object} utils.Response "Internal server error"
// @Router /api/v1/mahasantri [get]
func (s *MahasantriService) GetAllMahasantri(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	name := c.Query("nama", "") // Ambil parameter pencarian nama

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	var totalMahasantri int64
	var mahasantri []models.Mahasantri

	// Hitung total Mahasantri untuk pagination dengan filter nama
	query := s.DB.Model(&models.Mahasantri{})
	if name != "" {
		query = query.Where("nama ILIKE ?", "%"+name+"%") // Filter berdasarkan nama
	}
	query.Count(&totalMahasantri)

	// Preload Mentor jika ingin menampilkan informasi mentornya juga, dan paginate
	if err := query.Preload("Mentor").
		Limit(limit).Offset(offset).
		Find(&mahasantri).Error; err != nil {
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

	logrus.WithFields(logrus.Fields{
		"page":  page,
		"limit": limit,
		"name":  name,
	}).Info("Paginated mahasantri retrieved successfully")

	// Return response dengan pagination informasi
	return utils.SuccessResponse(c, fiber.StatusOK, "Mahasantri retrieved successfully", fiber.Map{
		"pagination": fiber.Map{
			"current_page": page,
			"total_data":   totalMahasantri,
			"total_pages":  int(math.Ceil(float64(totalMahasantri) / float64(limit))),
		},
		"mahasantri": response,
	})
}

// GetMahasantriByID - Mengambil mahasantri berdasarkan ID
// @Summary Mengambil mahasantri berdasarkan ID
// @Description Mendapatkan data mahasantri dengan mencocokkan ID yang diberikan.
// @Tags Mahasantri
// @Accept json
// @Produce json
// @Param id path int true "ID Mahasantri"
// @Success 200 {object} dto.MahasantriResponse "Mahasantri ditemukan"
// @Failure 400 {object} utils.Response "Invalid ID format"
// @Failure 404 {object} utils.Response "Mahasantri not found"
// @Security BearerAuth
// @Router /api/v1/mahasantri/{id} [get]
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
// @Summary Mengambil semua mahasantri berdasarkan mentor_id
// @Description Mengambil data mahasantri yang memiliki mentor_id yang sesuai dengan parameter mentor_id.
// @Tags Mahasantri
// @Accept json
// @Produce json
// @Param mentor_id path int true "ID Mentor"
// @Success 200 {array} dto.MahasantriResponse "List of Mahasantri"
// @Failure 400 {object} utils.Response "Invalid mentor ID format"
// @Failure 500 {object} utils.Response "Failed to fetch mahasantri for mentor"
// @Security BearerAuth
// @Router /api/v1/mahasantri/mentor/{mentor_id} [get]
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
// @Summary Memperbarui data mahasantri berdasarkan ID
// @Description Memperbarui data mahasantri seperti nama, NIM, jurusan, dan gender berdasarkan ID. Hanya dapat diakses oleh mentor.
// @Tags Mahasantri
// @Accept json
// @Produce json
// @Param id path int true "Mahasantri ID"
// @Param updateMahasantriRequest body dto.UpdateMahasantriRequest true "Data yang ingin diperbarui"
// @Success 200 {object} dto.MahasantriResponse "Mahasantri updated successfully"
// @Failure 400 {object} utils.Response "Invalid request body or No changes detected"
// @Failure 404 {object} utils.Response "Mahasantri not found"
// @Failure 500 {object} utils.Response "Failed to update mahasantri"
// @Security BearerAuth
// @Router /api/v1/mahasantri/{id} [put]
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
// @Summary Menghapus mahasantri berdasarkan ID
// @Description Menghapus data mahasantri berdasarkan ID. Hanya dapat diakses oleh mentor.
// @Tags Mahasantri
// @Accept json
// @Produce json
// @Param id path int true "Mahasantri ID"
// @Success 200 {object} utils.Response "Mahasantri deleted successfully"
// @Failure 404 {object} utils.Response "Mahasantri not found"
// @Failure 500 {object} utils.Response "Failed to delete mahasantri"
// @Security BearerAuth
// @Router /api/v1/mahasantri/{id} [delete]
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
