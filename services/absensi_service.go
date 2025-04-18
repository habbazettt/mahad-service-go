package services

import (
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
// @Failure 400 {object} utils.Response "Invalid request body or Absensi already recorded for this time"
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

	token = token[len("Bearer "):]

	claims, err := utils.VerifyToken(token)
	if err != nil {
		logrus.WithError(err).Error("Failed to verify JWT")
		return utils.ResponseError(c, fiber.StatusUnauthorized, "Unauthorized", err.Error())
	}

	mentorID := claims.ID

	var mahasantri models.Mahasantri
	if err := s.DB.First(&mahasantri, req.MahasantriID).Error; err != nil {
		logrus.WithError(err).Error("Mahasantri not found")
		return utils.ResponseError(c, fiber.StatusNotFound, "Mahasantri not found", err.Error())
	}

	var existingAbsensi models.Absensi
	if err := s.DB.Where("mahasantri_id = ? AND tanggal = ? AND waktu = ?", req.MahasantriID, time.Now().Format("2006-01-02"), req.Waktu).First(&existingAbsensi).Error; err == nil {
		logrus.Warn("Absensi sudah tercatat")
		return utils.ResponseError(c, fiber.StatusBadRequest, "Absensi already recorded for this time", "Absensi sudah tercatat untuk waktu tersebut")
	}

	absensi := models.Absensi{
		MahasantriID: req.MahasantriID,
		MentorID:     mentorID,
		Waktu:        req.Waktu,
		Status:       req.Status,
		Tanggal:      time.Now(),
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

// GetAbsensiByMahasantriID - Mengambil absensi berdasarkan Mahasantri ID dengan rentang tanggal dan filter waktu/status
// @Summary Mendapatkan daftar absensi berdasarkan Mahasantri ID dengan filter tanggal, waktu, dan status
// @Description Endpoint ini digunakan untuk mendapatkan daftar absensi dari Mahasantri tertentu berdasarkan filter tanggal, waktu, status, serta mendukung paginasi.
// @Tags Absensi
// @Accept json
// @Produce json
// @Param mahasantri_id path int true "Mahasantri ID"
// @Param start_date query string false "Tanggal awal filter (format: dd-mm-yyyy)"
// @Param end_date query string false "Tanggal akhir filter (format: dd-mm-yyyy)"
// @Param waktu query string false "Waktu for filtering absensi" (shubuh, isya)
// @Param status query string false "Status for filtering absensi" (hadir, izin, alpa)
// @Param page query int false "Page number for pagination" default(1)
// @Param limit query int false "Limit number of results per page" default(10)
// @Success 200 {object} utils.Response "Absensi retrieved successfully"
// @Failure 400 {object} utils.Response "Invalid input or query parameters"
// @Failure 404 {object} utils.Response "Mahasantri not found"
// @Failure 500 {object} utils.Response "Failed to retrieve absensi"
// @Security BearerAuth
// @Router /api/v1/absensi/mahasantri/{mahasantri_id} [get]
func (s *AbsensiService) GetAbsensiByMahasantriID(c *fiber.Ctx) error {
	id := c.Params("mahasantri_id")

	if id == "" {
		return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid Mahasantri ID", "Mahasantri ID is required")
	}

	mahasantriID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid Mahasantri ID format", err.Error())
	}

	// Get query parameters for start_date, end_date, waktu, status
	startDate := c.Query("start_date") // ex: 04-01-2023
	endDate := c.Query("end_date")     // ex: 04-01-2023
	waktuFilter := c.Query("waktu")
	statusFilter := c.Query("status")

	// Get pagination parameters
	page, err := strconv.Atoi(c.Query("page", "1")) // Default to page 1 if not provided
	if err != nil || page <= 0 {
		page = 1
	}
	limit, err := strconv.Atoi(c.Query("limit", "10")) // Default to 10 results per page if not provided
	if err != nil || limit <= 0 {
		limit = 10
	}
	offset := (page - 1) * limit // Calculate offset for pagination

	// Set date layout
	layout := "02-01-2006"
	var start, end time.Time

	if startDate != "" {
		start, err = time.Parse(layout, startDate)
		if err != nil {
			return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid start date format", err.Error())
		}
	}

	if endDate != "" {
		end, err = time.Parse(layout, endDate)
		if err != nil {
			return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid end date format", err.Error())
		}
		end = end.Add(24*time.Hour - time.Second) // Ensure the end date includes the whole day
	}

	// Query untuk mengambil absensi berdasarkan Mahasantri ID
	query := s.DB.Preload("Mentor").Preload("Mahasantri").Where("mahasantri_id = ?", mahasantriID)

	if !start.IsZero() && !end.IsZero() {
		query = query.Where("tanggal BETWEEN ? AND ?", start, end)
	}

	if waktuFilter != "" {
		query = query.Where("waktu = ?", waktuFilter)
	}

	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}

	// Count total absensi untuk pagination
	var totalAbsensi int64
	if err := query.Model(&models.Absensi{}).Count(&totalAbsensi).Error; err != nil {
		logrus.WithError(err).Error("Failed to count total absensi")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to count total absensi", err.Error())
	}

	// Apply pagination
	var absensi []models.Absensi
	if err := query.Limit(limit).Offset(offset).Find(&absensi).Error; err != nil {
		logrus.WithError(err).Error("Failed to fetch absensi")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to fetch absensi", err.Error())
	}

	// Calculate total absensi for different statuses and times (shubuh, isya, hadir, izin, alpa)
	var totalHadir, totalIzin, totalAlpa int64
	var totalHadirShubuh, totalHadirIsya int64
	var totalIzinShubuh, totalIzinIsya int64
	var totalAlpaShubuh, totalAlpaIsya int64

	// Calculate total absensi for "hadir"
	if err := s.DB.Model(&models.Absensi{}).Where("mahasantri_id = ?", mahasantriID).
		Where("status = ?", "hadir").
		Count(&totalHadir).Error; err != nil {
		logrus.WithError(err).Error("Failed to fetch total hadir")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to fetch total hadir", err.Error())
	}

	// Calculate total absensi for "izin"
	if err := s.DB.Model(&models.Absensi{}).Where("mahasantri_id = ?", mahasantriID).
		Where("status = ?", "izin").
		Count(&totalIzin).Error; err != nil {
		logrus.WithError(err).Error("Failed to fetch total izin")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to fetch total izin", err.Error())
	}

	// Calculate total absensi for "alpa"
	if err := s.DB.Model(&models.Absensi{}).Where("mahasantri_id = ?", mahasantriID).
		Where("status = ?", "alpa").
		Count(&totalAlpa).Error; err != nil {
		logrus.WithError(err).Error("Failed to fetch total alpa")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to fetch total alpa", err.Error())
	}

	//! Calculate total absensi for "shubuh" and "isya"
	// Calculate total hadir shubuh
	if err := s.DB.Model(&models.Absensi{}).Where("mahasantri_id = ?", mahasantriID).
		Where("waktu = ?", "shubuh").
		Where("status = ?", "hadir").
		Count(&totalHadirShubuh).Error; err != nil {
		logrus.WithError(err).Error("Failed to fetch total hadir shubuh")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to fetch total hadir shubuh", err.Error())
	}

	// Calculate total hadir isya
	if err := s.DB.Model(&models.Absensi{}).Where("mahasantri_id = ?", mahasantriID).
		Where("waktu = ?", "isya").
		Where("status = ?", "hadir").
		Count(&totalHadirIsya).Error; err != nil {
		logrus.WithError(err).Error("Failed to fetch total hadir isya")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to fetch total hadir isya", err.Error())
	}

	// Calculate total izin shubuh
	if err := s.DB.Model(&models.Absensi{}).Where("mahasantri_id = ?", mahasantriID).
		Where("waktu = ?", "shubuh").
		Where("status = ?", "izin").
		Count(&totalIzinShubuh).Error; err != nil {
		logrus.WithError(err).Error("Failed to fetch total izin shubuh")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to fetch total izin shubuh", err.Error())
	}

	// Calculate total izin isya
	if err := s.DB.Model(&models.Absensi{}).Where("mahasantri_id = ?", mahasantriID).
		Where("waktu = ?", "isya").
		Where("status = ?", "izin").
		Count(&totalIzinIsya).Error; err != nil {
		logrus.WithError(err).Error("Failed to fetch total izin isya")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to fetch total izin isya", err.Error())
	}

	// Calculate total alpa shubuh
	if err := s.DB.Model(&models.Absensi{}).Where("mahasantri_id = ?", mahasantriID).
		Where("waktu = ?", "shubuh").
		Where("status = ?", "alpa").
		Count(&totalAlpaShubuh).Error; err != nil {
		logrus.WithError(err).Error("Failed to fetch total alpa shubuh")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to fetch total alpa shubuh", err.Error())
	}

	// Calculate total alpa isya
	if err := s.DB.Model(&models.Absensi{}).Where("mahasantri_id = ?", mahasantriID).
		Where("waktu = ?", "isya").
		Where("status = ?", "alpa").
		Count(&totalAlpaIsya).Error; err != nil {
		logrus.WithError(err).Error("Failed to fetch total alpa isya")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to fetch total alpa isya", err.Error())
	}

	// Calculate total absensi per week
	var totalAbsensiPerWeek int64
	if err := s.DB.Model(&models.Absensi{}).Where("mahasantri_id = ?", mahasantriID).
		Where("tanggal BETWEEN ? AND ?", start, end).
		Select("count(*)").
		Scan(&totalAbsensiPerWeek).Error; err != nil {
		logrus.WithError(err).Error("Failed to fetch total absensi per week")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to fetch total absensi per week", err.Error())
	}

	// Calculate total absensi per month
	var totalAbsensiPerMonth int64
	if err := s.DB.Model(&models.Absensi{}).Where("mahasantri_id = ?", mahasantriID).
		Where("tanggal BETWEEN ? AND ?", start, end).
		Select("count(*)").
		Scan(&totalAbsensiPerMonth).Error; err != nil {
		logrus.WithError(err).Error("Failed to fetch total absensi per month")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to fetch total absensi per month", err.Error())
	}

	// Response for the absensi
	response := make([]dto.AbsensiResponseDTO, len(absensi))
	for i, a := range absensi {
		response[i] = dto.AbsensiResponseDTO{
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

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalAbsensi) / float64(limit)))

	// Return response with pagination info
	return utils.SuccessResponse(c, fiber.StatusOK, "Absensi retrieved successfully", fiber.Map{
		"absensi":     response,
		"total_hadir": totalHadir,
		"total_izin":  totalIzin,
		"total_alpa":  totalAlpa,
		"pagination": fiber.Map{
			"current_page":  page,
			"total_absensi": totalAbsensi,
			"total_pages":   totalPages,
			"limit":         limit,
		},
	})
}

// GetAttendancePerMonth godoc
// @Summary Mendapatkan ringkasan absensi bulanan Mahasantri
// @Description Mengambil total absensi bulanan berdasarkan waktu (shubuh & isya) dan status (hadir, izin, alpa) dalam satu bulan tertentu.
// @Tags Absensi
// @Security BearerAuth
// @Param mahasantri_id path int true "ID Mahasantri"
// @Param month query string true "Bulan (format: MM, contoh: 04 untuk April)"
// @Param year query string true "Tahun (format: YYYY, contoh: 2025)"
// @Success 200 {object} utils.SuccessResponseSwagger{data=dto.AbsensiMonthlySummaryDTO} "Berhasil mengambil ringkasan absensi bulanan"
// @Failure 400 {object} utils.ErrorResponseSwagger "Bad Request - Format salah atau parameter tidak lengkap"
// @Failure 500 {object} utils.ErrorResponseSwagger "Internal Server Error - Gagal mengambil data absensi"
// @Router /api/v1/absensi/mahasantri/{mahasantri_id}/per-month [get]
func (s *AbsensiService) GetAttendancePerMonth(c *fiber.Ctx) error {
	id := c.Params("mahasantri_id")

	if id == "" {
		return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid Mahasantri ID", "Mahasantri ID is required")
	}

	mahasantriID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid Mahasantri ID format", err.Error())
	}

	monthStr := c.Query("month") // format: 04
	yearStr := c.Query("year")   // format: 2025

	if monthStr == "" || yearStr == "" {
		return utils.ResponseError(c, fiber.StatusBadRequest, "Missing query parameters", "month and year are required")
	}

	layout := "02-01-2006"
	location := time.Now().Location()
	startDate, err := time.ParseInLocation(layout, fmt.Sprintf("01-%s-%s", monthStr, yearStr), location)
	if err != nil {
		return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid date format", err.Error())
	}
	endDate := startDate.AddDate(0, 1, -1)

	// Get numeric month & year
	monthInt, _ := strconv.Atoi(monthStr)
	yearInt, _ := strconv.Atoi(yearStr)

	// Query attendance grouped by waktu and status
	type Result struct {
		Waktu  string
		Status string
		Count  int64
	}

	var results []Result
	if err := s.DB.Model(&models.Absensi{}).
		Select("waktu, status, COUNT(*) as count").
		Where("mahasantri_id = ?", mahasantriID).
		Where("tanggal BETWEEN ? AND ?", startDate, endDate).
		Group("waktu, status").
		Scan(&results).Error; err != nil {
		logrus.WithError(err).Error("Failed to fetch attendance data")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to fetch attendance data", err.Error())
	}

	// Initialize counts
	response := fiber.Map{
		"month":       time.Month(monthInt).String(),
		"year":        yearInt,
		"total_hadir": 0,
		"total_izin":  0,
		"total_alpa":  0,
		"shubuh": fiber.Map{
			"hadir": 0, "izin": 0, "alpa": 0,
		},
		"isya": fiber.Map{
			"hadir": 0, "izin": 0, "alpa": 0,
		},
	}

	for _, r := range results {
		// Validate waktu & status
		if r.Waktu != "shubuh" && r.Waktu != "isya" {
			continue
		}
		if r.Status != "hadir" && r.Status != "izin" && r.Status != "alpa" {
			continue
		}

		// Set ke map
		if _, ok := response[r.Waktu].(fiber.Map); ok {
			response[r.Waktu].(fiber.Map)[r.Status] = r.Count
		}

		// Tambah total per status
		switch r.Status {
		case "hadir":
			response["total_hadir"] = response["total_hadir"].(int) + int(r.Count)
		case "izin":
			response["total_izin"] = response["total_izin"].(int) + int(r.Count)
		case "alpa":
			response["total_alpa"] = response["total_alpa"].(int) + int(r.Count)
		}
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Attendance per month retrieved successfully", response)
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

		if data, ok := absensiMap[tanggal]; ok {
			if val, exists := data["shubuh"]; exists {
				shubuh = val
			}
			if val, exists := data["isya"]; exists {
				isya = val
			}
		}

		summary = append(summary, dto.AbsensiDailySummaryDTO{
			Tanggal: tanggal,
			Shubuh:  shubuh,
			Isya:    isya,
		})
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Daily summary retrieved successfully", summary)
}
