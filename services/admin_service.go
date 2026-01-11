package services

import (
	"github.com/gofiber/fiber/v2"
	"github.com/habbazettt/mahad-service-go/dto"
	"github.com/habbazettt/mahad-service-go/models"
	"github.com/habbazettt/mahad-service-go/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// AdminService menangani logika admin
type AdminService struct {
	DB *gorm.DB
}

// LoginAdmin godoc
// @Summary Login Admin
// @Description Melakukan login untuk admin dengan email dan password
// @Tags Admin Auth
// @Accept json
// @Produce json
// @Param body body dto.LoginAdminRequest true "Data login Admin"
// @Success 200 {object} utils.SuccessResponseSwagger
// @Failure 400 {object} utils.ErrorResponseSwagger
// @Failure 401 {object} utils.ErrorResponseSwagger
// @Router /api/v1/auth/login/admin [post]
func (s *AdminService) LoginAdmin(c *fiber.Ctx) error {
	var req dto.LoginAdminRequest

	if err := c.BodyParser(&req); err != nil {
		logrus.WithError(err).Error("Failed to parse request body")
		return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	var admin models.Admin
	if err := s.DB.Where("email = ?", req.Email).First(&admin).Error; err != nil {
		logrus.Warn("Invalid email or password: ", req.Email)
		return utils.ResponseError(c, fiber.StatusUnauthorized, "Invalid email or password", nil)
	}

	if !utils.ComparePassword(admin.Password, req.Password) {
		logrus.Warn("Invalid password for email: ", req.Email)
		return utils.ResponseError(c, fiber.StatusUnauthorized, "Invalid email or password", nil)
	}

	token, err := utils.GenerateToken(admin.ID, RoleAdmin)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate token")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to generate token", err.Error())
	}

	logrus.WithFields(logrus.Fields{
		"admin_id": admin.ID,
		"email":    admin.Email,
	}).Info("Admin logged in successfully")

	return utils.SuccessResponse(c, fiber.StatusOK, "Login successful", dto.AuthResponse{
		Token: token,
		User: dto.UserAdminResponse{
			ID:       admin.ID,
			Nama:     admin.Nama,
			Email:    admin.Email,
			UserType: RoleAdmin,
		},
	})
}

// RegisterAdmin godoc
// @Summary Register Admin (Admin Only)
// @Description Mendaftarkan akun Admin baru (hanya admin yang bisa mengakses)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body dto.RegisterAdminRequest true "Data pendaftaran Admin"
// @Success 201 {object} utils.SuccessResponseSwagger
// @Failure 400 {object} utils.ErrorResponseSwagger
// @Failure 409 {object} utils.ErrorResponseSwagger
// @Router /api/v1/admin/register [post]
func (s *AdminService) RegisterAdmin(c *fiber.Ctx) error {
	var req dto.RegisterAdminRequest

	if err := c.BodyParser(&req); err != nil {
		logrus.WithError(err).Error("Failed to parse request body")
		return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	// Cek apakah email sudah terdaftar
	var existingAdmin models.Admin
	if err := s.DB.Where("email = ?", req.Email).First(&existingAdmin).Error; err == nil {
		logrus.Warn("Email already registered: ", req.Email)
		return utils.ResponseError(c, fiber.StatusConflict, "Email already registered", nil)
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		logrus.WithError(err).Error("Failed to hash password")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to hash password", err.Error())
	}

	// Simpan data Admin
	admin := models.Admin{
		Nama:     req.Nama,
		Email:    req.Email,
		Password: hashedPassword,
	}

	if err := s.DB.Create(&admin).Error; err != nil {
		logrus.WithError(err).Error("Failed to register admin")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to register admin", err.Error())
	}

	logrus.WithFields(logrus.Fields{
		"admin_id": admin.ID,
		"nama":     admin.Nama,
		"email":    admin.Email,
	}).Info("Admin registered successfully")

	return utils.SuccessResponse(c, fiber.StatusCreated, "Admin registered successfully", dto.UserAdminResponse{
		ID:       admin.ID,
		Nama:     admin.Nama,
		Email:    admin.Email,
		UserType: RoleAdmin,
	})
}

// GetDashboardStats godoc
// @Summary Get Dashboard Statistics
// @Description Mengambil statistik untuk dashboard admin
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponseSwagger
// @Failure 401 {object} utils.ErrorResponseSwagger
// @Router /api/v1/admin/dashboard [get]
func (s *AdminService) GetDashboardStats(c *fiber.Ctx) error {
	var totalMentor, totalMahasantri, totalHafalan, totalAbsensi int64

	s.DB.Model(&models.Mentor{}).Count(&totalMentor)
	s.DB.Model(&models.Mahasantri{}).Count(&totalMahasantri)
	s.DB.Model(&models.Hafalan{}).Count(&totalHafalan)
	s.DB.Model(&models.Absensi{}).Count(&totalAbsensi)

	return utils.SuccessResponse(c, fiber.StatusOK, "Dashboard statistics retrieved", dto.AdminDashboardResponse{
		TotalMentor:     totalMentor,
		TotalMahasantri: totalMahasantri,
		TotalHafalan:    totalHafalan,
		TotalAbsensi:    totalAbsensi,
	})
}

// GetAllMentors godoc
// @Summary Get All Mentors (Admin)
// @Description Mengambil semua data mentor dengan statistik
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponseSwagger
// @Failure 401 {object} utils.ErrorResponseSwagger
// @Router /api/v1/admin/mentors [get]
func (s *AdminService) GetAllMentors(c *fiber.Ctx) error {
	var mentors []models.Mentor
	if err := s.DB.Find(&mentors).Error; err != nil {
		logrus.WithError(err).Error("Failed to fetch mentors")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to fetch mentors", err.Error())
	}

	var response []dto.AdminMentorResponse
	for _, mentor := range mentors {
		var mahasantriCount int64
		s.DB.Model(&models.Mahasantri{}).Where("mentor_id = ?", mentor.ID).Count(&mahasantriCount)

		response = append(response, dto.AdminMentorResponse{
			ID:              mentor.ID,
			Nama:            mentor.Nama,
			Email:           mentor.Email,
			Gender:          mentor.Gender,
			MahasantriCount: mahasantriCount,
		})
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Mentors retrieved successfully", response)
}

// CreateMentor godoc
// @Summary Create Mentor (Admin)
// @Description Membuat mentor baru oleh admin
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body dto.AdminCreateMentorRequest true "Data mentor baru"
// @Success 201 {object} utils.SuccessResponseSwagger
// @Failure 400 {object} utils.ErrorResponseSwagger
// @Failure 409 {object} utils.ErrorResponseSwagger
// @Router /api/v1/admin/mentors [post]
func (s *AdminService) CreateMentor(c *fiber.Ctx) error {
	var req dto.AdminCreateMentorRequest

	if err := c.BodyParser(&req); err != nil {
		logrus.WithError(err).Error("Failed to parse request body")
		return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	// Cek apakah email sudah terdaftar
	var existingMentor models.Mentor
	if err := s.DB.Where("email = ?", req.Email).First(&existingMentor).Error; err == nil {
		logrus.Warn("Email already registered: ", req.Email)
		return utils.ResponseError(c, fiber.StatusConflict, "Email already registered", nil)
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		logrus.WithError(err).Error("Failed to hash password")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to hash password", err.Error())
	}

	mentor := models.Mentor{
		Nama:     req.Nama,
		Email:    req.Email,
		Gender:   req.Gender,
		Password: hashedPassword,
	}

	if err := s.DB.Create(&mentor).Error; err != nil {
		logrus.WithError(err).Error("Failed to create mentor")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to create mentor", err.Error())
	}

	logrus.WithFields(logrus.Fields{
		"mentor_id": mentor.ID,
		"nama":      mentor.Nama,
		"email":     mentor.Email,
	}).Info("Mentor created by admin")

	return utils.SuccessResponse(c, fiber.StatusCreated, "Mentor created successfully", dto.AdminMentorResponse{
		ID:              mentor.ID,
		Nama:            mentor.Nama,
		Email:           mentor.Email,
		Gender:          mentor.Gender,
		MahasantriCount: 0,
	})
}

// UpdateMentor godoc
// @Summary Update Mentor (Admin)
// @Description Memperbarui data mentor oleh admin
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Mentor ID"
// @Param body body dto.AdminUpdateMentorRequest true "Data mentor yang diupdate"
// @Success 200 {object} utils.SuccessResponseSwagger
// @Failure 400 {object} utils.ErrorResponseSwagger
// @Failure 404 {object} utils.ErrorResponseSwagger
// @Router /api/v1/admin/mentors/{id} [put]
func (s *AdminService) UpdateMentor(c *fiber.Ctx) error {
	id := c.Params("id")

	var mentor models.Mentor
	if err := s.DB.First(&mentor, id).Error; err != nil {
		logrus.Warn("Mentor not found: ", id)
		return utils.ResponseError(c, fiber.StatusNotFound, "Mentor not found", nil)
	}

	var req dto.AdminUpdateMentorRequest
	if err := c.BodyParser(&req); err != nil {
		logrus.WithError(err).Error("Failed to parse request body")
		return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	// Update fields jika ada
	if req.Nama != "" {
		mentor.Nama = req.Nama
	}
	if req.Email != "" {
		// Cek apakah email baru sudah digunakan
		var existingMentor models.Mentor
		if err := s.DB.Where("email = ? AND id != ?", req.Email, mentor.ID).First(&existingMentor).Error; err == nil {
			return utils.ResponseError(c, fiber.StatusConflict, "Email already used by another mentor", nil)
		}
		mentor.Email = req.Email
	}
	if req.Gender != "" {
		mentor.Gender = req.Gender
	}

	if err := s.DB.Save(&mentor).Error; err != nil {
		logrus.WithError(err).Error("Failed to update mentor")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to update mentor", err.Error())
	}

	var mahasantriCount int64
	s.DB.Model(&models.Mahasantri{}).Where("mentor_id = ?", mentor.ID).Count(&mahasantriCount)

	logrus.WithFields(logrus.Fields{
		"mentor_id": mentor.ID,
	}).Info("Mentor updated by admin")

	return utils.SuccessResponse(c, fiber.StatusOK, "Mentor updated successfully", dto.AdminMentorResponse{
		ID:              mentor.ID,
		Nama:            mentor.Nama,
		Email:           mentor.Email,
		Gender:          mentor.Gender,
		MahasantriCount: mahasantriCount,
	})
}

// DeleteMentor godoc
// @Summary Delete Mentor (Admin)
// @Description Menghapus mentor oleh admin
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Mentor ID"
// @Success 200 {object} utils.SuccessResponseSwagger
// @Failure 404 {object} utils.ErrorResponseSwagger
// @Router /api/v1/admin/mentors/{id} [delete]
func (s *AdminService) DeleteMentor(c *fiber.Ctx) error {
	id := c.Params("id")

	var mentor models.Mentor
	if err := s.DB.First(&mentor, id).Error; err != nil {
		logrus.Warn("Mentor not found: ", id)
		return utils.ResponseError(c, fiber.StatusNotFound, "Mentor not found", nil)
	}

	if err := s.DB.Delete(&mentor).Error; err != nil {
		logrus.WithError(err).Error("Failed to delete mentor")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to delete mentor", err.Error())
	}

	logrus.WithFields(logrus.Fields{
		"mentor_id": mentor.ID,
		"nama":      mentor.Nama,
	}).Info("Mentor deleted by admin")

	return utils.SuccessResponse(c, fiber.StatusOK, "Mentor deleted successfully", nil)
}

// GetAllMahasantri godoc
// @Summary Get All Mahasantri (Admin)
// @Description Mengambil semua data mahasantri
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponseSwagger
// @Failure 401 {object} utils.ErrorResponseSwagger
// @Router /api/v1/admin/mahasantri [get]
func (s *AdminService) GetAllMahasantri(c *fiber.Ctx) error {
	var mahasantriList []models.Mahasantri
	if err := s.DB.Preload("Mentor").Find(&mahasantriList).Error; err != nil {
		logrus.WithError(err).Error("Failed to fetch mahasantri")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to fetch mahasantri", err.Error())
	}

	var response []dto.AdminMahasantriResponse
	for _, m := range mahasantriList {
		mentorNama := ""
		if m.Mentor != nil {
			mentorNama = m.Mentor.Nama
		}

		response = append(response, dto.AdminMahasantriResponse{
			ID:         m.ID,
			Nama:       m.Nama,
			NIM:        m.NIM,
			Jurusan:    m.Jurusan,
			Gender:     m.Gender,
			MentorID:   m.MentorID,
			MentorNama: mentorNama,
		})
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Mahasantri retrieved successfully", response)
}

// CreateMahasantri godoc
// @Summary Create Mahasantri (Admin)
// @Description Membuat mahasantri baru oleh admin
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body dto.AdminCreateMahasantriRequest true "Data mahasantri baru"
// @Success 201 {object} utils.SuccessResponseSwagger
// @Failure 400 {object} utils.ErrorResponseSwagger
// @Failure 409 {object} utils.ErrorResponseSwagger
// @Router /api/v1/admin/mahasantri [post]
func (s *AdminService) CreateMahasantri(c *fiber.Ctx) error {
	var req dto.AdminCreateMahasantriRequest

	if err := c.BodyParser(&req); err != nil {
		logrus.WithError(err).Error("Failed to parse request body")
		return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	// Cek apakah mentor ID valid
	var mentor models.Mentor
	if err := s.DB.First(&mentor, req.MentorID).Error; err != nil {
		logrus.Warn("Invalid mentor ID: ", req.MentorID)
		return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid mentor ID", nil)
	}

	// Cek apakah NIM sudah terdaftar
	var existingMahasantri models.Mahasantri
	if err := s.DB.Where("nim = ?", req.NIM).First(&existingMahasantri).Error; err == nil {
		logrus.Warn("NIM already registered: ", req.NIM)
		return utils.ResponseError(c, fiber.StatusConflict, "NIM already registered", nil)
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		logrus.WithError(err).Error("Failed to hash password")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to hash password", err.Error())
	}

	mahasantri := models.Mahasantri{
		Nama:     req.Nama,
		NIM:      req.NIM,
		Jurusan:  req.Jurusan,
		Gender:   req.Gender,
		Password: hashedPassword,
		MentorID: req.MentorID,
	}

	if err := s.DB.Create(&mahasantri).Error; err != nil {
		logrus.WithError(err).Error("Failed to create mahasantri")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to create mahasantri", err.Error())
	}

	logrus.WithFields(logrus.Fields{
		"mahasantri_id": mahasantri.ID,
		"nama":          mahasantri.Nama,
		"nim":           mahasantri.NIM,
	}).Info("Mahasantri created by admin")

	return utils.SuccessResponse(c, fiber.StatusCreated, "Mahasantri created successfully", dto.AdminMahasantriResponse{
		ID:         mahasantri.ID,
		Nama:       mahasantri.Nama,
		NIM:        mahasantri.NIM,
		Jurusan:    mahasantri.Jurusan,
		Gender:     mahasantri.Gender,
		MentorID:   mahasantri.MentorID,
		MentorNama: mentor.Nama,
	})
}

// UpdateMahasantri godoc
// @Summary Update Mahasantri (Admin)
// @Description Memperbarui data mahasantri oleh admin
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Mahasantri ID"
// @Param body body dto.AdminUpdateMahasantriRequest true "Data mahasantri yang diupdate"
// @Success 200 {object} utils.SuccessResponseSwagger
// @Failure 400 {object} utils.ErrorResponseSwagger
// @Failure 404 {object} utils.ErrorResponseSwagger
// @Router /api/v1/admin/mahasantri/{id} [put]
func (s *AdminService) UpdateMahasantri(c *fiber.Ctx) error {
	id := c.Params("id")

	var mahasantri models.Mahasantri
	if err := s.DB.First(&mahasantri, id).Error; err != nil {
		logrus.Warn("Mahasantri not found: ", id)
		return utils.ResponseError(c, fiber.StatusNotFound, "Mahasantri not found", nil)
	}

	var req dto.AdminUpdateMahasantriRequest
	if err := c.BodyParser(&req); err != nil {
		logrus.WithError(err).Error("Failed to parse request body")
		return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	// Update fields jika ada
	if req.Nama != "" {
		mahasantri.Nama = req.Nama
	}
	if req.NIM != "" {
		// Cek apakah NIM baru sudah digunakan
		var existingMahasantri models.Mahasantri
		if err := s.DB.Where("nim = ? AND id != ?", req.NIM, mahasantri.ID).First(&existingMahasantri).Error; err == nil {
			return utils.ResponseError(c, fiber.StatusConflict, "NIM already used by another mahasantri", nil)
		}
		mahasantri.NIM = req.NIM
	}
	if req.Jurusan != "" {
		mahasantri.Jurusan = req.Jurusan
	}
	if req.Gender != "" {
		mahasantri.Gender = req.Gender
	}
	if req.MentorID != 0 {
		// Cek apakah mentor ID valid
		var mentor models.Mentor
		if err := s.DB.First(&mentor, req.MentorID).Error; err != nil {
			return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid mentor ID", nil)
		}
		mahasantri.MentorID = req.MentorID
	}

	if err := s.DB.Save(&mahasantri).Error; err != nil {
		logrus.WithError(err).Error("Failed to update mahasantri")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to update mahasantri", err.Error())
	}

	// Get mentor name
	var mentor models.Mentor
	s.DB.First(&mentor, mahasantri.MentorID)

	logrus.WithFields(logrus.Fields{
		"mahasantri_id": mahasantri.ID,
	}).Info("Mahasantri updated by admin")

	return utils.SuccessResponse(c, fiber.StatusOK, "Mahasantri updated successfully", dto.AdminMahasantriResponse{
		ID:         mahasantri.ID,
		Nama:       mahasantri.Nama,
		NIM:        mahasantri.NIM,
		Jurusan:    mahasantri.Jurusan,
		Gender:     mahasantri.Gender,
		MentorID:   mahasantri.MentorID,
		MentorNama: mentor.Nama,
	})
}

// DeleteMahasantri godoc
// @Summary Delete Mahasantri (Admin)
// @Description Menghapus mahasantri oleh admin
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Mahasantri ID"
// @Success 200 {object} utils.SuccessResponseSwagger
// @Failure 404 {object} utils.ErrorResponseSwagger
// @Router /api/v1/admin/mahasantri/{id} [delete]
func (s *AdminService) DeleteMahasantri(c *fiber.Ctx) error {
	id := c.Params("id")

	var mahasantri models.Mahasantri
	if err := s.DB.First(&mahasantri, id).Error; err != nil {
		logrus.Warn("Mahasantri not found: ", id)
		return utils.ResponseError(c, fiber.StatusNotFound, "Mahasantri not found", nil)
	}

	if err := s.DB.Delete(&mahasantri).Error; err != nil {
		logrus.WithError(err).Error("Failed to delete mahasantri")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to delete mahasantri", err.Error())
	}

	logrus.WithFields(logrus.Fields{
		"mahasantri_id": mahasantri.ID,
		"nama":          mahasantri.Nama,
	}).Info("Mahasantri deleted by admin")

	return utils.SuccessResponse(c, fiber.StatusOK, "Mahasantri deleted successfully", nil)
}
