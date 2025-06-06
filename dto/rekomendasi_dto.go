package dto

type RecommendationRequest struct {
	Kesibukan       string `json:"kesibukan"`
	KategoriHafalan string `json:"kategori_hafalan"`
}

type RecommendationResponse struct {
	ID                        uint     `json:"id"`
	State                     string   `json:"state"`
	RekomendasiJadwal         string   `json:"rekomendasi_jadwal"`
	TipeRekomendasi           string   `json:"tipe_rekomendasi"`
	EstimasiQValue            *float64 `json:"estimasi_q_value,omitempty"`
	PersentaseEfektifHistoris *float64 `json:"persentase_efektif_historis,omitempty"`
}
