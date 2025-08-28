package services

import (
	"math"
	"strconv"

	"github.com/habbazettt/mahad-service-go/dto"
	"github.com/habbazettt/mahad-service-go/models"
	"github.com/habbazettt/mahad-service-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type JadwalPersonalService interface {
	GetJadwalPersonal(c *fiber.Ctx) error
	CreateJadwalPersonal(c *fiber.Ctx) error
	UpdateJadwalPersonal(c *fiber.Ctx) error
	GetAllJadwalPersonal(c *fiber.Ctx) error
}

type jadwalPersonalService struct {
	DB *gorm.DB
}

func NewJadwalPersonalService(db *gorm.DB) JadwalPersonalService {
	return &jadwalPersonalService{DB: db}
}

// GetAllJadwalPersonal - Mengambil semua jadwal personal dengan pagination (Hanya untuk Mentor)
// @Summary Mengambil semua jadwal personal
// @Description Endpoint untuk mengambil daftar semua jadwal personal yang telah dibuat oleh pengguna, dengan pagination.
// @Tags Jadwal Personal
// @Accept json
// @Produce json
// @Param page query int false "Nomor halaman" default(1)
// @Param limit query int false "Jumlah data per halaman" default(10)
// @Param kesibukan query string false "Filter berdasarkan kesibukan"
// @Success 200 {object} utils.Response "Daftar jadwal personal berhasil diambil"
// @Failure 400 {object} utils.Response "Permintaan tidak valid"
// @Failure 403 {object} utils.Response "Tidak terautentikasi"
// @Failure 404 {object} utils.Response "Jadwal personal tidak ditemukan"
// @Failure 500 {object} utils.Response "Gagal mengambil data"
// @Security BearerAuth
// @Router /api/v1/jadwal-personal/all [get]
func (s *jadwalPersonalService) GetAllJadwalPersonal(c *fiber.Ctx) error {
	log := logrus.WithField("handler", "GetAllJadwalPersonal")
	log.Info("Menerima permintaan untuk mengambil semua jadwal personal")

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	kesibukan := c.Query("kesibukan", "")
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	var jadwalPersonals []models.JadwalPersonal
	var totalJadwals int64

	query := s.DB.Model(&models.JadwalPersonal{})
	if kesibukan != "" {
		query = query.Where("kesibukan ILIKE ?", "%"+kesibukan+"%")
	}

	if err := query.Count(&totalJadwals).Error; err != nil {
		log.WithError(err).Error("Gagal menghitung total jadwal personal")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal memproses permintaan", err.Error())
	}

	if err := query.Preload("Mahasantri").Preload("Mentor").
		Order("updated_at DESC").Limit(limit).Offset(offset).
		Find(&jadwalPersonals).Error; err != nil {
		log.WithError(err).Error("Gagal mengambil daftar jadwal personal")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal mengambil data", err.Error())
	}

	responseDTOs := make([]dto.JadwalPersonalDetailResponse, len(jadwalPersonals))
	for i, jadwal := range jadwalPersonals {
		ownerName := "N/A"
		ownerRole := "N/A"

		if jadwal.MahasantriID != nil && jadwal.Mahasantri != nil {
			ownerName = jadwal.Mahasantri.Nama
			ownerRole = "mahasantri"
		} else if jadwal.MentorID != nil && jadwal.Mentor != nil {
			ownerName = jadwal.Mentor.Nama
			ownerRole = "mentor"
		}

		responseDTOs[i] = dto.JadwalPersonalDetailResponse{
			ID:                jadwal.ID,
			OwnerName:         ownerName,
			OwnerRole:         ownerRole,
			TotalHafalan:      jadwal.TotalHafalan,
			Jadwal:            jadwal.Jadwal,
			Kesibukan:         jadwal.Kesibukan,
			EfektifitasJadwal: jadwal.EfektifitasJadwal,
			UpdatedAt:         jadwal.UpdatedAt,
		}
	}

	log.WithFields(logrus.Fields{
		"page":  page,
		"limit": limit,
	}).Info("Berhasil mengambil semua jadwal personal dengan pagination")

	return utils.SuccessResponse(c, fiber.StatusOK, "Semua jadwal personal berhasil diambil", fiber.Map{
		"pagination": fiber.Map{
			"current_page": page,
			"total_data":   totalJadwals,
			"total_pages":  int(math.Ceil(float64(totalJadwals) / float64(limit))),
		},
		"jadwal_personals": responseDTOs,
	})
}

func (s *jadwalPersonalService) CreateJadwalPersonal(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)
	userID := claims.ID
	userRole := claims.Role

	log := logrus.WithFields(logrus.Fields{"userID": userID, "userRole": userRole})

	var req dto.CreateJadwalPersonalRequest
	if err := c.BodyParser(&req); err != nil {
		log.WithError(err).Warn("Gagal mem-parsing request body untuk jadwal personal")
		return utils.ResponseError(c, fiber.StatusBadRequest, "Request body tidak valid", err.Error())
	}

	err := s.DB.Transaction(func(tx *gorm.DB) error {
		jadwalPersonal := models.JadwalPersonal{
			TotalHafalan:      req.TotalHafalan,
			Jadwal:            req.Jadwal,
			Kesibukan:         req.Kesibukan,
			EfektifitasJadwal: req.EfektifitasJadwal,
		}

		var foreignKeyColumn string
		switch userRole {
		case "mahasantri":
			foreignKeyColumn = "mahasantri_id"
			jadwalPersonal.MahasantriID = &userID
		case "mentor":
			foreignKeyColumn = "mentor_id"
			jadwalPersonal.MentorID = &userID
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: foreignKeyColumn}},
			DoUpdates: clause.AssignmentColumns([]string{"jadwal", "kesibukan", "efektifitas_jadwal", "updated_at"}),
		}).Create(&jadwalPersonal).Error; err != nil {
			return err
		}

		switch userRole {
		case "mahasantri":
			if err := tx.Model(&models.Mahasantri{}).Where("id = ?", userID).Update("is_data_murojaah_filled", true).Error; err != nil {
				return err
			}
		case "mentor":
			if err := tx.Model(&models.Mentor{}).Where("id = ?", userID).Update("is_data_murojaah_filled", true).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		log.WithError(err).Error("Gagal menyimpan jadwal personal ke database (transaksi gagal)")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal menyimpan jadwal", err.Error())
	}

	var finalJadwal models.JadwalPersonal
	if userRole == "mahasantri" {
		s.DB.Where("mahasantri_id = ?", userID).First(&finalJadwal)
	} else {
		s.DB.Where("mentor_id = ?", userID).First(&finalJadwal)
	}

	log.Info("Jadwal personal berhasil disimpan/diperbarui dan status pengguna diperbarui")

	response := dto.JadwalPersonalResponse{
		ID:                finalJadwal.ID,
		MahasantriID:      finalJadwal.MahasantriID,
		MentorID:          finalJadwal.MentorID,
		TotalHafalan:      finalJadwal.TotalHafalan,
		Jadwal:            finalJadwal.Jadwal,
		Kesibukan:         finalJadwal.Kesibukan,
		EfektifitasJadwal: finalJadwal.EfektifitasJadwal,
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Jadwal personal berhasil disimpan", response)
}

// GetJadwalPersonal mengambil jadwal personal milik pengguna yang sedang login.
func (s *jadwalPersonalService) GetJadwalPersonal(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)
	userID := claims.ID
	userRole := claims.Role

	log := logrus.WithFields(logrus.Fields{"userID": userID, "userRole": userRole})

	var jadwalPersonal models.JadwalPersonal
	var err error

	switch userRole {
	case "mahasantri":
		err = s.DB.Where("mahasantri_id = ?", userID).First(&jadwalPersonal).Error
	case "mentor":
		err = s.DB.Where("mentor_id = ?", userID).First(&jadwalPersonal).Error
	}

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn("Jadwal personal tidak ditemukan")
			return utils.SuccessResponse(c, fiber.StatusOK, "Jadwal personal tidak ditemukan", nil)
		}
		log.WithError(err).Error("Gagal mengambil jadwal personal")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal mengambil jadwal", err.Error())
	}

	response := dto.JadwalPersonalResponse{
		ID:                jadwalPersonal.ID,
		MahasantriID:      jadwalPersonal.MahasantriID,
		MentorID:          jadwalPersonal.MentorID,
		TotalHafalan:      jadwalPersonal.TotalHafalan,
		Jadwal:            jadwalPersonal.Jadwal,
		Kesibukan:         jadwalPersonal.Kesibukan,
		EfektifitasJadwal: jadwalPersonal.EfektifitasJadwal,
	}

	log.Info("Jadwal personal berhasil diambil")
	return utils.SuccessResponse(c, fiber.StatusOK, "Jadwal personal berhasil diambil", response)
}

func (s *jadwalPersonalService) UpdateJadwalPersonal(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)
	userID := claims.ID
	userRole := claims.Role

	log := logrus.WithFields(logrus.Fields{"userID": userID, "userRole": userRole})

	var req dto.UpdateJadwalPersonalRequest
	if err := c.BodyParser(&req); err != nil {
		log.WithError(err).Warn("Gagal mem-parsing request body untuk update jadwal personal")
		return utils.ResponseError(c, fiber.StatusBadRequest, "Request body tidak valid", err.Error())
	}

	var jadwalPersonal models.JadwalPersonal
	query := s.DB
	switch userRole {
	case "mahasantri":
		query = query.Where("mahasantri_id = ?", userID)
	case "mentor":
		query = query.Where("mentor_id = ?", userID)
	}

	if err := query.First(&jadwalPersonal).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn("Jadwal personal tidak ditemukan untuk diupdate")
			return utils.ResponseError(c, fiber.StatusNotFound, "Jadwal personal tidak ditemukan", nil)
		}
		log.WithError(err).Error("Gagal mencari jadwal personal")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal memproses permintaan", err.Error())
	}

	updated := false
	if req.TotalHafalan != nil {
		jadwalPersonal.TotalHafalan = *req.TotalHafalan
		updated = true
	}
	if req.Jadwal != nil {
		jadwalPersonal.Jadwal = *req.Jadwal
		updated = true
	}
	if req.Kesibukan != nil {
		jadwalPersonal.Kesibukan = *req.Kesibukan
		updated = true
	}
	if req.EfektifitasJadwal != nil {
		jadwalPersonal.EfektifitasJadwal = *req.EfektifitasJadwal
		updated = true
	}

	if !updated {
		return utils.ResponseError(c, fiber.StatusBadRequest, "Tidak ada data yang diubah", nil)
	}

	if err := s.DB.Save(&jadwalPersonal).Error; err != nil {
		log.WithError(err).Error("Gagal memperbarui jadwal personal di database")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal memperbarui jadwal", err.Error())
	}

	log.Info("Jadwal personal berhasil diperbarui")

	response := dto.JadwalPersonalResponse{
		ID:                jadwalPersonal.ID,
		MahasantriID:      jadwalPersonal.MahasantriID,
		MentorID:          jadwalPersonal.MentorID,
		TotalHafalan:      jadwalPersonal.TotalHafalan,
		Jadwal:            jadwalPersonal.Jadwal,
		Kesibukan:         jadwalPersonal.Kesibukan,
		EfektifitasJadwal: jadwalPersonal.EfektifitasJadwal,
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Jadwal personal berhasil diperbarui", response)
}
