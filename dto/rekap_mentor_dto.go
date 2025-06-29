package dto

type RekapBimbinganResponse struct {
	MahasantriID                uint    `json:"mahasantri_id"`
	NamaMahasantri              string  `json:"nama_mahasantri"`
	TotalTargetHalamanMingguan  int     `json:"total_target_halaman_mingguan"`
	TotalSelesaiHalamanMingguan int     `json:"total_selesai_halaman_mingguan"`
	PersentasePencapaian        float64 `json:"persentase_pencapaian"`
}
