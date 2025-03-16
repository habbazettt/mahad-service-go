package dto

type CreateHafalanRequest struct {
	MahasantriID uint    `json:"mahasantri_id" validate:"required"`
	Juz          int     `json:"juz" validate:"required,min=1,max=30"`
	Halaman      string  `json:"halaman" validate:"required"`
	TotalSetoran float32 `json:"total_setoran" validate:"required,min=0"`
	Kategori     string  `json:"kategori" validate:"required,oneof=ziyadah murojaah"`
	Waktu        string  `json:"waktu" validate:"required,oneof=subuh isya"`
	Catatan      string  `json:"catatan,omitempty"`
}

type UpdateHafalanRequest struct {
	Juz          *int     `json:"juz,omitempty"`
	Halaman      *string  `json:"halaman,omitempty"`
	TotalSetoran *float32 `json:"total_setoran,omitempty"`
	Kategori     *string  `json:"kategori,omitempty"`
	Waktu        *string  `json:"waktu,omitempty"`
	Catatan      *string  `json:"catatan,omitempty"`
}

type HafalanResponse struct {
	ID           uint    `json:"id"`
	MahasantriID uint    `json:"mahasantri_id"`
	Juz          int     `json:"juz"`
	Halaman      string  `json:"halaman"`
	TotalSetoran float32 `json:"total_setoran"`
	Kategori     string  `json:"kategori"`
	Waktu        string  `json:"waktu"`
	Catatan      string  `json:"catatan,omitempty"`
}
