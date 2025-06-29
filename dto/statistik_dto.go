package dto

type StatistikMurojaahResponse struct {
	TotalSelesaiHalaman    int                `json:"total_selesai_halaman"`
	TotalHariAktif         int                `json:"total_hari_aktif"`
	RataRataHalamanPerHari float64            `json:"rata_rata_halaman_per_hari"`
	SesiPalingProduktif    string             `json:"sesi_paling_produktif"`
	HariPalingProduktif    *RecapHarianSimple `json:"hari_paling_produktif"`
}

type RecapHarianSimple struct {
	Tanggal             string `json:"tanggal"`
	TotalSelesaiHalaman int    `json:"total_selesai_halaman"`
}
