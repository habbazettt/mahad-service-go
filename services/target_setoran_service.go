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

type TargetSemesterService struct {
	DB *gorm.DB
}

// NewTargetSemesterService membuat instance service target semester
func NewTargetSemesterService(db *gorm.DB) *TargetSemesterService {
	return &TargetSemesterService{DB: db}
}

// CreateTargetSemester - Membuat target semester baru
// @Summary Membuat target semester baru
// @Description Endpoint ini digunakan untuk membuat target semester untuk mahasantri
// @Tags TargetSemester
// @Accept json
// @Produce json
// @Param request body dto.CreateTargetSemesterRequest true "Create Target Semester Request"
// @Success 201 {object} utils.Response "Target semester created successfully"
// @Failure 400 {object} utils.Response "Invalid request body"
// @Failure 404 {object} utils.Response "Mahasantri not found"
// @Failure 409 {object} utils.Response "Target semester already exists"
// @Failure 500 {object} utils.Response "Failed to create target semester"
// @Security BearerAuth
// @Router /api/v1/target_semester [post]
func (s *TargetSemesterService) CreateTargetSemester(c *fiber.Ctx) error {
	var req dto.CreateTargetSemesterRequest
	if err := c.BodyParser(&req); err != nil {
		logrus.WithError(err).Error("Invalid request body")
		return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	// Cek apakah Mahasantri ada
	var mahasantri models.Mahasantri
	if err := s.DB.First(&mahasantri, req.MahasantriID).Error; err != nil {
		logrus.WithFields(logrus.Fields{
			"mahasantri_id": req.MahasantriID,
		}).Warn("Mahasantri not found")
		return utils.ResponseError(c, fiber.StatusNotFound, "Mahasantri not found", nil)
	}

	// Cek apakah sudah ada target semester yang sama (MahasantriID + Semester + TahunAjaran)
	var existingTarget models.TargetSemester
	err := s.DB.Where("mahasantri_id = ? AND semester = ? AND tahun_ajaran = ?",
		req.MahasantriID, req.Semester, req.TahunAjaran).
		First(&existingTarget).Error

	if err == nil {
		logrus.WithFields(logrus.Fields{
			"mahasantri_id": req.MahasantriID,
			"semester":      req.Semester,
			"tahun_ajaran":  req.TahunAjaran,
		}).Warn("Target semester already exists")
		return utils.ResponseError(c, fiber.StatusConflict, "Target semester already exists for this Mahasantri, semester, and academic year", nil)
	} else if err != gorm.ErrRecordNotFound {
		logrus.WithError(err).Error("Database error when checking existing target semester")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to check existing target semester", err.Error())
	}

	// Buat target semester baru
	targetSemester := models.TargetSemester{
		MahasantriID: req.MahasantriID,
		Semester:     req.Semester,
		TahunAjaran:  req.TahunAjaran,
		Target:       req.Target,
		Keterangan:   req.Keterangan,
	}

	if err := s.DB.Create(&targetSemester).Error; err != nil {
		logrus.WithError(err).Error("Failed to create target semester")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to create target semester", err.Error())
	}

	// Build DTO Response
	response := dto.TargetSemesterResponse{
		ID:           targetSemester.ID,
		MahasantriID: targetSemester.MahasantriID,
		Target:       targetSemester.Target,
		Semester:     targetSemester.Semester,
		TahunAjaran:  targetSemester.TahunAjaran,
		Keterangan:   targetSemester.Keterangan,
	}

	logrus.WithFields(logrus.Fields{
		"target_semester_id": targetSemester.ID,
		"mahasantri_id":      targetSemester.MahasantriID,
		"semester":           targetSemester.Semester,
		"tahun_ajaran":       targetSemester.TahunAjaran,
		"target":             targetSemester.Target,
		"keterangan":         targetSemester.Keterangan,
	}).Info("Target semester created successfully")

	return utils.SuccessResponse(c, fiber.StatusCreated, "Target semester created successfully", response)
}

// CreateTargetSemester - Membuat target semester baru
// @Summary Membuat target semester baru
// @Description Endpoint ini digunakan untuk membuat target semester untuk mahasantri
// @Tags TargetSemester
// @Accept json
// @Produce json
// @Param request body dto.CreateTargetSemesterRequest true "Create Target Semester Request"
// @Success 201 {object} utils.Response "Target semester created successfully"
// @Failure 400 {object} utils.Response "Invalid request body"
// @Failure 404 {object} utils.Response "Mahasantri not found"
// @Failure 500 {object} utils.Response "Failed to create target semester"
// @Security BearerAuth
// @Router /api/v1/target_semester [post]
func (s *TargetSemesterService) GetAllTargetSemesters(c *fiber.Ctx) error {
	// Ambil query parameter untuk filtering
	semester := c.Query("semester")        // Optional filter by semester (ganjil/genap)
	tahunAjaran := c.Query("tahun_ajaran") // Optional filter by tahun ajaran

	// Ambil query parameter untuk pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	// Query TargetSemester
	query := s.DB.Model(&models.TargetSemester{})

	// Apply filter kalau ada
	if semester != "" {
		query = query.Where("semester = ?", semester)
	}
	if tahunAjaran != "" {
		query = query.Where("tahun_ajaran = ?", tahunAjaran)
	}

	// Hitung total TargetSemester untuk pagination
	var totalTargetSemester int64
	if err := query.Count(&totalTargetSemester).Error; err != nil {
		logrus.WithError(err).Error("Failed to count target semesters")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to count target semesters", err.Error())
	}

	// Ambil data TargetSemester
	var targetSemesters []models.TargetSemester
	if err := query.
		Preload("Mahasantri", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "nama", "nim", "jurusan", "gender", "mentor_id")
		}).
		Preload("Mahasantri.Mentor", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "nama", "email")
		}).
		Limit(limit).
		Offset(offset).
		Order("tahun_ajaran DESC, semester ASC").
		Find(&targetSemesters).Error; err != nil {
		logrus.WithError(err).Error("Failed to fetch target semesters")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to fetch target semesters", err.Error())
	}

	// Format response
	response := fiber.Map{
		"target_semesters": targetSemesters,
		"pagination": fiber.Map{
			"current_page": page,
			"total_data":   totalTargetSemester,
			"total_pages":  int(math.Ceil(float64(totalTargetSemester) / float64(limit))),
		},
	}

	logrus.Info("Fetched all target semesters successfully")
	return utils.SuccessResponse(c, fiber.StatusOK, "Fetched all target semesters successfully", response)
}

// GetTargetSemesterByMahasantriID - Mengambil target semester berdasarkan ID mahasantri
// @Summary Mengambil target semester berdasarkan ID mahasantri
// @Description Endpoint ini digunakan untuk mengambil target semester milik mahasantri tertentu, dengan optional filter dan pagination
// @Tags TargetSemester
// @Accept json
// @Produce json
// @Param mahasantri_id path string true "Mahasantri ID"
// @Param semester query string false "Filter by semester"
// @Param tahun_ajaran query string false "Filter by tahun ajaran"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} utils.Response "Fetched target semesters successfully"
// @Failure 404 {object} utils.Response "Mahasantri not found"
// @Failure 500 {object} utils.Response "Failed to fetch target semesters"
// @Security BearerAuth
// @Router /api/v1/target_semester/mahasantri/{mahasantri_id} [get]
func (s *TargetSemesterService) GetTargetSemesterByMahasantriID(c *fiber.Ctx) error {
	mahasantriID := c.Params("mahasantri_id")

	// Ambil query parameter untuk filtering
	semester := c.Query("semester")        // Optional filter by semester (ganjil/genap)
	tahunAjaran := c.Query("tahun_ajaran") // Optional filter by tahun ajaran

	// Ambil query parameter untuk pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	// Cek Mahasantri apakah ada
	var mahasantri models.Mahasantri
	if err := s.DB.Select("id", "nama", "nim", "jurusan", "gender", "mentor_id").
		Preload("Mentor", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "nama", "email")
		}).
		First(&mahasantri, mahasantriID).Error; err != nil {
		logrus.WithError(err).WithField("mahasantri_id", mahasantriID).Warn("Mahasantri not found")
		return utils.ResponseError(c, fiber.StatusNotFound, "Mahasantri not found", nil)
	}

	// Query TargetSemester dengan filter
	query := s.DB.Where("mahasantri_id = ?", mahasantriID)

	if semester != "" {
		query = query.Where("semester = ?", semester)
	}
	if tahunAjaran != "" {
		query = query.Where("tahun_ajaran = ?", tahunAjaran)
	}

	// Hitung total TargetSemester untuk pagination
	var totalTargetSemester int64
	query.Model(&models.TargetSemester{}).Count(&totalTargetSemester)

	// Ambil data TargetSemester
	var targetSemesters []models.TargetSemester
	if err := query.
		Preload("Mahasantri", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "nama", "nim", "mentor_id")
		}).
		Preload("Mahasantri.Mentor", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "nama", "email")
		}).
		Limit(limit).
		Offset(offset).
		Find(&targetSemesters).Error; err != nil {
		logrus.WithError(err).WithField("mahasantri_id", mahasantriID).Error("Failed to fetch target semesters")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to fetch target semesters", err.Error())
	}

	// Format response
	response := fiber.Map{
		"mahasantri": fiber.Map{
			"id":       mahasantri.ID,
			"mentorID": mahasantri.MentorID,
			"nama":     mahasantri.Nama,
			"nim":      mahasantri.NIM,
			"jurusan":  mahasantri.Jurusan,
			"gender":   mahasantri.Gender,
			"mentor": fiber.Map{
				"id":    mahasantri.Mentor.ID,
				"nama":  mahasantri.Mentor.Nama,
				"email": mahasantri.Mentor.Email,
			},
		},
		"target_semester": targetSemesters,
		"pagination": fiber.Map{
			"current_page": page,
			"total_data":   totalTargetSemester,
			"total_pages":  int(math.Ceil(float64(totalTargetSemester) / float64(limit))),
		},
	}

	logrus.WithField("mahasantri_id", mahasantriID).Info("Fetched target semesters with pagination successfully")
	return utils.SuccessResponse(c, fiber.StatusOK, "Fetched target semesters successfully", response)
}

// UpdateTargetSemester - Update data target semester
// @Summary Update data target semester
// @Description Endpoint ini digunakan untuk mengupdate data target semester berdasarkan ID
// @Tags TargetSemester
// @Accept json
// @Produce json
// @Param id path string true "Target Semester ID"
// @Param request body dto.UpdateTargetSemesterRequest true "Update Target Semester Request"
// @Success 200 {object} utils.Response "Target semester updated successfully"
// @Failure 400 {object} utils.Response "Invalid request body or no changes detected"
// @Failure 404 {object} utils.Response "Target semester not found"
// @Failure 500 {object} utils.Response "Failed to update target semester"
// @Security BearerAuth
// @Router /api/v1/target_semester/{id} [put]
func (s *TargetSemesterService) UpdateTargetSemester(c *fiber.Ctx) error {
	id := c.Params("id")
	var targetSemester models.TargetSemester

	// Cari dulu record-nya
	if err := s.DB.First(&targetSemester, id).Error; err != nil {
		logrus.WithField("target_semester_id", id).Warn("Target semester not found")
		return utils.ResponseError(c, fiber.StatusNotFound, "Target semester not found", nil)
	}

	// Bind request body
	var updateRequest dto.UpdateTargetSemesterRequest
	if err := c.BodyParser(&updateRequest); err != nil {
		logrus.WithError(err).Error("Failed to parse request body")
		return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	updated := false
	if updateRequest.Semester != nil && *updateRequest.Semester != targetSemester.Semester {
		targetSemester.Semester = *updateRequest.Semester
		updated = true
	}
	if updateRequest.TahunAjaran != nil && *updateRequest.TahunAjaran != targetSemester.TahunAjaran {
		targetSemester.TahunAjaran = *updateRequest.TahunAjaran
		updated = true
	}
	if updateRequest.Target != nil && *updateRequest.Target != targetSemester.Target {
		targetSemester.Target = *updateRequest.Target
		updated = true
	}
	if updateRequest.Keterangan != nil && *updateRequest.Keterangan != targetSemester.Keterangan {
		targetSemester.Keterangan = *updateRequest.Keterangan
		updated = true
	}

	// Kalau gak ada perubahan, return error
	if !updated {
		return utils.ResponseError(c, fiber.StatusBadRequest, "No changes detected", nil)
	}

	// Save perubahan
	if err := s.DB.Save(&targetSemester).Error; err != nil {
		logrus.WithError(err).Error("Failed to update target semester")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to update target semester", err.Error())
	}

	response := dto.TargetSemesterResponse{
		ID:           targetSemester.ID,
		MahasantriID: targetSemester.MahasantriID,
		Semester:     targetSemester.Semester,
		TahunAjaran:  targetSemester.TahunAjaran,
		Target:       targetSemester.Target,
		Keterangan:   targetSemester.Keterangan,
	}

	logrus.WithField("target_semester_id", id).Info("Target semester updated successfully")
	return utils.SuccessResponse(c, fiber.StatusOK, "Target semester updated successfully", response)
}

// DeleteTargetSemester - Menghapus target semester
// @Summary Menghapus target semester
// @Description Endpoint ini digunakan untuk menghapus target semester berdasarkan ID
// @Tags TargetSemester
// @Accept json
// @Produce json
// @Param id path string true "Target Semester ID"
// @Success 200 {object} utils.Response "Target semester deleted successfully"
// @Failure 404 {object} utils.Response "Target semester not found"
// @Failure 500 {object} utils.Response "Failed to delete target semester"
// @Security BearerAuth
// @Router /api/v1/target_semester/{id} [delete]
func (s *TargetSemesterService) DeleteTargetSemester(c *fiber.Ctx) error {
	id := c.Params("id")

	// Cari target semester
	var targetSemester models.TargetSemester
	if err := s.DB.First(&targetSemester, id).Error; err != nil {
		logrus.WithField("target_semester_id", id).Warn("Target semester not found")
		return utils.ResponseError(c, fiber.StatusNotFound, "Target semester not found", nil)
	}

	// Hapus target semester
	if err := s.DB.Delete(&targetSemester).Error; err != nil {
		logrus.WithError(err).Error("Failed to delete target semester")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to delete target semester", err.Error())
	}

	logrus.WithField("target_semester_id", id).Info("Target semester deleted successfully")
	return utils.SuccessResponse(c, fiber.StatusOK, "Target semester deleted successfully", nil)
}
