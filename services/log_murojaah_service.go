package services

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/habbazettt/mahad-service-go/dto"
	"github.com/habbazettt/mahad-service-go/models"
	"github.com/habbazettt/mahad-service-go/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type LogMurojaahService interface {
	GetOrCreateLogHarian(c *fiber.Ctx) error
	AddDetailToLog(c *fiber.Ctx) error
	UpdateDetailLog(c *fiber.Ctx) error
	DeleteDetailLog(c *fiber.Ctx) error
	GetRecapMingguan(c *fiber.Ctx) error
	GetAllLogsForMentorDashboard(c *fiber.Ctx) error
	GetStatistikMurojaah(c *fiber.Ctx) error
	GetRekapBimbinganMingguan(c *fiber.Ctx) error
	ApplyAIRekomendasi(c *fiber.Ctx) error
}

type logMurojaahService struct {
	DB *gorm.DB
}

func NewLogMurojaahService(db *gorm.DB) LogMurojaahService {
	return &logMurojaahService{DB: db}
}

func calculateTotalPages(startJuz, startHalaman, endJuz, endHalaman int) (int, error) {
	if startJuz > endJuz || (startJuz == endJuz && startHalaman > endHalaman) {
		return 0, errors.New("target/progres akhir tidak boleh lebih kecil dari awal")
	}

	const halamanPerJuz = 20

	if startJuz == endJuz {
		return (endHalaman - startHalaman) + 1, nil
	}

	halamanDiJuzAwal := (halamanPerJuz - startHalaman) + 1

	halamanDiJuzAkhir := endHalaman

	juzPerantara := (endJuz - startJuz) - 1
	halamanDiJuzPerantara := juzPerantara * halamanPerJuz

	return halamanDiJuzAwal + halamanDiJuzAkhir + halamanDiJuzPerantara, nil
}

func (s *logMurojaahService) recalculateTotals(tx *gorm.DB, logHarianID uint) error {
	var totals struct {
		TotalTarget  int
		TotalSelesai int
	}

	err := tx.Model(&models.DetailLog{}).
		Select("COALESCE(SUM(total_target_halaman), 0) as total_target, COALESCE(SUM(total_selesai_halaman), 0) as total_selesai").
		Where("log_harian_id = ?", logHarianID).
		Scan(&totals).Error

	if err != nil {
		return err
	}

	return tx.Model(&models.LogHarian{}).Where("id = ?", logHarianID).Updates(map[string]interface{}{
		"total_target_halaman":  totals.TotalTarget,
		"total_selesai_halaman": totals.TotalSelesai,
	}).Error
}

func (s *logMurojaahService) GetOrCreateLogHarian(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)
	userRole := claims.Role
	userID := claims.ID

	var mahasantriID uint
	var err error

	if userRole == "mahasantri" {
		mahasantriID = userID
	} else if userRole == "mentor" {
		mahasantriID_int, err_parse := strconv.Atoi(c.Params("mahasantriID"))
		if err_parse != nil {
			return utils.ResponseError(c, fiber.StatusBadRequest, "ID Mahasantri tidak valid pada parameter URL", nil)
		}
		mahasantriID = uint(mahasantriID_int)

		var mahasantri models.Mahasantri
		if err := s.DB.Select("mentor_id").First(&mahasantri, mahasantriID).Error; err != nil || mahasantri.MentorID != userID {
			return utils.ResponseError(c, fiber.StatusForbidden, "Anda tidak memiliki hak akses untuk melihat log mahasantri ini", nil)
		}
	}

	log := logrus.WithFields(logrus.Fields{"handler": "GetOrCreateLogHarian", "mahasantriID": mahasantriID})

	tanggalStr := c.Query("tanggal")
	var tanggal time.Time
	if tanggalStr == "" {
		tanggal = time.Now()
	} else {
		tanggal, err = time.Parse("2006-01-02", tanggalStr)
		if err != nil {
			log.WithError(err).Warn("Format tanggal tidak valid")
			return utils.ResponseError(c, fiber.StatusBadRequest, "Format tanggal tidak valid, gunakan YYYY-MM-DD", nil)
		}
	}
	tanggal = time.Date(tanggal.Year(), tanggal.Month(), tanggal.Day(), 0, 0, 0, 0, time.UTC)

	var logHarian models.LogHarian
	err = s.DB.Preload("DetailLogs").
		Where(models.LogHarian{MahasantriID: mahasantriID, Tanggal: tanggal}).
		FirstOrCreate(&logHarian).Error

	if err != nil {
		log.WithError(err).Error("Gagal mengambil atau membuat log harian")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal memproses data log harian", err.Error())
	}

	detailDTOs := make([]dto.DetailLogResponse, len(logHarian.DetailLogs))
	for i, detail := range logHarian.DetailLogs {
		detailDTOs[i] = dto.DetailLogResponse{
			ID:                  detail.ID,
			WaktuMurojaah:       detail.WaktuMurojaah,
			TargetStartJuz:      detail.TargetStartJuz,
			TargetStartHalaman:  detail.TargetStartHalaman,
			TargetEndJuz:        detail.TargetEndJuz,
			TargetEndHalaman:    detail.TargetEndHalaman,
			TotalTargetHalaman:  detail.TotalTargetHalaman,
			SelesaiEndJuz:       detail.SelesaiEndJuz,
			SelesaiEndHalaman:   detail.SelesaiEndHalaman,
			TotalSelesaiHalaman: detail.TotalSelesaiHalaman,
			Status:              string(detail.Status),
			Catatan:             detail.Catatan,
			UpdatedAt:           detail.UpdatedAt,
		}
	}

	response := dto.LogHarianResponse{
		ID:                  logHarian.ID,
		Tanggal:             logHarian.Tanggal.Format("02-01-2006"),
		TotalTargetHalaman:  logHarian.TotalTargetHalaman,
		TotalSelesaiHalaman: logHarian.TotalSelesaiHalaman,
		DetailLogs:          detailDTOs,
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Log harian berhasil diproses", response)
}

func (s *logMurojaahService) AddDetailToLog(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)
	mahasantriID := claims.ID

	log := logrus.WithFields(logrus.Fields{"handler": "AddDetailToLog", "mahasantriID": mahasantriID})

	var req dto.AddDetailLogRequest
	if err := c.BodyParser(&req); err != nil {
		log.WithError(err).Error("Gagal parsing body request")
		return utils.ResponseError(c, fiber.StatusBadRequest, "Request body tidak valid", err.Error())
	}

	var newDetail models.DetailLog

	err := s.DB.Transaction(func(tx *gorm.DB) error {
		today := time.Now()
		today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)

		var logHarian models.LogHarian
		if err := tx.Where(models.LogHarian{MahasantriID: mahasantriID, Tanggal: today}).FirstOrCreate(&logHarian).Error; err != nil {
			return err
		}

		totalTarget, err := calculateTotalPages(req.TargetStartJuz, req.TargetStartHalaman, req.TargetEndJuz, req.TargetEndHalaman)
		if err != nil {
			return err
		}
		if totalTarget <= 0 {
			return errors.New("target murojaah harus lebih dari 0 halaman")
		}

		newDetail = models.DetailLog{
			LogHarianID:        logHarian.ID,
			WaktuMurojaah:      req.WaktuMurojaah,
			TargetStartJuz:     req.TargetStartJuz,
			TargetStartHalaman: req.TargetStartHalaman,
			TargetEndJuz:       req.TargetEndJuz,
			TargetEndHalaman:   req.TargetEndHalaman,
			TotalTargetHalaman: totalTarget,
			Status:             models.StatusSesiBelumSelesai,
			Catatan:            req.Catatan,
		}
		if err := tx.Create(&newDetail).Error; err != nil {
			return err
		}

		return s.recalculateTotals(tx, logHarian.ID)
	})

	if err != nil {
		log.WithError(err).Error("Gagal menambahkan detail log dalam transaksi")
		if err.Error() == "target murojaah harus lebih dari 0 halaman" || err.Error() == "target/progres akhir tidak boleh lebih kecil dari awal" {
			return utils.ResponseError(c, fiber.StatusBadRequest, err.Error(), nil)
		}
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal menyimpan sesi murojaah", err.Error())
	}

	response := dto.DetailLogResponse{
		ID:                  newDetail.ID,
		WaktuMurojaah:       newDetail.WaktuMurojaah,
		TargetStartJuz:      newDetail.TargetStartJuz,
		TargetStartHalaman:  newDetail.TargetStartHalaman,
		TargetEndJuz:        newDetail.TargetEndJuz,
		TargetEndHalaman:    newDetail.TargetEndHalaman,
		TotalTargetHalaman:  newDetail.TotalTargetHalaman,
		SelesaiEndJuz:       newDetail.SelesaiEndJuz,
		SelesaiEndHalaman:   newDetail.SelesaiEndHalaman,
		TotalSelesaiHalaman: newDetail.TotalSelesaiHalaman,
		Status:              string(newDetail.Status),
		Catatan:             newDetail.Catatan,
		UpdatedAt:           newDetail.UpdatedAt,
	}

	log.Info("Berhasil menambahkan detail sesi murojaah baru")
	return utils.SuccessResponse(c, fiber.StatusCreated, "Sesi murojaah berhasil ditambahkan", response)
}

func (s *logMurojaahService) UpdateDetailLog(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)
	mahasantriID := claims.ID
	detailID, err := c.ParamsInt("detailID")
	if err != nil {
		return utils.ResponseError(c, fiber.StatusBadRequest, "ID detail log tidak valid", nil)
	}

	log := logrus.WithFields(logrus.Fields{"handler": "UpdateDetailLog", "mahasantriID": mahasantriID, "detailID": detailID})

	var req dto.UpdateDetailLogRequest
	if err := c.BodyParser(&req); err != nil {
		log.WithError(err).Error("Gagal parsing body request")
		return utils.ResponseError(c, fiber.StatusBadRequest, "Request body tidak valid", err.Error())
	}

	var detailLog models.DetailLog

	err = s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Joins("JOIN log_harians ON log_harians.id = detail_logs.log_harian_id").
			Where("detail_logs.id = ? AND log_harians.mahasantri_id = ?", detailID, mahasantriID).
			First(&detailLog).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("detail log tidak ditemukan atau Anda tidak punya hak akses")
			}
			return err
		}

		totalSelesai, err := calculateTotalPages(detailLog.TargetStartJuz, detailLog.TargetStartHalaman, req.SelesaiEndJuz, req.SelesaiEndHalaman)
		if err != nil {
			return err
		}

		if totalSelesai > detailLog.TotalTargetHalaman {
			totalSelesai = detailLog.TotalTargetHalaman
		}

		var newStatus models.StatusDetailLog
		if totalSelesai >= detailLog.TotalTargetHalaman {
			newStatus = "Selesai"
			log.Info("Progres mencapai target. Status diatur ke 'Selesai'.")
		} else if totalSelesai > 0 {
			newStatus = "Berjalan"
			log.Info("Progres parsial terdeteksi. Status diatur ke 'Berjalan'.")
		} else {
			newStatus = "Belum Selesai"
			log.Info("Tidak ada progres. Status diatur ke 'Belum Selesai'.")
		}

		detailLog.SelesaiEndJuz = req.SelesaiEndJuz
		detailLog.SelesaiEndHalaman = req.SelesaiEndHalaman
		detailLog.TotalSelesaiHalaman = totalSelesai
		detailLog.Catatan = req.Catatan
		detailLog.Status = newStatus

		if err := tx.Save(&detailLog).Error; err != nil {
			return err
		}

		return s.recalculateTotals(tx, detailLog.LogHarianID)
	})

	if err != nil {
		log.WithError(err).Error("Gagal memperbarui detail log dalam transaksi")
		if err.Error() == "progres selesai tidak boleh melebihi total target halaman" || err.Error() == "target/progres akhir tidak boleh lebih kecil dari awal" {
			return utils.ResponseError(c, fiber.StatusBadRequest, err.Error(), nil)
		}
		if err.Error() == "detail log tidak ditemukan atau Anda tidak punya hak akses" {
			return utils.ResponseError(c, fiber.StatusNotFound, "Detail log tidak ditemukan atau Anda tidak punya hak akses", nil)
		}
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal memperbarui sesi murojaah", err.Error())
	}

	response := dto.DetailLogResponse{
		ID:                  detailLog.ID,
		WaktuMurojaah:       detailLog.WaktuMurojaah,
		TargetStartJuz:      detailLog.TargetStartJuz,
		TargetStartHalaman:  detailLog.TargetStartHalaman,
		TargetEndJuz:        detailLog.TargetEndJuz,
		TargetEndHalaman:    detailLog.TargetEndHalaman,
		TotalTargetHalaman:  detailLog.TotalTargetHalaman,
		SelesaiEndJuz:       detailLog.SelesaiEndJuz,
		SelesaiEndHalaman:   detailLog.SelesaiEndHalaman,
		TotalSelesaiHalaman: detailLog.TotalSelesaiHalaman,
		Status:              string(detailLog.Status),
		Catatan:             detailLog.Catatan,
		UpdatedAt:           detailLog.UpdatedAt,
	}

	log.Info("Berhasil memperbarui detail sesi murojaah")
	return utils.SuccessResponse(c, fiber.StatusOK, "Sesi murojaah berhasil diperbarui", response)
}

func (s *logMurojaahService) DeleteDetailLog(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)
	mahasantriID := claims.ID
	detailID, err := c.ParamsInt("detailID")
	if err != nil {
		return utils.ResponseError(c, fiber.StatusBadRequest, "ID detail log tidak valid", nil)
	}

	log := logrus.WithFields(logrus.Fields{"handler": "DeleteDetailLog", "mahasantriID": mahasantriID, "detailID": detailID})

	var detailLog models.DetailLog

	err = s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Joins("JOIN log_harians ON log_harians.id = detail_logs.log_harian_id").
			Where("detail_logs.id = ? AND log_harians.mahasantri_id = ?", detailID, mahasantriID).
			First(&detailLog).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("detail log tidak ditemukan atau Anda tidak punya hak akses")
			}
			return err
		}

		if err := tx.Delete(&detailLog).Error; err != nil {
			return err
		}

		return s.recalculateTotals(tx, detailLog.LogHarianID)
	})

	if err != nil {
		log.WithError(err).Error("Gagal menghapus detail log dalam transaksi")
		if err.Error() == "detail log tidak ditemukan atau Anda tidak punya hak akses" {
			return utils.ResponseError(c, fiber.StatusNotFound, "Detail log tidak ditemukan atau Anda tidak punya hak akses", nil)
		}
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal menghapus sesi murojaah", err.Error())
	}

	log.Info("Berhasil menghapus detail sesi murojaah")
	return utils.SuccessResponse(c, fiber.StatusOK, "Sesi murojaah berhasil dihapus", nil)
}

func (s *logMurojaahService) GetRecapMingguan(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)
	mahasantriID := claims.ID

	log := logrus.WithFields(logrus.Fields{"handler": "GetRecapMingguan", "mahasantriID": mahasantriID})
	log.Info("Menerima permintaan untuk rekap mingguan")

	type RecapResult struct {
		Tanggal             string `json:"tanggal"`
		TotalSelesaiHalaman int    `json:"total_selesai_halaman"`
	}

	var results []RecapResult

	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -6)

	err := s.DB.Model(&models.LogHarian{}).
		Select("to_char(tanggal, 'DD-MM-YYYY') as tanggal, total_selesai_halaman").
		Where("mahasantri_id = ? AND tanggal BETWEEN ? AND ?", mahasantriID, startDate, endDate).
		Order("tanggal ASC").
		Scan(&results).Error

	if err != nil {
		log.WithError(err).Error("Gagal mengambil data rekap mingguan")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal mengambil rekap", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Rekap mingguan berhasil diambil", results)
}

func (s *logMurojaahService) GetAllLogsForMentorDashboard(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)
	mentorID := claims.ID

	log := logrus.WithFields(logrus.Fields{
		"handler":  "GetAllLogsForMentorDashboard",
		"mentorID": mentorID,
	})
	log.Info("Menerima permintaan untuk dasbor log harian mentor")

	tanggalStr := c.Query("tanggal")
	var tanggal time.Time
	var err error
	if tanggalStr == "" {
		tanggal = time.Now()
	} else {
		tanggal, err = time.Parse("2006-01-02", tanggalStr)
		if err != nil {
			return utils.ResponseError(c, fiber.StatusBadRequest, "Format tanggal tidak valid, gunakan YYYY-MM-DD", nil)
		}
	}
	tanggal = time.Date(tanggal.Year(), tanggal.Month(), tanggal.Day(), 0, 0, 0, 0, time.UTC)
	log = log.WithField("tanggal", tanggal.Format("2006-01-02"))

	var mahasantriIDs []uint
	if err := s.DB.Model(&models.Mahasantri{}).Where("mentor_id = ?", mentorID).Pluck("id", &mahasantriIDs).Error; err != nil {
		log.WithError(err).Error("Gagal mengambil daftar ID mahasantri bimbingan")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal memproses data", err.Error())
	}

	if len(mahasantriIDs) == 0 {
		log.Warn("Mentor tidak memiliki mahasantri bimbingan")
		return utils.SuccessResponse(c, fiber.StatusOK, "Anda tidak memiliki mahasantri bimbingan", []interface{}{})
	}

	var logHarians []models.LogHarian
	if err := s.DB.Preload("Mahasantri").Preload("DetailLogs").
		Where("mahasantri_id IN ?", mahasantriIDs).
		Where("tanggal = ?", tanggal).
		Find(&logHarians).Error; err != nil {
		log.WithError(err).Error("Gagal mengambil log harian untuk dasbor mentor")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal mengambil data log", err.Error())
	}

	responseDTOs := make([]dto.LogHarianForMentorResponse, len(logHarians))
	for i, logHarian := range logHarians {
		detailDTOs := make([]dto.DetailLogResponse, len(logHarian.DetailLogs))
		for j, detail := range logHarian.DetailLogs {
			detailDTOs[j] = dto.DetailLogResponse{
				ID:                  detail.ID,
				WaktuMurojaah:       detail.WaktuMurojaah,
				TargetStartJuz:      detail.TargetStartJuz,
				TargetStartHalaman:  detail.TargetStartHalaman,
				TargetEndJuz:        detail.TargetEndJuz,
				TargetEndHalaman:    detail.TargetEndHalaman,
				TotalTargetHalaman:  detail.TotalTargetHalaman,
				SelesaiEndJuz:       detail.SelesaiEndJuz,
				SelesaiEndHalaman:   detail.SelesaiEndHalaman,
				TotalSelesaiHalaman: detail.TotalSelesaiHalaman,
				Status:              string(detail.Status),
				Catatan:             detail.Catatan,
				UpdatedAt:           detail.UpdatedAt,
			}
		}

		responseDTOs[i] = dto.LogHarianForMentorResponse{
			LogID:               logHarian.ID,
			Tanggal:             logHarian.Tanggal.Format("02-01-2006"),
			TotalTargetHalaman:  logHarian.TotalTargetHalaman,
			TotalSelesaiHalaman: logHarian.TotalSelesaiHalaman,
			Mahasantri: dto.MahasantriInfoForLog{
				ID:   logHarian.Mahasantri.ID,
				Nama: logHarian.Mahasantri.Nama,
			},
			DetailLogs: detailDTOs,
		}
	}

	log.WithField("count", len(responseDTOs)).Info("Berhasil mengambil data untuk dasbor mentor")
	return utils.SuccessResponse(c, fiber.StatusOK, "Data log harian untuk semua mahasantri berhasil diambil", responseDTOs)
}

func (s *logMurojaahService) GetStatistikMurojaah(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)
	mahasantriID := claims.ID

	log := logrus.WithFields(logrus.Fields{"handler": "GetStatistikMurojaah", "mahasantriID": mahasantriID})
	log.Info("Menerima permintaan untuk statistik murojaah")

	var stats struct {
		TotalSelesai int
		HariAktif    int
	}
	err := s.DB.Model(&models.LogHarian{}).
		Select("SUM(total_selesai_halaman) as total_selesai, COUNT(id) as hari_aktif").
		Where("mahasantri_id = ? AND total_selesai_halaman > 0", mahasantriID).
		Scan(&stats).Error
	if err != nil {
		log.WithError(err).Error("Gagal menghitung statistik dasar")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal memproses statistik", err.Error())
	}

	var rataRata float64
	if stats.HariAktif > 0 {
		rataRata = float64(stats.TotalSelesai) / float64(stats.HariAktif)
	}

	var hariProduktif dto.RecapHarianSimple
	err = s.DB.Model(&models.LogHarian{}).
		Select("to_char(tanggal, 'DD-MM-YYYY') as tanggal, total_selesai_halaman").
		Where("mahasantri_id = ?", mahasantriID).
		Order("total_selesai_halaman DESC").
		Limit(1).
		Scan(&hariProduktif).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.WithError(err).Error("Gagal mencari hari paling produktif")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal memproses statistik", err.Error())
	}

	var sesiProduktif struct {
		WaktuMurojaah string
	}
	err = s.DB.Model(&models.DetailLog{}).
		Select("waktu_murojaah").
		Joins("JOIN log_harians ON log_harians.id = detail_logs.log_harian_id").
		Where("log_harians.mahasantri_id = ? AND detail_logs.status = ?", mahasantriID, models.StatusSesiSelesai).
		Group("waktu_murojaah").
		Order("COUNT(detail_logs.id) DESC").
		Limit(1).
		Scan(&sesiProduktif).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.WithError(err).Error("Gagal mencari sesi paling produktif")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal memproses statistik", err.Error())
	}

	var hariProduktifPtr *dto.RecapHarianSimple
	if hariProduktif.Tanggal != "" {
		hariProduktifPtr = &hariProduktif
	}

	response := dto.StatistikMurojaahResponse{
		TotalSelesaiHalaman:    stats.TotalSelesai,
		TotalHariAktif:         stats.HariAktif,
		RataRataHalamanPerHari: rataRata,
		SesiPalingProduktif:    sesiProduktif.WaktuMurojaah,
		HariPalingProduktif:    hariProduktifPtr,
	}

	log.Info("Berhasil mengambil data statistik murojaah")
	return utils.SuccessResponse(c, fiber.StatusOK, "Statistik murojaah berhasil diambil", response)
}

// GetRekapBimbinganMingguan - Mengambil rekapitulasi progres mingguan semua mahasantri bimbingan.
// @Summary Rekapitulasi Mingguan Bimbingan
// @Description Endpoint untuk mengambil rekapitulasi progres muroja'ah (total halaman target vs selesai) selama 7 hari terakhir untuk semua mahasantri yang diampu oleh mentor yang sedang login. Hasil diurutkan berdasarkan halaman selesai terbanyak.
// @Tags Mentor
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response "Rekapitulasi mingguan berhasil diambil"
// @Failure 500 {object} utils.Response "Gagal mengambil data rekapitulasi"
// @Security BearerAuth
// @Router /api/v1/mentor/rekap-bimbingan/mingguan [get]
func (s *logMurojaahService) GetRekapBimbinganMingguan(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)
	mentorID := claims.ID

	log := logrus.WithFields(logrus.Fields{
		"handler":  "GetRekapBimbinganMingguan",
		"mentorID": mentorID,
	})
	log.Info("Menerima permintaan untuk rekap bimbingan mingguan")

	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -6)
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)

	var results []dto.RekapBimbinganResponse

	err := s.DB.Table("mahasantris as m").
		Select(`
			m.id as mahasantri_id,
			m.nama as nama_mahasantri,
			COALESCE(SUM(lh.total_target_halaman), 0) as total_target_halaman_mingguan,
			COALESCE(SUM(lh.total_selesai_halaman), 0) as total_selesai_halaman_mingguan
		`).
		Joins(
			"LEFT JOIN log_harians as lh ON m.id = lh.mahasantri_id AND lh.tanggal BETWEEN ? AND ?",
			startDate,
			endDate,
		).
		Where("m.mentor_id = ?", mentorID).
		Group("m.id, m.nama").
		Order("total_selesai_halaman_mingguan DESC").
		Scan(&results).Error

	if err != nil {
		log.WithError(err).Error("Gagal mengambil data rekapitulasi bimbingan")
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal mengambil data rekapitulasi", err.Error())
	}

	for i := range results {
		if results[i].TotalTargetHalamanMingguan > 0 {
			results[i].PersentasePencapaian = (float64(results[i].TotalSelesaiHalamanMingguan) / float64(results[i].TotalTargetHalamanMingguan)) * 100
		} else {
			results[i].PersentasePencapaian = 0
		}
	}

	log.WithField("count", len(results)).Info("Berhasil mengambil rekapitulasi bimbingan mingguan")
	return utils.SuccessResponse(c, fiber.StatusOK, "Rekapitulasi bimbingan mingguan berhasil diambil", results)
}

func (s *logMurojaahService) ApplyAIRekomendasi(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)
	mahasantriID := claims.ID

	log := logrus.WithFields(logrus.Fields{"handler": "ApplyAIRekomendasi", "mahasantriID": mahasantriID})

	var req dto.ApplyAIRekomendasiRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ResponseError(c, fiber.StatusBadRequest, "Request body tidak valid", err.Error())
	}

	var newDetail models.DetailLog
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		var rekomendasi models.JadwalRekomendasi
		if err := tx.Where("id = ? AND mahasantri_id = ?", req.RekomendasiID, mahasantriID).First(&rekomendasi).Error; err != nil {
			return errors.New("riwayat rekomendasi tidak ditemukan atau bukan milik anda")
		}

		today := time.Now()
		today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
		var logHarian models.LogHarian
		if err := tx.Where(models.LogHarian{MahasantriID: mahasantriID, Tanggal: today}).FirstOrCreate(&logHarian).Error; err != nil {
			return err
		}

		totalTarget, err := calculateTotalPages(req.TargetStartJuz, req.TargetStartHalaman, req.TargetEndJuz, req.TargetEndHalaman)
		if err != nil {
			return err
		}

		newDetail = models.DetailLog{
			LogHarianID:        logHarian.ID,
			WaktuMurojaah:      fmt.Sprintf("AI: %s", rekomendasi.RekomendasiJadwal),
			TargetStartJuz:     req.TargetStartJuz,
			TargetStartHalaman: req.TargetStartHalaman,
			TargetEndJuz:       req.TargetEndJuz,
			TargetEndHalaman:   req.TargetEndHalaman,
			TotalTargetHalaman: totalTarget,
			Status:             models.StatusSesiBelumSelesai,
			Catatan:            req.Catatan,
		}
		if err := tx.Create(&newDetail).Error; err != nil {
			return err
		}

		return s.recalculateTotals(tx, logHarian.ID)
	})

	if err != nil {
		log.WithError(err).Error("Gagal menerapkan rekomendasi AI dalam transaksi")
		if err.Error() == "riwayat rekomendasi tidak ditemukan atau bukan milik anda" {
			return utils.ResponseError(c, fiber.StatusNotFound, err.Error(), nil)
		}
		if err.Error() == "target/progres akhir tidak boleh lebih kecil dari awal" {
			return utils.ResponseError(c, fiber.StatusBadRequest, err.Error(), nil)
		}
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal menerapkan rekomendasi", err.Error())
	}

	response := dto.DetailLogResponse{
		ID:                  newDetail.ID,
		WaktuMurojaah:       newDetail.WaktuMurojaah,
		TargetStartJuz:      newDetail.TargetStartJuz,
		TargetStartHalaman:  newDetail.TargetStartHalaman,
		TargetEndJuz:        newDetail.TargetEndJuz,
		TargetEndHalaman:    newDetail.TargetEndHalaman,
		TotalTargetHalaman:  newDetail.TotalTargetHalaman,
		SelesaiEndJuz:       newDetail.SelesaiEndJuz,
		SelesaiEndHalaman:   newDetail.SelesaiEndHalaman,
		TotalSelesaiHalaman: newDetail.TotalSelesaiHalaman,
		Status:              string(newDetail.Status),
		Catatan:             newDetail.Catatan,
		UpdatedAt:           newDetail.UpdatedAt,
	}
	return utils.SuccessResponse(c, fiber.StatusCreated, "Rekomendasi berhasil diterapkan ke log harian", response)
}
