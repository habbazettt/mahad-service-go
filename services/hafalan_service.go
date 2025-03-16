package services

import (
	"github.com/gofiber/fiber/v2"
	"github.com/habbazettt/mahad-service-go/dto"
	"github.com/habbazettt/mahad-service-go/models"
	"github.com/habbazettt/mahad-service-go/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type HafalanService struct {
	DB *gorm.DB
}

// NewHafalanService membuat instance service hafalan
func NewHafalanService(db *gorm.DB) *HafalanService {
	return &HafalanService{DB: db}
}

// CreateHafalan menambahkan hafalan baru
func (s *HafalanService) CreateHafalan(c *fiber.Ctx) error {
	var req dto.CreateHafalanRequest
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

	// Simpan Hafalan
	hafalan := models.Hafalan{
		MahasantriID: req.MahasantriID,
		Juz:          req.Juz,
		Halaman:      req.Halaman,
		TotalSetoran: req.TotalSetoran,
		Kategori:     req.Kategori,
		Waktu:        req.Waktu,
		Catatan:      req.Catatan,
	}

	if err := s.DB.Create(&hafalan).Error; err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"mahasantri_id": req.MahasantriID,
			"juz":           req.Juz,
			"halaman":       req.Halaman,
		}).Error("Failed to create hafalan")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to create hafalan", err.Error())
	}

	logrus.WithFields(logrus.Fields{
		"hafalan_id": hafalan.ID,
		"mahasantri": req.MahasantriID,
		"juz":        req.Juz,
		"halaman":    req.Halaman,
	}).Info("Hafalan created successfully")

	return utils.SuccessResponse(c, fiber.StatusCreated, "Hafalan created successfully", hafalan)
}

// GetAllHafalan mengambil semua data hafalan
func (s *HafalanService) GetAllHafalan(c *fiber.Ctx) error {
	var hafalan []models.Hafalan

	if err := s.DB.Find(&hafalan).Error; err != nil {
		logrus.WithError(err).Error("Failed to fetch hafalan")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to fetch hafalan", err.Error())
	}

	logrus.WithField("count", len(hafalan)).Info("Fetched all hafalan successfully")
	return utils.SuccessResponse(c, fiber.StatusOK, "Hafalan fetched successfully", hafalan)
}

// GetHafalanByID mendapatkan hafalan berdasarkan ID
func (s *HafalanService) GetHafalanByID(c *fiber.Ctx) error {
	id := c.Params("id")
	var hafalan models.Hafalan

	if err := s.DB.First(&hafalan, id).Error; err != nil {
		logrus.WithError(err).Warn("Hafalan not found")
		return utils.ResponseError(c, fiber.StatusNotFound, "Hafalan not found", nil)
	}

	logrus.WithFields(logrus.Fields{
		"hafalan_id": hafalan.ID,
	}).Info("Hafalan found")

	return utils.SuccessResponse(c, fiber.StatusOK, "Hafalan found", hafalan)
}

// GetHafalanByMahasantriID mengambil semua hafalan berdasarkan MahasantriID dan menyertakan data Mahasantri
func (s *HafalanService) GetHafalanByMahasantriID(c *fiber.Ctx) error {
	mahasantriID := c.Params("mahasantri_id")

	// Ambil query parameters untuk filtering
	kategori := c.Query("kategori") // Optional filter by kategori
	juz := c.Query("juz")           // Optional filter by juz

	// Ambil data Mahasantri
	var mahasantri models.Mahasantri
	if err := s.DB.First(&mahasantri, mahasantriID).Error; err != nil {
		logrus.WithError(err).WithField("mahasantri_id", mahasantriID).Warn("Mahasantri not found")
		return utils.ResponseError(c, fiber.StatusNotFound, "Mahasantri not found", nil)
	}

	// Ambil semua Hafalan berdasarkan MahasantriID dengan filtering
	query := s.DB.Where("mahasantri_id = ?", mahasantriID)

	// Apply kategori filter jika ada
	if kategori != "" {
		query = query.Where("kategori = ?", kategori)
	}

	// Apply juz filter jika ada
	if juz != "" {
		query = query.Where("juz = ?", juz)
	}

	// Ambil data Hafalan berdasarkan query yang sudah difilter
	var hafalan []models.Hafalan
	if err := query.Find(&hafalan).Error; err != nil {
		logrus.WithError(err).WithField("mahasantri_id", mahasantriID).Error("Failed to fetch hafalan")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to fetch hafalan", err.Error())
	}

	// Inisialisasi variabel total setoran dan kategori
	var totalSetoran float32
	totalPerJuz := make(map[int]float32) // Map untuk menyimpan total setoran per juz
	var totalZiyadah float32
	var totalMurojaah float32

	// Hitung total setoran dan total per kategori
	for _, h := range hafalan {
		totalSetoran += h.TotalSetoran
		totalPerJuz[h.Juz] += h.TotalSetoran

		// Hitung total berdasarkan kategori
		if h.Kategori == "ziyadah" {
			totalZiyadah += h.TotalSetoran
		} else if h.Kategori == "murojaah" {
			totalMurojaah += h.TotalSetoran
		}
	}

	// Konversi totalPerJuz ke bentuk array agar lebih rapi dalam JSON response
	var totalPerJuzArray []fiber.Map
	for juz, total := range totalPerJuz {
		totalPerJuzArray = append(totalPerJuzArray, fiber.Map{
			"juz":           juz,
			"total_setoran": total,
		})
	}

	// Format response dengan data Mahasantri, Hafalan, dan Total Setoran
	response := fiber.Map{
		"mahasantri": fiber.Map{
			"id":      mahasantri.ID,
			"nama":    mahasantri.Nama,
			"nim":     mahasantri.NIM,
			"jurusan": mahasantri.Jurusan,
			"gender":  mahasantri.Gender,
		},
		"hafalan":       hafalan,
		"total_setoran": totalSetoran,
		"total_per_juz": totalPerJuzArray,
		"total_per_kategori": fiber.Map{
			"ziyadah":  totalZiyadah,
			"murojaah": totalMurojaah,
		},
	}

	logrus.WithField("mahasantri_id", mahasantriID).Info("Fetched hafalan with total setoran successfully")
	return utils.SuccessResponse(c, fiber.StatusOK, "Hafalan fetched successfully", response)
}

func (s *HafalanService) GetHafalanByKategori(c *fiber.Ctx) error {
	mahasantriID := c.Params("mahasantri_id")
	kategori := c.Query("kategori")

	if kategori != "ziyadah" && kategori != "murojaah" {
		return utils.ResponseError(c, fiber.StatusBadRequest, "Kategori harus 'ziyadah' atau 'murojaah'", nil)
	}

	var mahasantri models.Mahasantri
	if err := s.DB.First(&mahasantri, mahasantriID).Error; err != nil {
		logrus.WithError(err).WithField("mahasantri_id", mahasantriID).Warn("Mahasantri not found")
		return utils.ResponseError(c, fiber.StatusNotFound, "Mahasantri not found", nil)
	}

	var hafalan []models.Hafalan
	if err := s.DB.Where("mahasantri_id = ? AND kategori = ?", mahasantriID, kategori).Find(&hafalan).Error; err != nil {
		logrus.WithError(err).WithField("mahasantri_id", mahasantriID).Error("Failed to fetch hafalan")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to fetch hafalan", err.Error())
	}

	var totalSetoran float32
	for _, h := range hafalan {
		totalSetoran += h.TotalSetoran
	}

	response := fiber.Map{
		"mahasantri": fiber.Map{
			"id":      mahasantri.ID,
			"nama":    mahasantri.Nama,
			"nim":     mahasantri.NIM,
			"jurusan": mahasantri.Jurusan,
			"gender":  mahasantri.Gender,
		},
		"kategori":      kategori,
		"hafalan":       hafalan,
		"total_setoran": totalSetoran,
	}

	logrus.WithFields(logrus.Fields{
		"mahasantri_id": mahasantriID,
		"kategori":      kategori,
	}).Info("Fetched hafalan by category successfully")

	return utils.SuccessResponse(c, fiber.StatusOK, "Hafalan by category fetched successfully", response)
}

// UpdateHafalan memperbarui data hafalan
func (s *HafalanService) UpdateHafalan(c *fiber.Ctx) error {
	id := c.Params("id")
	var hafalan models.Hafalan

	if err := s.DB.First(&hafalan, id).Error; err != nil {
		logrus.WithField("hafalan_id", id).Warn("Hafalan not found")
		return utils.ResponseError(c, fiber.StatusNotFound, "Hafalan not found", nil)
	}

	var req dto.UpdateHafalanRequest
	if err := c.BodyParser(&req); err != nil {
		logrus.WithError(err).Error("Invalid request body")
		return utils.ResponseError(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	updated := false
	updateFields := logrus.Fields{"hafalan_id": id}

	if req.Juz != nil && *req.Juz != hafalan.Juz {
		hafalan.Juz = *req.Juz
		updateFields["juz"] = *req.Juz
		updated = true
	}
	if req.Halaman != nil && *req.Halaman != hafalan.Halaman {
		hafalan.Halaman = *req.Halaman
		updateFields["halaman"] = *req.Halaman
		updated = true
	}
	if req.TotalSetoran != nil && *req.TotalSetoran != hafalan.TotalSetoran {
		hafalan.TotalSetoran = *req.TotalSetoran
		updateFields["total_setoran"] = *req.TotalSetoran
		updated = true
	}
	if req.Kategori != nil && *req.Kategori != hafalan.Kategori {
		hafalan.Kategori = *req.Kategori
		updateFields["kategori"] = *req.Kategori
		updated = true
	}
	if req.Waktu != nil && *req.Waktu != hafalan.Waktu {
		hafalan.Waktu = *req.Waktu
		updateFields["waktu"] = *req.Waktu
		updated = true
	}
	if req.Catatan != nil && *req.Catatan != hafalan.Catatan {
		hafalan.Catatan = *req.Catatan
		updateFields["catatan"] = *req.Catatan
		updated = true
	}

	if !updated {
		logrus.WithField("hafalan_id", id).Warn("No changes detected")
		return utils.ResponseError(c, fiber.StatusBadRequest, "No changes detected", nil)
	}

	if err := s.DB.Save(&hafalan).Error; err != nil {
		logrus.WithError(err).WithFields(updateFields).Error("Failed to update hafalan")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to update hafalan", err.Error())
	}

	logrus.WithFields(updateFields).Info("Hafalan updated successfully")
	return utils.SuccessResponse(c, fiber.StatusOK, "Hafalan updated successfully", hafalan)
}

// DeleteHafalan menghapus hafalan berdasarkan ID
func (s *HafalanService) DeleteHafalan(c *fiber.Ctx) error {
	id := c.Params("id")
	var hafalan models.Hafalan

	if err := s.DB.First(&hafalan, id).Error; err != nil {
		logrus.WithField("hafalan_id", id).Warn("Hafalan not found")
		return utils.ResponseError(c, fiber.StatusNotFound, "Hafalan not found", nil)
	}

	if err := s.DB.Delete(&hafalan).Error; err != nil {
		logrus.WithError(err).WithField("hafalan_id", id).Error("Failed to delete hafalan")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Failed to delete hafalan", err.Error())
	}

	logrus.WithField("hafalan_id", id).Info("Hafalan deleted successfully")
	return utils.SuccessResponse(c, fiber.StatusOK, "Hafalan deleted successfully", nil)
}
