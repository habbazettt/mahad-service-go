package dto

type CreateTargetSemesterRequest struct {
	MahasantriID uint   `json:"mahasantri_id" validate:"required"`
	Target       int    `json:"target" validate:"required,gt=0"`
	Semester     string `json:"semester" validate:"required,oneof=Ganjil Genap"`
	TahunAjaran  string `json:"tahun_ajaran" validate:"required,regexp=^[0-9]{4}/[0-9]{4}$"`
	Keterangan   string `json:"keterangan,omitempty"`
}

type UpdateTargetSemesterRequest struct {
	Target      *int    `json:"target,omitempty"`
	Semester    *string `json:"semester,omitempty"`
	TahunAjaran *string `json:"tahun_ajaran,omitempty"`
	Keterangan  *string `json:"keterangan,omitempty"`
}

type TargetSemesterResponse struct {
	ID           uint   `json:"id"`
	MahasantriID uint   `json:"mahasantri_id"`
	Target       int    `json:"target"`
	Semester     string `json:"semester"`
	TahunAjaran  string `json:"tahun_ajaran"`
	Keterangan   string `json:"keterangan,omitempty"`
}
