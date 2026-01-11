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
	"gorm.io/gorm"
)

type LogMurojaahService interface {
	GetOrCreateLogHarian(c *fiber.Ctx) error
	AddDetailToLog(c *fiber.Ctx) error
	AddProgressToDetail(c *fiber.Ctx) error
	DeleteDetailLog(c *fiber.Ctx) error
	DeleteProgressLog(c *fiber.Ctx) error
	GetRecapMingguan(c *fiber.Ctx) error
	GetAllLogsForMentorDashboard(c *fiber.Ctx) error
	GetStatistikMurojaah(c *fiber.Ctx) error
	GetRekapBimbinganMingguan(c *fiber.Ctx) error
	ApplyAIRekomendasi(c *fiber.Ctx) error
	GetWaktuMurojaahOptions(c *fiber.Ctx) error
}

type logMurojaahService struct {
	DB           *gorm.DB
	NotifService *NotificationService
}

func NewLogMurojaahService(db *gorm.DB, notifService *NotificationService) LogMurojaahService {
	return &logMurojaahService{
		DB:           db,
		NotifService: notifService,
	}
}

// Helper: Build Detail Log Response
func buildDetailLogResponse(detail models.DetailLog) dto.DetailLogResponse {
	waktuLabel := models.GetWaktuMurojaahLabel(detail.WaktuMurojaah)

	var percentage float64
	if detail.TotalTargetHalaman > 0 {
		percentage = (float64(detail.TotalSelesaiHalaman) / float64(detail.TotalTargetHalaman)) * 100
	}

	progressResponses := make([]dto.ProgressLogResponse, len(detail.ProgressLogs))
	for j, pl := range detail.ProgressLogs {
		progressResponses[j] = dto.ProgressLogResponse{
			ID:            pl.ID,
			JumlahHalaman: pl.JumlahHalaman,
			Catatan:       pl.Catatan,
			CreatedAt:     pl.CreatedAt,
		}
	}

	return dto.DetailLogResponse{
		ID:                  detail.ID,
		WaktuMurojaah:       detail.WaktuMurojaah,
		WaktuMurojaahLabel:  waktuLabel,
		TargetStartJuz:      detail.TargetStartJuz,
		TargetStartHalaman:  detail.TargetStartHalaman,
		TargetEndJuz:        detail.TargetEndJuz,
		TargetEndHalaman:    detail.TargetEndHalaman,
		TotalTargetHalaman:  detail.TotalTargetHalaman,
		TotalSelesaiHalaman: detail.TotalSelesaiHalaman,
		PersentaseSelesai:   percentage,
		Status:              string(detail.Status),
		Catatan:             detail.Catatan,
		CreatedAt:           detail.CreatedAt,
		UpdatedAt:           detail.UpdatedAt,
		ProgressLogs:        progressResponses,
	}
}

// calculateTotalPages menghitung total halaman dari range juz dan halaman
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

// recalculateDetailLogTotals menghitung ulang total selesai dari ProgressLogs
func (s *logMurojaahService) recalculateDetailLogTotals(tx *gorm.DB, detailLogID uint) (models.StatusDetailLog, error) {
	var totalSelesai int
	err := tx.Model(&models.ProgressLog{}).
		Select("COALESCE(SUM(jumlah_halaman), 0)").
		Where("detail_log_id = ?", detailLogID).
		Scan(&totalSelesai).Error
	if err != nil {
		return "", err
	}

	var detail models.DetailLog
	if err := tx.First(&detail, detailLogID).Error; err != nil {
		return "", err
	}

	var newStatus models.StatusDetailLog
	if totalSelesai >= detail.TotalTargetHalaman {
		newStatus = models.StatusSesiSelesai
	} else if totalSelesai > 0 {
		newStatus = models.StatusSesiBerjalan
	} else {
		newStatus = models.StatusSesiBelumSelesai
	}

	err = tx.Model(&models.DetailLog{}).Where("id = ?", detailLogID).Updates(map[string]interface{}{
		"total_selesai_halaman": totalSelesai,
		"status":                newStatus,
	}).Error

	return newStatus, err
}

// recalculateLogHarianTotals menghitung ulang total dari DetailLogs
func (s *logMurojaahService) recalculateLogHarianTotals(tx *gorm.DB, logHarianID uint) error {
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

// GetWaktuMurojaahOptions mengembalikan semua opsi waktu murojaah
func (s *logMurojaahService) GetWaktuMurojaahOptions(c *fiber.Ctx) error {
	options := make([]dto.WaktuMurojaahOptionResponse, len(models.WaktuMurojaahOptions))
	for i, opt := range models.WaktuMurojaahOptions {
		options[i] = dto.WaktuMurojaahOptionResponse{
			Kode:      opt.Kode,
			Label:     opt.Label,
			JamMulai:  opt.JamMulai,
			JamAkhir:  opt.JamAkhir,
			FullLabel: opt.FullLabel,
		}
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Opsi waktu murojaah berhasil diambil", options)
}

// GetOrCreateLogHarian
func (s *logMurojaahService) GetOrCreateLogHarian(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)
	userID := claims.ID

	// Check if date query param exists
	dateParam := c.Query("date")
	targetDate := time.Now()
	if dateParam != "" {
		parsedDate, err := time.Parse("2006-01-02", dateParam)
		if err == nil {
			targetDate = parsedDate
		}
	}

	dateString := targetDate.Format("2006-01-02")

	var log models.LogHarian
	err := s.DB.Preload("DetailLogs.ProgressLogs", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC")
	}).Preload("DetailLogs", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).Where("mahasantri_id = ? AND tanggal = ?", userID, dateString).First(&log).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		log = models.LogHarian{
			MahasantriID: userID,
			Tanggal:      targetDate,
		}
		if err := s.DB.Create(&log).Error; err != nil {
			return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal membuat log harian", err.Error())
		}
		s.DB.Preload("DetailLogs").First(&log, log.ID)
	} else if err != nil {
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal mengambil log harian", err.Error())
	}

	response := dto.LogHarianResponse{
		ID:                  log.ID,
		Tanggal:             log.Tanggal.Format("2006-01-02"),
		TotalTargetHalaman:  log.TotalTargetHalaman,
		TotalSelesaiHalaman: log.TotalSelesaiHalaman,
		DetailLogs:          make([]dto.DetailLogResponse, len(log.DetailLogs)),
	}

	var persentase float64
	if log.TotalTargetHalaman > 0 {
		persentase = (float64(log.TotalSelesaiHalaman) / float64(log.TotalTargetHalaman)) * 100
	}
	response.PersentaseSelesai = persentase

	for i, detail := range log.DetailLogs {
		response.DetailLogs[i] = buildDetailLogResponse(detail)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Log harian berhasil diambil", response)
}

// AddDetailToLog
func (s *logMurojaahService) AddDetailToLog(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)
	userID := claims.ID

	idParam := c.Params("id")
	logID, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.ResponseError(c, fiber.StatusBadRequest, "ID log tidak valid", err.Error())
	}

	var req dto.AddDetailLogRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ResponseError(c, fiber.StatusBadRequest, "Request body tidak valid", err.Error())
	}

	totalHalaman, err := calculateTotalPages(req.TargetStartJuz, req.TargetStartHalaman, req.TargetEndJuz, req.TargetEndHalaman)
	if err != nil {
		return utils.ResponseError(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	tx := s.DB.Begin()

	var log models.LogHarian
	if err := tx.Where("id = ? AND mahasantri_id = ?", logID, userID).First(&log).Error; err != nil {
		tx.Rollback()
		return utils.ResponseError(c, fiber.StatusNotFound, "Log harian tidak ditemukan", nil)
	}

	detail := models.DetailLog{
		LogHarianID:         uint(logID),
		WaktuMurojaah:       req.WaktuMurojaah,
		TargetStartJuz:      req.TargetStartJuz,
		TargetStartHalaman:  req.TargetStartHalaman,
		TargetEndJuz:        req.TargetEndJuz,
		TargetEndHalaman:    req.TargetEndHalaman,
		TotalTargetHalaman:  totalHalaman,
		TotalSelesaiHalaman: 0,
		Status:              models.StatusSesiBelumSelesai,
		Catatan:             req.Catatan,
	}

	if err := tx.Create(&detail).Error; err != nil {
		tx.Rollback()
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal menambahkan detail log", err.Error())
	}

	if err := s.recalculateLogHarianTotals(tx, uint(logID)); err != nil {
		tx.Rollback()
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal mengupdate total log", err.Error())
	}

	tx.Commit()

	// NOTIFICATION
	waktuLabel := models.GetWaktuMurojaahLabel(req.WaktuMurojaah)
	s.NotifService.CreateNotificationForMahasantri(
		userID,
		models.NotifTargetHarian,
		"Jadwal Murojaah Baru",
		fmt.Sprintf("Antum telah menambahkan jadwal murojaah: %s (%d halaman). Semangat!", waktuLabel, totalHalaman),
	)

	return utils.SuccessResponse(c, fiber.StatusCreated, "Detail log berhasil ditambahkan", nil)
}

// AddProgressToDetail
func (s *logMurojaahService) AddProgressToDetail(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)
	userID := claims.ID

	idParam := c.Params("id")
	detailID, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.ResponseError(c, fiber.StatusBadRequest, "ID detail tidak valid", err.Error())
	}

	var req dto.AddProgressRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ResponseError(c, fiber.StatusBadRequest, "Request body tidak valid", err.Error())
	}

	tx := s.DB.Begin()

	var detail models.DetailLog
	if err := tx.Joins("JOIN log_harians ON log_harians.id = detail_logs.log_harian_id").
		Where("detail_logs.id = ? AND log_harians.mahasantri_id = ?", detailID, userID).
		First(&detail).Error; err != nil {
		tx.Rollback()
		return utils.ResponseError(c, fiber.StatusNotFound, "Sesi murojaah tidak ditemukan", nil)
	}

	progress := models.ProgressLog{
		DetailLogID:   uint(detailID),
		JumlahHalaman: req.JumlahHalaman,
		Catatan:       req.Catatan,
	}

	if err := tx.Create(&progress).Error; err != nil {
		tx.Rollback()
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal menyimpan progress", err.Error())
	}

	newStatus, err := s.recalculateDetailLogTotals(tx, uint(detailID))
	if err != nil {
		tx.Rollback()
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal mengupdate status log", err.Error())
	}

	if err := s.recalculateLogHarianTotals(tx, detail.LogHarianID); err != nil {
		tx.Rollback()
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal mengupdate total harian", err.Error())
	}

	tx.Commit()

	// NOTIFICATION
	s.NotifService.CreateNotificationForMahasantri(
		userID,
		models.NotifHafalanTercatat,
		"Progress Tercatat",
		fmt.Sprintf("Alhamdulillah, bertambah %d halaman. Teruskan semangatnya!", req.JumlahHalaman),
	)

	if newStatus == models.StatusSesiSelesai {
		s.NotifService.CreateNotificationForMahasantri(
			userID,
			models.NotifMotivasi,
			"Sesi Murojaah Selesai!",
			"Masya Allah! Antum telah menyelesaikan target sesi ini. Barakallahu fiikum.",
		)
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Progress berhasil ditambahkan", dto.ProgressLogResponse{
		ID:            progress.ID,
		JumlahHalaman: progress.JumlahHalaman,
		Catatan:       progress.Catatan,
		CreatedAt:     progress.CreatedAt,
	})
}

// DeleteDetailLog
func (s *logMurojaahService) DeleteDetailLog(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)
	userID := claims.ID

	idParam := c.Params("detailID")
	detailID, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.ResponseError(c, fiber.StatusBadRequest, "ID detail log tidak valid", nil)
	}

	var detailLog models.DetailLog

	err = s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Joins("JOIN log_harians ON log_harians.id = detail_logs.log_harian_id").
			Where("detail_logs.id = ? AND log_harians.mahasantri_id = ?", detailID, userID).
			First(&detailLog).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("detail log tidak ditemukan atau Anda tidak punya hak akses")
			}
			return err
		}

		logHarianID := detailLog.LogHarianID

		if err := tx.Delete(&detailLog).Error; err != nil {
			return err
		}

		return s.recalculateLogHarianTotals(tx, logHarianID)
	})

	if err != nil {
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal menghapus sesi murojaah", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Sesi murojaah berhasil dihapus", nil)
}

// DeleteProgressLog
func (s *logMurojaahService) DeleteProgressLog(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)
	userID := claims.ID

	idParam := c.Params("progressID")
	progressID, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.ResponseError(c, fiber.StatusBadRequest, "ID progress log tidak valid", nil)
	}

	err = s.DB.Transaction(func(tx *gorm.DB) error {
		var progressLog models.ProgressLog
		err := tx.
			Joins("JOIN detail_logs ON detail_logs.id = progress_logs.detail_log_id").
			Joins("JOIN log_harians ON log_harians.id = detail_logs.log_harian_id").
			Where("progress_logs.id = ? AND log_harians.mahasantri_id = ?", progressID, userID).
			First(&progressLog).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("progress log tidak ditemukan atau Anda tidak punya hak akses")
			}
			return err
		}

		detailLogID := progressLog.DetailLogID

		var detailLog models.DetailLog
		if err := tx.First(&detailLog, detailLogID).Error; err != nil {
			return err
		}

		if err := tx.Delete(&progressLog).Error; err != nil {
			return err
		}

		if _, err := s.recalculateDetailLogTotals(tx, detailLogID); err != nil {
			return err
		}

		return s.recalculateLogHarianTotals(tx, detailLog.LogHarianID)
	})

	if err != nil {
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal menghapus progress", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Progress berhasil dihapus", nil)
}

// GetRecapMingguan
func (s *logMurojaahService) GetRecapMingguan(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)
	mahasantriID := claims.ID

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
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal mengambil rekap", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Rekap mingguan berhasil diambil", results)
}

// GetAllLogsForMentorDashboard
func (s *logMurojaahService) GetAllLogsForMentorDashboard(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)
	mentorID := claims.ID

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

	var mahasantriIDs []uint
	if err := s.DB.Model(&models.Mahasantri{}).Where("mentor_id = ?", mentorID).Pluck("id", &mahasantriIDs).Error; err != nil {
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal memproses data", err.Error())
	}

	if len(mahasantriIDs) == 0 {
		return utils.SuccessResponse(c, fiber.StatusOK, "Anda tidak memiliki mahasantri bimbingan", []interface{}{})
	}

	var logHarians []models.LogHarian
	if err := s.DB.Preload("Mahasantri").Preload("DetailLogs.ProgressLogs").
		Where("mahasantri_id IN ?", mahasantriIDs).
		Where("tanggal = ?", tanggal).
		Find(&logHarians).Error; err != nil {
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal mengambil data log", err.Error())
	}

	responseDTOs := make([]dto.LogHarianForMentorResponse, len(logHarians))
	for i, logHarian := range logHarians {
		detailDTOs := make([]dto.DetailLogResponse, len(logHarian.DetailLogs))
		for j, detail := range logHarian.DetailLogs {
			detailDTOs[j] = buildDetailLogResponse(detail)
		}

		var persentase float64
		if logHarian.TotalTargetHalaman > 0 {
			persentase = (float64(logHarian.TotalSelesaiHalaman) / float64(logHarian.TotalTargetHalaman)) * 100
		}

		responseDTOs[i] = dto.LogHarianForMentorResponse{
			LogID:               logHarian.ID,
			Tanggal:             logHarian.Tanggal.Format("02-01-2006"),
			TotalTargetHalaman:  logHarian.TotalTargetHalaman,
			TotalSelesaiHalaman: logHarian.TotalSelesaiHalaman,
			PersentaseSelesai:   persentase,
			Mahasantri: dto.MahasantriInfoForLog{
				ID:   logHarian.Mahasantri.ID,
				Nama: logHarian.Mahasantri.Nama,
			},
			DetailLogs: detailDTOs,
		}
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Data log harian untuk semua mahasantri berhasil diambil", responseDTOs)
}

// GetStatistikMurojaah
func (s *logMurojaahService) GetStatistikMurojaah(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)
	mahasantriID := claims.ID

	var stats struct {
		TotalSelesai int
		HariAktif    int
	}
	err := s.DB.Model(&models.LogHarian{}).
		Select("SUM(total_selesai_halaman) as total_selesai, COUNT(id) as hari_aktif").
		Where("mahasantri_id = ? AND total_selesai_halaman > 0", mahasantriID).
		Scan(&stats).Error
	if err != nil {
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
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal memproses statistik", err.Error())
	}

	var hariProduktifPtr *dto.RecapHarianSimple
	if hariProduktif.Tanggal != "" {
		hariProduktifPtr = &hariProduktif
	}

	sesiLabel := ""
	if sesiProduktif.WaktuMurojaah != "" {
		sesiLabel = models.GetWaktuMurojaahLabel(sesiProduktif.WaktuMurojaah)
	}

	response := dto.StatistikMurojaahResponse{
		TotalSelesaiHalaman:    stats.TotalSelesai,
		TotalHariAktif:         stats.HariAktif,
		RataRataHalamanPerHari: rataRata,
		SesiPalingProduktif:    sesiLabel,
		HariPalingProduktif:    hariProduktifPtr,
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Statistik murojaah berhasil diambil", response)
}

// GetRekapBimbinganMingguan
func (s *logMurojaahService) GetRekapBimbinganMingguan(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)
	mentorID := claims.ID

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
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal mengambil data rekapitulasi", err.Error())
	}

	for i := range results {
		if results[i].TotalTargetHalamanMingguan > 0 {
			results[i].PersentasePencapaian = (float64(results[i].TotalSelesaiHalamanMingguan) / float64(results[i].TotalTargetHalamanMingguan)) * 100
		} else {
			results[i].PersentasePencapaian = 0
		}
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Rekapitulasi bimbingan mingguan berhasil diambil", results)
}

// ApplyAIRekomendasi
func (s *logMurojaahService) ApplyAIRekomendasi(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)
	mahasantriID := claims.ID

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

		// VALIDASI: Cek apakah ada DetailLog yang belum selesai
		var unfinishedCount int64
		if err := tx.Model(&models.DetailLog{}).
			Where("log_harian_id = ? AND status != ?", logHarian.ID, models.StatusSesiSelesai).
			Count(&unfinishedCount).Error; err != nil {
			return err
		}

		if unfinishedCount > 0 {
			return errors.New("tidak bisa membuat sesi baru, masih ada sesi yang belum selesai")
		}

		totalTarget, err := calculateTotalPages(req.TargetStartJuz, req.TargetStartHalaman, req.TargetEndJuz, req.TargetEndHalaman)
		if err != nil {
			return err
		}

		newDetail = models.DetailLog{
			LogHarianID:        logHarian.ID,
			WaktuMurojaah:      fmt.Sprintf("AI_%s", rekomendasi.RekomendasiJadwal),
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

		return s.recalculateLogHarianTotals(tx, logHarian.ID)
	})

	if err != nil {
		errMsg := err.Error()
		if errMsg == "riwayat rekomendasi tidak ditemukan atau bukan milik anda" {
			return utils.ResponseError(c, fiber.StatusNotFound, errMsg, nil)
		}
		if errMsg == "target/progres akhir tidak boleh lebih kecil dari awal" ||
			errMsg == "tidak bisa membuat sesi baru, masih ada sesi yang belum selesai" {
			return utils.ResponseError(c, fiber.StatusBadRequest, errMsg, nil)
		}
		return utils.ResponseError(c, fiber.StatusInternalServerError, "Gagal menerapkan rekomendasi", err.Error())
	}

	response := buildDetailLogResponse(newDetail)

	// NOTIFICATION for AI Applied
	s.NotifService.CreateNotificationForMahasantri(
		mahasantriID,
		models.NotifTargetHarian,
		"Rekomendasi AI Diterapkan",
		"Jadwal dari rekomendasi AI berhasil ditambahkan. Semangat!",
	)

	return utils.SuccessResponse(c, fiber.StatusCreated, "Rekomendasi berhasil diterapkan ke log harian", response)
}
