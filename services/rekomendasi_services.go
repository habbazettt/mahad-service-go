package services

import (
	"fmt"
	"math"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/habbazettt/mahad-service-go/config"
	"github.com/habbazettt/mahad-service-go/dto"
	"github.com/habbazettt/mahad-service-go/models"
	"github.com/habbazettt/mahad-service-go/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type RekomendasiService interface {
	GetRecommendation(c *fiber.Ctx) error
	GetAllRekomendasi(c *fiber.Ctx) error
}

type rekomendasiService struct {
	DB *gorm.DB
}

func NewRekomendasiService(db *gorm.DB) RekomendasiService {
	return &rekomendasiService{DB: db}
}

// GetRecommendation - Mendapatkan Rekomendasi Jadwal Muroja'ah
// @Summary Mendapatkan rekomendasi jadwal muroja'ah
// @Description Endpoint ini menghasilkan rekomendasi jadwal muroja'ah yang dipersonalisasi berdasarkan kondisi pengguna (kesibukan dan kategori hafalan).
// @Tags Rekomendasi
// @Accept json
// @Produce json
// @Param rekomendasiRequest body dto.RecommendationRequest true "Data kondisi pengguna untuk menghasilkan rekomendasi"
// @Success 200 {object} utils.Response "Rekomendasi berhasil dibuat"
// @Failure 400 {object} utils.Response "Request body tidak valid"
// @Failure 401 {object} utils.Response "Tidak terautentikasi (token tidak valid)"
// @Failure 403 {object} utils.Response "Tidak memiliki hak akses (role tidak sesuai)"
// @Security BearerAuth
// @Router /api/v1/rekomendasi [post]
func (s *rekomendasiService) GetRecommendation(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)

	log := logrus.WithFields(logrus.Fields{
		"handler":  "GetRecommendation",
		"userID":   claims.ID,
		"userRole": claims.Role,
	})
	log.Info("Menerima permintaan rekomendasi jadwal")

	var req dto.RecommendationRequest
	if err := c.BodyParser(&req); err != nil {
		log.WithError(err).Error("Gagal mem-parsing request body")
		return utils.ResponseError(c, fiber.StatusBadRequest, "Cannot parse request body", err.Error())
	}

	stateString := fmt.Sprintf("%s_%s", req.Kesibukan, req.KategoriHafalan)
	log = log.WithField("state", stateString)

	var bestAction string
	var qValue *float64
	var recType string
	var persentaseEfektif *float64

	if stateActions, ok := config.QTableModel[stateString]; ok {
		var maxQ float64 = -1.0
		isFirst := true
		for action, val := range stateActions {
			if isFirst || val > maxQ {
				maxQ = val
				bestAction = action
				isFirst = false
			}
		}
		qValue = &maxQ
		recType = "Spesifik"
	} else {
		if len(config.HistoricalBest) > 0 {
			bestAction = config.HistoricalBest[0].Jadwal
			recType = "Umum (Historis Terbaik)"
		} else {
			bestAction = "Tidak ada jadwal default"
			recType = "Tidak Ada Rekomendasi"
		}
	}

	if bestAction != "Tidak ada jadwal default" && len(config.HistoricalBest) > 0 {
		for _, info := range config.HistoricalBest {
			if info.Jadwal == bestAction {
				persen := info.PersentaseEfektif
				persentaseEfektif = &persen
				break
			}
		}
	}

	log = log.WithFields(logrus.Fields{
		"rekomendasi": bestAction,
		"tipe":        recType,
	})

	response := dto.RecommendationResponse{
		State:                     stateString,
		RekomendasiJadwal:         bestAction,
		TipeRekomendasi:           recType,
		EstimasiQValue:            qValue,
		PersentaseEfektifHistoris: persentaseEfektif,
	}

	if bestAction != "Tidak ada jadwal default" {
		rekomendasiRecord := models.JadwalRekomendasi{
			State:             stateString,
			RekomendasiJadwal: response.RekomendasiJadwal,
			TipeRekomendasi:   response.TipeRekomendasi,
			EstimasiQValue:    response.EstimasiQValue,
		}

		if claims.Role == "mahasantri" {
			rekomendasiRecord.MahasantriID = &claims.ID
		} else if claims.Role == "mentor" {
			rekomendasiRecord.MentorID = &claims.ID
		}

		if err := s.DB.Create(&rekomendasiRecord).Error; err != nil {
			log.WithError(err).Error("Gagal menyimpan riwayat rekomendasi ke database")
		} else {
			log.WithField("recordID", rekomendasiRecord.ID).Info("Riwayat rekomendasi berhasil disimpan")
			response.ID = rekomendasiRecord.ID
		}
	}

	log.Info("Rekomendasi berhasil dikirim ke pengguna")
	return utils.SuccessResponse(c, fiber.StatusOK, "Rekomendasi berhasil dibuat", response)
}

// GetAllRekomendasi - Mengambil riwayat rekomendasi untuk pengguna dengan pagination
// @Summary Mengambil riwayat rekomendasi dengan pagination
// @Description Endpoint untuk mengambil riwayat rekomendasi jadwal yang pernah diberikan kepada pengguna yang sedang login.
// @Tags Rekomendasi
// @Accept json
// @Produce json
// @Param page query int false "Nomor halaman" default(1)
// @Param limit query int false "Jumlah data per halaman" default(10)
// @Success 200 {object} utils.Response "Daftar riwayat rekomendasi berhasil diambil"
// @Failure 401 {object} utils.Response "Tidak terautentikasi (token tidak valid)"
// @Failure 403 {object} utils.Response "Tidak memiliki hak akses (role tidak sesuai)"
// @Failure 500 {object} utils.Response "Gagal mengambil riwayat rekomendasi"
// @Security BearerAuth
// @Router /api/v1/rekomendasi [get]
func (s *rekomendasiService) GetAllRekomendasi(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)
	userID := claims.ID
	userRole := claims.Role

	log := logrus.WithFields(logrus.Fields{
		"handler":  "GetAllRekomendasi",
		"userID":   userID,
		"userRole": userRole,
	})
	log.Info("Menerima permintaan untuk mengambil riwayat rekomendasi")

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	var riwayatRekomendasi []models.JadwalRekomendasi
	var totalRiwayat int64

	query := s.DB.Model(&models.JadwalRekomendasi{})
	if userRole == "mahasantri" {
		query = query.Where("mahasantri_id = ?", userID)
	} else if userRole == "mentor" {
		query = query.Where("mentor_id = ?", userID)
	}

	if err := query.Count(&totalRiwayat).Error; err != nil {
		log.WithError(err).Error("Gagal menghitung total riwayat rekomendasi")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal menghitung total data", err.Error())
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&riwayatRekomendasi).Error; err != nil {
		log.WithError(err).Error("Gagal mengambil riwayat rekomendasi dari database")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal mengambil riwayat rekomendasi", err.Error())
	}

	responseDTOs := make([]dto.RecommendationResponse, len(riwayatRekomendasi))
	for i, rec := range riwayatRekomendasi {
		var persentaseEfektif *float64
		for _, info := range config.HistoricalBest {
			if info.Jadwal == rec.RekomendasiJadwal {
				persen := info.PersentaseEfektif
				persentaseEfektif = &persen
				break
			}
		}
		responseDTOs[i] = dto.RecommendationResponse{
			ID:                        rec.ID,
			State:                     rec.State,
			RekomendasiJadwal:         rec.RekomendasiJadwal,
			TipeRekomendasi:           rec.TipeRekomendasi,
			EstimasiQValue:            rec.EstimasiQValue,
			PersentaseEfektifHistoris: persentaseEfektif,
		}
	}

	log.WithFields(logrus.Fields{
		"page":       page,
		"limit":      limit,
		"total_data": totalRiwayat,
	}).Info("Berhasil mengambil riwayat rekomendasi dengan pagination")

	return utils.SuccessResponse(c, fiber.StatusOK, "Riwayat rekomendasi berhasil diambil", fiber.Map{
		"pagination": fiber.Map{
			"current_page": page,
			"total_data":   totalRiwayat,
			"total_pages":  int(math.Ceil(float64(totalRiwayat) / float64(limit))),
		},
		"riwayat_rekomendasi": responseDTOs,
	})
}
