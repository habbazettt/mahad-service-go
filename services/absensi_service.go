package services

import (
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
func (s *AbsensiService) GetAbsensiByMahasantriID(c *fiber.Ctx) error {
	id := c.Params("mahasantri_id")

	if id == "" {
		return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid Mahasantri ID", "Mahasantri ID is required")
	}

	mahasantriID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid Mahasantri ID format", err.Error())
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	waktuFilter := c.Query("waktu")
	statusFilter := c.Query("status")

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

	var absensi []models.Absensi
	if err := query.Find(&absensi).Error; err != nil {
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

	// Calculate total absensi for "shubuh" and "isya"
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

	// Return response with calculated totals
	return utils.SuccessResponse(c, fiber.StatusOK, "Absensi retrieved successfully", fiber.Map{
		"absensi":                 response,
		"total_hadir":             totalHadir,
		"total_izin":              totalIzin,
		"total_alpa":              totalAlpa,
		"total_hadir_shubuh":      totalHadirShubuh,
		"total_hadir_isya":        totalHadirIsya,
		"total_izin_shubuh":       totalIzinShubuh,
		"total_izin_isya":         totalIzinIsya,
		"total_alpa_shubuh":       totalAlpaShubuh,
		"total_alpa_isya":         totalAlpaIsya,
		"total_absensi_per_week":  totalAbsensiPerWeek,
		"total_absensi_per_month": totalAbsensiPerMonth,
	})
}

// GetAttendancePerMonth - Mengambil total kehadiran per bulan berdasarkan Mahasantri ID
func (s *AbsensiService) GetAttendancePerMonth(c *fiber.Ctx) error {
	id := c.Params("mahasantri_id")

	if id == "" {
		return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid Mahasantri ID", "Mahasantri ID is required")
	}

	mahasantriID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid Mahasantri ID format", err.Error())
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

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

	// Query to get absensi per month
	var absensiPerMonth []struct {
		Month string `json:"month"`
		Year  int    `json:"year"`
	}

	// Main query for filtering by Mahasantri ID, status = "hadir"
	query := s.DB.Model(&models.Absensi{}).
		Where("mahasantri_id = ?", mahasantriID).
		Where("status = ?", "hadir").
		Select("EXTRACT(MONTH FROM tanggal) AS month, EXTRACT(YEAR FROM tanggal) AS year").
		Group("year, month").
		Order("year DESC, month DESC")

	// Apply date filtering if present
	if !start.IsZero() && !end.IsZero() {
		query = query.Where("tanggal BETWEEN ? AND ?", start, end)
	}

	if err := query.Scan(&absensiPerMonth).Error; err != nil {
		logrus.WithError(err).Error("Failed to fetch attendance per month")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to fetch attendance per month", err.Error())
	}

	// Mapping numeric month to month name
	monthNames := map[string]string{
		"1": "January", "2": "February", "3": "March", "4": "April",
		"5": "May", "6": "June", "7": "July", "8": "August",
		"9": "September", "10": "October", "11": "November", "12": "December",
	}

	var attendanceData []fiber.Map
	var totalAttendance int64

	// Loop through the months and calculate total counts for each status (Hadir, Izin, Alpa)
	for _, item := range absensiPerMonth {
		var countHadirShubuh, countHadirIsya int64
		var countIzinShubuh, countIzinIsya int64
		var countAlpaShubuh, countAlpaIsya int64

		// Calculate total "hadir" for Shubuh
		if err := s.DB.Model(&models.Absensi{}).
			Where("mahasantri_id = ?", mahasantriID).
			Where("EXTRACT(MONTH FROM tanggal) = ?", item.Month).
			Where("EXTRACT(YEAR FROM tanggal) = ?", item.Year).
			Where("waktu = ?", "shubuh").
			Where("status = ?", "hadir").
			Count(&countHadirShubuh).Error; err != nil {
			logrus.WithError(err).Error("Failed to count hadir shubuh attendance for month")
			return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to count hadir shubuh attendance for month", err.Error())
		}

		// Calculate total "hadir" for Isya
		if err := s.DB.Model(&models.Absensi{}).
			Where("mahasantri_id = ?", mahasantriID).
			Where("EXTRACT(MONTH FROM tanggal) = ?", item.Month).
			Where("EXTRACT(YEAR FROM tanggal) = ?", item.Year).
			Where("waktu = ?", "isya").
			Where("status = ?", "hadir").
			Count(&countHadirIsya).Error; err != nil {
			logrus.WithError(err).Error("Failed to count hadir isya attendance for month")
			return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to count hadir isya attendance for month", err.Error())
		}

		// Calculate total "izin" for Shubuh
		if err := s.DB.Model(&models.Absensi{}).
			Where("mahasantri_id = ?", mahasantriID).
			Where("EXTRACT(MONTH FROM tanggal) = ?", item.Month).
			Where("EXTRACT(YEAR FROM tanggal) = ?", item.Year).
			Where("waktu = ?", "shubuh").
			Where("status = ?", "izin").
			Count(&countIzinShubuh).Error; err != nil {
			logrus.WithError(err).Error("Failed to count izin shubuh attendance for month")
			return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to count izin shubuh attendance for month", err.Error())
		}

		// Calculate total "izin" for Isya
		if err := s.DB.Model(&models.Absensi{}).
			Where("mahasantri_id = ?", mahasantriID).
			Where("EXTRACT(MONTH FROM tanggal) = ?", item.Month).
			Where("EXTRACT(YEAR FROM tanggal) = ?", item.Year).
			Where("waktu = ?", "isya").
			Where("status = ?", "izin").
			Count(&countIzinIsya).Error; err != nil {
			logrus.WithError(err).Error("Failed to count izin isya attendance for month")
			return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to count izin isya attendance for month", err.Error())
		}

		// Calculate total "alpa" for Shubuh
		if err := s.DB.Model(&models.Absensi{}).
			Where("mahasantri_id = ?", mahasantriID).
			Where("EXTRACT(MONTH FROM tanggal) = ?", item.Month).
			Where("EXTRACT(YEAR FROM tanggal) = ?", item.Year).
			Where("waktu = ?", "shubuh").
			Where("status = ?", "alpa").
			Count(&countAlpaShubuh).Error; err != nil {
			logrus.WithError(err).Error("Failed to count alpa shubuh attendance for month")
			return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to count alpa shubuh attendance for month", err.Error())
		}

		// Calculate total "alpa" for Isya
		if err := s.DB.Model(&models.Absensi{}).
			Where("mahasantri_id = ?", mahasantriID).
			Where("EXTRACT(MONTH FROM tanggal) = ?", item.Month).
			Where("EXTRACT(YEAR FROM tanggal) = ?", item.Year).
			Where("waktu = ?", "isya").
			Where("status = ?", "alpa").
			Count(&countAlpaIsya).Error; err != nil {
			logrus.WithError(err).Error("Failed to count alpa isya attendance for month")
			return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to count alpa isya attendance for month", err.Error())
		}

		attendanceData = append(attendanceData, fiber.Map{
			"month":        monthNames[item.Month],
			"year":         item.Year,
			"hadir_shubuh": countHadirShubuh,
			"hadir_isya":   countHadirIsya,
			"izin_shubuh":  countIzinShubuh,
			"izin_isya":    countIzinIsya,
			"alpa_shubuh":  countAlpaShubuh,
			"alpa_isya":    countAlpaIsya,
		})

		totalAttendance += countHadirShubuh + countHadirIsya + countIzinShubuh + countIzinIsya + countAlpaShubuh + countAlpaIsya
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Attendance per month retrieved successfully", fiber.Map{
		"mahasantri_id":        mahasantriID,
		"attendance_per_month": attendanceData,
		"total_attendance":     totalAttendance,
	})
}
