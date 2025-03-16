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

// RegisterMahasantri menangani registrasi Mahasantri
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

// RegisterMentor menangani registrasi Mentor
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

// LoginMentor menangani login untuk Mentor
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

	token, err := utils.GenerateToken(mentor.ID, RoleMentor)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate token")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to generate token", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Login successful", dto.AuthResponse{
		Token: token,
		User: dto.UserMentorResponse{
			ID:       mentor.ID,
			Nama:     mentor.Nama,
			Email:    mentor.Email,
			Gender:   mentor.Gender,
			UserType: RoleMentor,
		},
	})
}

// LoginMahasantri menangani login untuk Mahasantri
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
			UserType: RoleMahasantri,
		},
	})
}

// GetCurrentUser mendapatkan user yang sedang login
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

		// Format response for Mentor
		response = dto.UserMentorResponse{
			ID:       mentor.ID,
			Nama:     mentor.Nama,
			Email:    mentor.Email,
			Gender:   mentor.Gender,
			UserType: RoleMentor,
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
