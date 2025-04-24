package services

import (
	"github.com/gofiber/fiber/v2"
	"github.com/habbazettt/mahad-service-go/dto"
	"github.com/habbazettt/mahad-service-go/models"
	"github.com/habbazettt/mahad-service-go/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	RoleMentor     = "mentor"
	RoleMahasantri = "mahasantri"
)

// AuthService menangani logika autentikasi pengguna
type AuthService struct {
	DB *gorm.DB
}

// RegisterMahasantri godoc
// @Summary Register Mahasantri
// @Description Mendaftarkan akun Mahasantri baru
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body dto.RegisterMahasantriRequest true "Data pendaftaran Mahasantri"
// @Success 201 {object} utils.SuccessExample
// @Failure 400 {object} utils.ErrorExample
// @Failure 409 {object} utils.ErrorExample
// @Router /api/v1/auth/register/mahasantri [post]
func (s *AuthService) RegisterMahasantri(c *fiber.Ctx) error {
	var req dto.RegisterMahasantriRequest

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

	// Simpan data Mahasantri
	mahasantri := models.Mahasantri{
		Nama:     req.Nama,
		NIM:      req.NIM,
		Jurusan:  req.Jurusan,
		Gender:   req.Gender,
		Password: hashedPassword,
		MentorID: req.MentorID,
	}

	if err := s.DB.Create(&mahasantri).Error; err != nil {
		logrus.WithError(err).Error("Failed to register mahasantri")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to register mahasantri", err.Error())
	}

	logrus.WithFields(logrus.Fields{
		"user_id":  mahasantri.ID,
		"nama":     mahasantri.Nama,
		"nim":      mahasantri.NIM,
		"jurusan":  mahasantri.Jurusan,
		"gender":   mahasantri.Gender,
		"mentorID": mahasantri.MentorID,
	}).Info("Mahasantri registered successfully")

	return utils.SuccessResponse(c, fiber.StatusCreated, "Mahasantri registered successfully", fiber.Map{
		"id":       mahasantri.ID,
		"nama":     mahasantri.Nama,
		"nim":      mahasantri.NIM,
		"jurusan":  mahasantri.Jurusan,
		"gender":   mahasantri.Gender,
		"mentorID": mahasantri.MentorID,
	})
}

// RegisterMentor godoc
// @Summary Register Mentor
// @Description Mendaftarkan akun Mentor baru
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body dto.RegisterMentorRequest true "Data pendaftaran Mentor"
// @Success 201 {object} utils.SuccessMentorExample
// @Failure 400 {object} utils.ErrorMentorExample
// @Failure 409 {object} utils.ErrorMentorExample
// @Router /api/v1/auth/register/mentor [post]
func (s *AuthService) RegisterMentor(c *fiber.Ctx) error {
	var req dto.RegisterMentorRequest

	if err := c.BodyParser(&req); err != nil {
		logrus.WithError(err).Error("Failed to parse request body")
		return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	var existingMentor models.Mentor
	if err := s.DB.Where("email = ?", req.Email).First(&existingMentor).Error; err == nil {
		logrus.Warn("Email already registered: ", req.Email)
		return utils.ResponseError(c, fiber.StatusConflict, "Email already registered", nil)
	}

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
		logrus.WithError(err).Error("Failed to register mentor")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to register mentor", err.Error())
	}

	logrus.WithFields(logrus.Fields{
		"user_id": mentor.ID,
		"nama":    mentor.Nama,
		"email":   mentor.Email,
		"gender":  mentor.Gender,
	}).Info("Mentor registered successfully")

	return utils.SuccessResponse(c, fiber.StatusCreated, "Mentor registered successfully", fiber.Map{
		"id":     mentor.ID,
		"nama":   mentor.Nama,
		"email":  mentor.Email,
		"gender": mentor.Gender,
	})
}

// LoginMentor godoc
// @Summary Login Mentor
// @Description Melakukan login untuk mentor dengan email dan password
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body dto.LoginMentorRequest true "Data login Mentor"
// @Success 200 {object} utils.SuccessLoginMentorExample
// @Failure 400 {object} utils.ErrorLoginMentorExample
// @Failure 401 {object} utils.ErrorLoginMentorExample
// @Router /api/v1/auth/login/mentor [post]
func (s *AuthService) LoginMentor(c *fiber.Ctx) error {
	var req dto.LoginMentorRequest

	if err := c.BodyParser(&req); err != nil {
		logrus.WithError(err).Error("Failed to parse request body")
		return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	var mentor models.Mentor
	if err := s.DB.Where("email = ?", req.Email).First(&mentor).Error; err != nil {
		logrus.Warn("Invalid email or password: ", req.Email)
		return utils.ResponseError(c, fiber.StatusUnauthorized, "Invalid email or password", nil)
	}

	if !utils.ComparePassword(mentor.Password, req.Password) {
		logrus.Warn("Invalid password for email: ", req.Email)
		return utils.ResponseError(c, fiber.StatusUnauthorized, "Invalid email or password", nil)
	}

	// Hitung jumlah Mahasantri yang terkait dengan mentor ini
	var mahasantriCount int64
	if err := s.DB.Model(&models.Mahasantri{}).Where("mentor_id = ?", mentor.ID).Count(&mahasantriCount).Error; err != nil {
		logrus.WithError(err).Error("Failed to count mahasantri")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to fetch mahasantri count", err.Error())
	}

	// Generate token
	token, err := utils.GenerateToken(mentor.ID, RoleMentor)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate token")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to generate token", err.Error())
	}

	// Return response dengan menambahkan field mahasantri_count
	return utils.SuccessResponse(c, fiber.StatusOK, "Login successful", dto.AuthResponse{
		Token: token,
		User: dto.UserMentorResponse{
			ID:              mentor.ID,
			Nama:            mentor.Nama,
			Email:           mentor.Email,
			Gender:          mentor.Gender,
			UserType:        RoleMentor,
			MahasantriCount: int(mahasantriCount), // Tambahkan jumlah Mahasantri
		},
	})
}

// LoginMahasantri godoc
// @Summary Login Mahasantri
// @Description Melakukan login untuk mahasantri dengan NIM dan password
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body dto.LoginMahasantriRequest true "Data login Mahasantri"
// @Success 200 {object} utils.SuccessLoginMahasantriExample
// @Failure 400 {object} utils.ErrorLoginMahasantriExample
// @Failure 401 {object} utils.ErrorLoginMahasantriExample
// @Router /api/v1/auth/login/mahasantri [post]
func (s *AuthService) LoginMahasantri(c *fiber.Ctx) error {
	var req dto.LoginMahasantriRequest

	if err := c.BodyParser(&req); err != nil {
		logrus.WithError(err).Error("Failed to parse request body")
		return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	var mahasantri models.Mahasantri
	if err := s.DB.Where("nim = ?", req.NIM).First(&mahasantri).Error; err != nil {
		logrus.Warn("Invalid NIM or password: ", req.NIM)
		return utils.ResponseError(c, fiber.StatusUnauthorized, "Invalid NIM or password", nil)
	}

	if !utils.ComparePassword(mahasantri.Password, req.Password) {
		logrus.Warn("Invalid password for NIM: ", req.NIM)
		return utils.ResponseError(c, fiber.StatusUnauthorized, "Invalid NIM or password", nil)
	}

	token, err := utils.GenerateToken(mahasantri.ID, RoleMahasantri)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate token")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to generate token", err.Error())
	}

	logrus.WithFields(logrus.Fields{
		"user_id": mahasantri.ID,
		"nim":     mahasantri.NIM,
	}).Info("Mahasantri logged in successfully")

	return utils.SuccessResponse(c, fiber.StatusOK, "Login successful", dto.AuthResponse{
		Token: token,
		User: dto.UserMahasantriResponse{
			ID:       mahasantri.ID,
			Nama:     mahasantri.Nama,
			NIM:      mahasantri.NIM,
			Jurusan:  mahasantri.Jurusan,
			Gender:   mahasantri.Gender,
			MentorID: mahasantri.MentorID,
			UserType: RoleMahasantri,
		},
	})
}

// ForgotPassword godoc
// @Summary Forgot Password
// @Description Endpoint untuk mengupdate password untuk Mahasantri atau Mentor berdasarkan NIM atau Email
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body dto.ForgotPasswordRequest true "Data untuk reset password"
// @Success 200 {object} utils.SuccessExample
// @Failure 400 {object} utils.ErrorExample
// @Failure 404 {object} utils.ErrorExample
// @Failure 500 {object} utils.ErrorExample
// @Router /api/v1/auth/forgot-password [post]
func (s *AuthService) ForgotPassword(c *fiber.Ctx) error {
	var req dto.ForgotPasswordRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if req.Email == "" && req.NIM == "" {
		return utils.ResponseError(c, fiber.StatusBadRequest, "Email or NIM is required", nil)
	}

	hashed, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to hash password", err.Error())
	}

	if req.Email != "" {
		var mentor models.Mentor
		if err := s.DB.Where("email = ?", req.Email).First(&mentor).Error; err != nil {
			return utils.ResponseError(c, fiber.StatusNotFound, "Mentor not found", nil)
		}
		mentor.Password = hashed
		s.DB.Save(&mentor)
		return utils.SuccessResponse(c, fiber.StatusOK, "Password updated successfully", map[string]interface{}{
			"email":        mentor.Email,
			"new_password": req.NewPassword, // Kembalikan password baru
		})
	}

	if req.NIM != "" {
		var mahasantri models.Mahasantri
		if err := s.DB.Where("nim = ?", req.NIM).First(&mahasantri).Error; err != nil {
			return utils.ResponseError(c, fiber.StatusNotFound, "Mahasantri not found", nil)
		}
		mahasantri.Password = hashed
		s.DB.Save(&mahasantri)
		return utils.SuccessResponse(c, fiber.StatusOK, "Password updated successfully", map[string]interface{}{
			"nim":          mahasantri.NIM,
			"new_password": req.NewPassword, // Kembalikan password baru
		})
	}

	return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid request", nil)
}

// Logout godoc
// @Summary Logout
// @Description Endpoint untuk logout dan menghapus token
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} utils.SuccessExample
// @Failure 400 {object} utils.ErrorExample
// @Router /api/v1/auth/logout [post]
func (s *AuthService) Logout(c *fiber.Ctx) error {
	// Mendapatkan claims dari token JWT yang sudah terverifikasi
	claims := c.Locals("user").(*utils.Claims)

	// Logging untuk mencatat bahwa pengguna berhasil logout
	logrus.WithFields(logrus.Fields{
		"user_id": claims.ID,
		"role":    claims.Role,
	}).Info("User logged out successfully")

	// Menghapus token hanya dilakukan di sisi klien (frontend)
	// Pada sisi server, tidak ada sesi atau data yang perlu dihapus
	// Karena JWT tidak disimpan di server, maka tidak ada aksi yang perlu dilakukan di backend selain mengirim response

	return utils.SuccessResponse(c, fiber.StatusOK, "Successfully logged out", map[string]interface{}{
		"user_id": claims.ID,
		"role":    claims.Role,
	})
}

// GetCurrentUser godoc
// @Summary Get current user data
// @Description Mengambil data user yang sedang login (baik Mentor atau Mahasantri)
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessGetCurrentUserExample
// @Failure 401 {object} utils.ErrorGetCurrentUserExample
// @Failure 404 {object} utils.ErrorGetCurrentUserExample
// @Router /api/v1/auth/me [get]
func (s *AuthService) GetCurrentUser(c *fiber.Ctx) error {
	userClaims, ok := c.Locals("user").(*utils.Claims)
	if !ok || userClaims == nil {
		logrus.Warn("Unauthorized access: Missing user claims")
		return utils.ResponseError(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	var response interface{}
	switch userClaims.Role {
	case RoleMentor:
		var mentor models.Mentor
		// Get the mentor's basic data without preloading other relations
		if err := s.DB.First(&mentor, userClaims.ID).Error; err != nil {
			logrus.Warn("Mentor not found: ", userClaims.ID)
			return utils.ResponseError(c, fiber.StatusNotFound, "User not found", nil)
		}

		// Hitung jumlah Mahasantri yang terkait dengan mentor ini
		var mahasantriCount int64
		if err := s.DB.Model(&models.Mahasantri{}).Where("mentor_id = ?", mentor.ID).Count(&mahasantriCount).Error; err != nil {
			logrus.WithError(err).Error("Failed to count mahasantri")
			return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to fetch mahasantri count", err.Error())
		}

		// Format response for Mentor
		response = dto.UserMentorResponse{
			ID:              mentor.ID,
			Nama:            mentor.Nama,
			Email:           mentor.Email,
			Gender:          mentor.Gender,
			UserType:        RoleMentor,
			MahasantriCount: int(mahasantriCount), // Tambahkan jumlah Mahasantri

		}

	case RoleMahasantri:
		var mahasantri models.Mahasantri
		// Get the mahasantri's basic data without preloading other relations
		if err := s.DB.First(&mahasantri, userClaims.ID).Error; err != nil {
			logrus.Warn("Mahasantri not found: ", userClaims.ID)
			return utils.ResponseError(c, fiber.StatusNotFound, "User not found", nil)
		}

		// Format response for Mahasantri
		response = dto.UserMahasantriResponse{
			ID:       mahasantri.ID,
			Nama:     mahasantri.Nama,
			NIM:      mahasantri.NIM,
			Jurusan:  mahasantri.Jurusan,
			Gender:   mahasantri.Gender,
			MentorID: mahasantri.MentorID,
			UserType: RoleMahasantri,
		}

	default:
		logrus.Warn("Unauthorized access: Invalid role")
		return utils.ResponseError(c, fiber.StatusUnauthorized, "Invalid user role", nil)
	}

	logrus.WithFields(logrus.Fields{
		"user_id": userClaims.ID,
		"role":    userClaims.Role,
	}).Info("User data retrieved successfully")

	return utils.SuccessResponse(c, fiber.StatusOK, "User data retrieved", response)
}
