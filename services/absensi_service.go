package services

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/habbazettt/mahad-service-go/dto"
	"github.com/habbazettt/mahad-service-go/models"
	"github.com/habbazettt/mahad-service-go/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type AbsensiService struct {
	DB *gorm.DB
}

// CreateAbsensi - Membuat absensi baru
// @Summary Membuat absensi baru untuk Mahasantri
// @Description Endpoint ini digunakan untuk membuat absensi baru untuk Mahasantri berdasarkan data yang dikirimkan oleh mentor.
// @Tags Absensi
// @Accept json
// @Produce json
// @Param request body dto.AbsensiRequestDTO true "Data Absensi"
// @Success 201 {object} utils.Response "Absensi created successfully"
// @Failure 400 {object} utils.Response "Invalid request body or Absensi already recorded for this date and time"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Failure 404 {object} utils.Response "Mahasantri not found"
// @Failure 500 {object} utils.Response "Failed to create absensi"
// @Security BearerAuth
// @Router /api/v1/absensi [post]
func (s *AbsensiService) CreateAbsensi(c *fiber.Ctx) error {
	var req dto.AbsensiRequestDTO

	if err := c.BodyParser(&req); err != nil {
		logrus.WithError(err).Error("Failed to parse request body")
		return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	token := c.Get("Authorization")
	if token == "" {
		logrus.Error("Authorization token is missing")
		return utils.ResponseError(c, fiber.StatusUnauthorized, "Unauthorized", "Authorization token is missing")
	}

	var mahasantri models.Mahasantri
	if err := s.DB.First(&mahasantri, req.MahasantriID).Error; err != nil {
		logrus.WithError(err).Error("Mahasantri not found")
		return utils.ResponseError(c, fiber.StatusNotFound, "Mahasantri not found", err.Error())
	}

	// Parsing tanggal dari string ke time.Time
	tanggal, err := time.Parse("02-01-2006", req.Tanggal)
	if err != nil {
		logrus.WithError(err).Error("Invalid date format")
		return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid date format", err.Error())
	}

	// Memeriksa apakah hari adalah Sabtu atau Minggu
	if tanggal.Weekday() == time.Saturday || tanggal.Weekday() == time.Sunday {
		logrus.Warn("Absensi tidak diperbolehkan pada hari Sabtu atau Minggu")
		return utils.ResponseError(c, fiber.StatusBadRequest, "Absensi is not allowed on Saturdays or Sundays", "Absensi tidak diperbolehkan pada hari Sabtu atau Minggu")
	}

	// Memeriksa apakah absensi sudah tercatat untuk tanggal dan waktu yang diinput
	var existingAbsensi models.Absensi
	if err := s.DB.Where("mahasantri_id = ? AND tanggal = ? AND waktu = ?", req.MahasantriID, tanggal, req.Waktu).First(&existingAbsensi).Error; err == nil {
		logrus.Warn("Absensi sudah tercatat untuk tanggal dan waktu ini")
		return utils.ResponseError(c, fiber.StatusBadRequest, "Absensi already recorded for this date and time", "Absensi sudah tercatat untuk tanggal dan waktu ini")
	}

	absensi := models.Absensi{
		MahasantriID: req.MahasantriID,
		MentorID:     req.MentorID,
		Waktu:        req.Waktu,
		Status:       req.Status,
		Tanggal:      tanggal,
	}

	if err := s.DB.Create(&absensi).Error; err != nil {
		logrus.WithError(err).Error("Failed to create absensi")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to create absensi", err.Error())
	}

	var absensiWithRelations models.Absensi
	if err := s.DB.Preload("Mentor").Preload("Mahasantri").First(&absensiWithRelations, absensi.ID).Error; err != nil {
		logrus.WithError(err).Error("Failed to preload relations")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to preload relations", err.Error())
	}

	response := dto.AbsensiResponseDTO{
		ID:           absensiWithRelations.ID,
		MahasantriID: absensiWithRelations.MahasantriID,
		MentorID:     absensiWithRelations.MentorID,
		Waktu:        absensiWithRelations.Waktu,
		Status:       absensiWithRelations.Status,
		Tanggal:      absensiWithRelations.GetFormattedTanggal(),
		CreatedAt:    absensiWithRelations.CreatedAt,
		UpdatedAt:    absensiWithRelations.UpdatedAt,
		Mentor: dto.MentorResponseDTO{
			ID:     absensiWithRelations.Mentor.ID,
			Nama:   absensiWithRelations.Mentor.Nama,
			Email:  absensiWithRelations.Mentor.Email,
			Gender: absensiWithRelations.Mentor.Gender,
		},
		Mahasantri: dto.MahasantriResponseDTO{
			ID:      absensiWithRelations.Mahasantri.ID,
			Nama:    absensiWithRelations.Mahasantri.Nama,
			NIM:     absensiWithRelations.Mahasantri.NIM,
			Jurusan: absensiWithRelations.Mahasantri.Jurusan,
			Gender:  absensiWithRelations.Mahasantri.Gender,
		},
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Absensi created successfully", response)
}

// GetAbsensi - Mengambil data absensi
// @Summary Mengambil data absensi
// @Description Endpoint ini digunakan untuk mengambil data absensi dengan pagination, filter, dan sorting.
// @Tags Absensi
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Nomor halaman" default(1)
// @Param limit query int false "Jumlah data per halaman" default(10)
// @Param month query string false "Filter berdasarkan bulan (MM)"
// @Param status query string false "Filter berdasarkan status"
// @Param waktu query string false "Filter berdasarkan waktu"
// @Param mahasantri_id query int false "Filter berdasarkan ID Mahasantri"
// @Param mentor_id query int false "Filter berdasarkan ID Mentor"
// @Param tanggal query string false "Filter berdasarkan tanggal (DD-MM-YYYY)"
// @Param sort query string false "Sort by created_at" Enums(asc, desc) Default(desc)
// @Success 200 {object} utils.Response "Data absensi retrieved successfully"
// @Failure 400 {object} utils.Response "Invalid request"
// @Failure 500 {object} utils.Response "Failed to retrieve absensi"
// @Router /api/v1/absensi [get]
func (s *AbsensiService) GetAbsensi(c *fiber.Ctx) error {
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	month := c.Query("month")
	status := c.Query("status")
	waktu := c.Query("waktu")
	mahasantriID := c.Query("mahasantri_id")
	mentorID := c.Query("mentor_id")
	tanggal := c.Query("tanggal")
	sort := c.Query("sort", "desc")

	var absensi []models.Absensi
	var total int64

	// Build query
	query := s.DB.Model(&models.Absensi{})

	// Apply filters
	if month != "" {
		query = query.Where("EXTRACT(MONTH FROM tanggal) = ?", month)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if waktu != "" {
		query = query.Where("waktu = ?", waktu)
	}
	if mahasantriID != "" {
		query = query.Where("mahasantri_id = ?", mahasantriID)
	}
	if mentorID != "" {
		query = query.Where("mentor_id = ?", mentorID)
	}
	if tanggal != "" {
		query = query.Where("tanggal = ?", tanggal)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		logrus.WithError(err).Error("Failed to count absensi")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to retrieve absensi", err.Error())
	}

	orderDirection := "desc"
	if sort == "asc" {
		orderDirection = "asc"
	}

	// Apply pagination and preload relations
	if err := query.Offset((page - 1) * limit).
		Limit(limit).
		Preload("Mentor").
		Preload("Mahasantri").
		Order("created_at " + orderDirection).
		Find(&absensi).Error; err != nil {
		logrus.WithError(err).Error("Failed to retrieve absensi")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to retrieve absensi", err.Error())
	}

	// Prepare response DTOs
	responseAbsensi := make([]dto.AbsensiResponseDTO, len(absensi))
	for i, a := range absensi {
		responseAbsensi[i] = dto.AbsensiResponseDTO{
			ID:           a.ID,
			MahasantriID: a.MahasantriID,
			MentorID:     a.MentorID,
			Waktu:        a.Waktu,
			Status:       a.Status,
			Tanggal:      a.GetFormattedTanggal(),
			CreatedAt:    a.CreatedAt,
			UpdatedAt:    a.UpdatedAt,
			Mentor: dto.MentorResponseDTO{
				ID:     a.Mentor.ID,
				Nama:   a.Mentor.Nama,
				Email:  a.Mentor.Email,
				Gender: a.Mentor.Gender,
			},
			Mahasantri: dto.MahasantriResponseDTO{
				ID:      a.Mahasantri.ID,
				Nama:    a.Mahasantri.Nama,
				NIM:     a.Mahasantri.NIM,
				Jurusan: a.Mahasantri.Jurusan,
				Gender:  a.Mahasantri.Gender,
			},
		}
	}

	// Pagination info
	pagination := fiber.Map{
		"current_page": page,
		"total_data":   total,
		"total_pages":  int(math.Ceil(float64(total) / float64(limit))),
	}

	response := fiber.Map{
		"absensi":    responseAbsensi,
		"pagination": pagination,
	}

	logrus.WithFields(logrus.Fields{
		"page":  page,
		"limit": limit,
	}).Info("Paginated absensi retrieved successfully")

	return utils.SuccessResponse(c, fiber.StatusOK, "Berhasil mengambil data absensi", response)
}

// GetAbsensiByID - Mengambil data absensi berdasarkan ID
// @Summary Mengambil data absensi berdasarkan ID
// @Description Endpoint ini digunakan untuk mengambil data absensi berdasarkan ID.
// @Tags Absensi
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID Absensi"
// @Success 200 {object} utils.Response "Data absensi retrieved successfully"
// @Failure 404 {object} utils.Response "Absensi not found"
// @Failure 500 {object} utils.Response "Failed to retrieve absensi"
// @Router /api/v1/absensi/{id} [get]
func (s *AbsensiService) GetAbsensiByID(c *fiber.Ctx) error {
	id := c.Params("id")

	var absensi models.Absensi

	// Find absensi by ID and preload relations
	if err := s.DB.Preload("Mentor").Preload("Mahasantri").First(&absensi, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ResponseError(c, fiber.StatusNotFound, "Absensi not found", err.Error())
		}
		logrus.WithError(err).Error("Failed to retrieve absensi")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to retrieve absensi", err.Error())
	}

	// Prepare response DTO
	responseAbsensi := dto.AbsensiResponseDTO{
		ID:           absensi.ID,
		MahasantriID: absensi.MahasantriID,
		MentorID:     absensi.MentorID,
		Waktu:        absensi.Waktu,
		Status:       absensi.Status,
		Tanggal:      absensi.GetFormattedTanggal(),
		CreatedAt:    absensi.CreatedAt,
		UpdatedAt:    absensi.UpdatedAt,
		Mentor: dto.MentorResponseDTO{
			ID:     absensi.Mentor.ID,
			Nama:   absensi.Mentor.Nama,
			Email:  absensi.Mentor.Email,
			Gender: absensi.Mentor.Gender,
		},
		Mahasantri: dto.MahasantriResponseDTO{
			ID:      absensi.Mahasantri.ID,
			Nama:    absensi.Mahasantri.Nama,
			NIM:     absensi.Mahasantri.NIM,
			Jurusan: absensi.Mahasantri.Jurusan,
			Gender:  absensi.Mahasantri.Gender,
		},
	}

	logrus.WithFields(logrus.Fields{
		"id": id,
	}).Info("Absensi retrieved successfully")

	return utils.SuccessResponse(c, fiber.StatusOK, "Berhasil mengambil data absensi", fiber.Map{
		"absensi": responseAbsensi,
	})
}

// GetAbsensiDailySummary godoc
// @Summary Mendapatkan ringkasan absensi harian Mahasantri
// @Description Mengambil data absensi harian Mahasantri selama 1 bulan berdasarkan waktu shubuh dan isya. Data akan mengisi status absen per hari, default "belum-absen" jika belum mengisi.
// @Tags Absensi
// @Security BearerAuth
// @Param mahasantri_id path int true "ID Mahasantri"
// @Param month query string true "Bulan (format: MM, contoh: 04 untuk April)"
// @Param year query string true "Tahun (format: YYYY, contoh: 2025)"
// @Success 200 {object} utils.SuccessResponseSwagger{data=[]dto.AbsensiDailySummaryDTO} "Berhasil mengambil ringkasan absensi harian"
// @Failure 400 {object} utils.ErrorResponseSwagger "Bad Request - Format salah atau parameter tidak lengkap"
// @Failure 500 {object} utils.ErrorResponseSwagger "Internal Server Error - Gagal mengambil data absensi"
// @Router /api/v1/absensi/mahasantri/{mahasantri_id}/daily-summary [get]
func (s *AbsensiService) GetAbsensiDailySummary(c *fiber.Ctx) error {
	id := c.Params("mahasantri_id")
	mahasantriID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid Mahasantri ID format", err.Error())
	}

	month := c.Query("month") // ex: 04
	year := c.Query("year")   // ex: 2025

	if month == "" || year == "" {
		return utils.ResponseError(c, fiber.StatusBadRequest, "Missing query parameters", "month and year are required")
	}

	layout := "02-01-2006"
	location := time.Now().Location()
	startDate, err := time.ParseInLocation("02-01-2006", fmt.Sprintf("01-%s-%s", month, year), location)
	if err != nil {
		return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid date format", err.Error())
	}
	endDate := startDate.AddDate(0, 1, -1)

	// Fetch all absensi for that month
	var absensi []models.Absensi
	if err := s.DB.Where("mahasantri_id = ?", mahasantriID).
		Where("tanggal BETWEEN ? AND ?", startDate, endDate).
		Find(&absensi).Error; err != nil {
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to fetch absensi", err.Error())
	}

	// Fetch Mahasantri details
	var mahasantri models.Mahasantri
	if err := s.DB.Where("id = ?", mahasantriID).First(&mahasantri).Error; err != nil {
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to fetch Mahasantri details", err.Error())
	}

	// Fetch Mentor details
	var mentor models.Mentor
	if err := s.DB.Where("id = ?", mahasantri.MentorID).First(&mentor).Error; err != nil {
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to fetch Mentor details", err.Error())
	}

	// Indexing absensi per tanggal & waktu
	absensiMap := make(map[string]map[string]string) // tanggal -> waktu -> status
	for _, a := range absensi {
		tanggal := a.Tanggal.Format(layout)
		if _, ok := absensiMap[tanggal]; !ok {
			absensiMap[tanggal] = make(map[string]string)
		}
		absensiMap[tanggal][a.Waktu] = a.Status
	}

	// Build daily summary
	var summary []dto.AbsensiDailySummaryDTO
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		tanggal := d.Format(layout)
		shubuh := "belum-absen"
		isya := "belum-absen"

		// Cek apakah hari tersebut adalah Sabtu atau Minggu
		if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
			shubuh = "libur"
			isya = "libur"
		} else {
			if data, ok := absensiMap[tanggal]; ok {
				if val, exists := data["shubuh"]; exists {
					shubuh = val
				}
				if val, exists := data["isya"]; exists {
					isya = val
				}
			}
		}

		// Tambahkan detail hari ke dalam ringkasan
		summary = append(summary, dto.AbsensiDailySummaryDTO{
			Tanggal: tanggal,
			Hari:    getNamaHari(d.Weekday()),
			Shubuh:  shubuh,
			Isya:    isya,
		})
	}

	info := fiber.Map{
		"month": month,
		"year":  year,
	}

	// Membuat response dengan format standar dan pagination info jika diperlukan
	responseData := fiber.Map{
		"mahasantri": fiber.Map{
			"id":      mahasantri.ID,
			"nama":    mahasantri.Nama,
			"nim":     mahasantri.NIM,
			"jurusan": mahasantri.Jurusan,
			"gender":  mahasantri.Gender,
		},
		"mentor": fiber.Map{
			"id":     mentor.ID,
			"nama":   mentor.Nama,
			"email":  mentor.Email,
			"gender": mentor.Gender,
		},
		"daily_summary": summary,
		"info":          info,
	}

	// Return response dengan status 200 OK
	return c.JSON(fiber.Map{
		"status":  true,
		"message": "Daily summary retrieved successfully",
		"data":    responseData,
	})
}

// UpdateAbsensi - Mengupdate data absensi
// @Summary Mengupdate data absensi
// @Description Endpoint ini digunakan untuk mengupdate data absensi berdasarkan ID.
// @Tags Absensi
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID Absensi"
// @Param absensi body dto.UpdateAbsensiRequestDTO true "Data absensi"
// @Success 200 {object} utils.Response "Absensi updated successfully"
// @Failure 400 {object} utils.Response "Invalid request"
// @Failure 404 {object} utils.Response "Absensi not found"
// @Failure 500 {object} utils.Response "Failed to update absensi"
// @Router /api/v1/absensi/{id} [put]
func (s *AbsensiService) UpdateAbsensi(c *fiber.Ctx) error {
	id := c.Params("id")
	var absensi models.Absensi

	// Mencari absensi berdasarkan ID
	if err := s.DB.First(&absensi, id).Error; err != nil {
		logrus.WithField("absensi_id", id).Warn("Absensi not found")
		return utils.ResponseError(c, fiber.StatusNotFound, "Absensi not found", nil)
	}

	var req dto.UpdateAbsensiRequestDTO
	if err := c.BodyParser(&req); err != nil {
		logrus.WithError(err).Error("Invalid request body")
		return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	updated := false
	updateFields := logrus.Fields{"absensi_id": id}

	// Memperbarui field yang diisi
	if req.Waktu != nil && *req.Waktu != absensi.Waktu {
		absensi.Waktu = *req.Waktu
		updateFields["waktu"] = *req.Waktu
		updated = true
	}
	if req.Status != nil && *req.Status != absensi.Status {
		absensi.Status = *req.Status
		updateFields["status"] = *req.Status
		updated = true
	}
	if req.Tanggal != nil {
		tanggal, err := time.Parse("02-01-2006", *req.Tanggal)
		if err != nil {
			logrus.WithError(err).Error("Invalid date format")
			return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid date format", err.Error())
		}
		if tanggal != absensi.Tanggal {
			absensi.Tanggal = tanggal
			updateFields["tanggal"] = tanggal
			updated = true
		}
	}

	if !updated {
		logrus.WithField("absensi_id", id).Warn("No changes detected")
		return utils.ResponseError(c, fiber.StatusBadRequest, "No changes detected", nil)
	}

	// Menyimpan perubahan ke database
	if err := s.DB.Save(&absensi).Error; err != nil {
		logrus.WithError(err).WithFields(updateFields).Error("Failed to update absensi")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to update absensi", err.Error())
	}

	logrus.WithFields(updateFields).Info("Absensi updated successfully")
	response := dto.AbsensiResponseDTO{
		ID:           absensi.ID,
		MahasantriID: absensi.MahasantriID,
		MentorID:     absensi.MentorID,
		Waktu:        absensi.Waktu,
		Status:       absensi.Status,
		Tanggal:      absensi.GetFormattedTanggal(),
		CreatedAt:    absensi.CreatedAt,
		UpdatedAt:    absensi.UpdatedAt,
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Absensi updated successfully", response)
}

// DeleteAbsensi - Menghapus data absensi berdasarkan ID
// @Summary Menghapus data absensi berdasarkan ID
// @Description Endpoint ini digunakan untuk menghapus data absensi berdasarkan ID yang diberikan.
// @Tags Absensi
// @Security BearerAuth
// @Param id path int true "ID Absensi"
// @Success 200 {object} utils.Response "Absensi deleted successfully"
// @Failure 404 {object} utils.Response "Absensi not found"
// @Failure 500 {object} utils.Response "Failed to delete absensi"
// @Router /api/v1/absensi/{id} [delete]
func (s *AbsensiService) DeleteAbsensi(c *fiber.Ctx) error {
	id := c.Params("id")
	var absensi models.Absensi

	// Mencari absensi berdasarkan ID
	if err := s.DB.First(&absensi, id).Error; err != nil {
		logrus.WithField("absensi_id", id).Warn("Absensi not found")
		return utils.ResponseError(c, fiber.StatusNotFound, "Absensi not found", nil)
	}

	// Menghapus absensi dari database
	if err := s.DB.Delete(&absensi).Error; err != nil {
		logrus.WithError(err).WithField("absensi_id", id).Error("Failed to delete absensi")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to delete absensi", err.Error())
	}

	logrus.WithField("absensi_id", id).Info("Absensi deleted successfully")
	return utils.SuccessResponse(c, fiber.StatusOK, "Absensi deleted successfully", nil)
}

// Fungsi untuk mengonversi nama hari ke dalam bahasa Indonesia
func getNamaHari(weekday time.Weekday) string {
	switch weekday {
	case time.Monday:
		return "Senin"
	case time.Tuesday:
		return "Selasa"
	case time.Wednesday:
		return "Rabu"
	case time.Thursday:
		return "Kamis"
	case time.Friday:
		return "Jumat"
	case time.Saturday:
		return "Sabtu"
	case time.Sunday:
		return "Minggu"
	default:
		return ""
	}
}
