package dto

type MahasantriResponse struct {
	ID                   uint                    `json:"id"`
	Nama                 string                  `json:"nama"`
	NIM                  string                  `json:"nim"`
	Jurusan              string                  `json:"jurusan"`
	Gender               string                  `json:"gender"`
	MentorID             uint                    `json:"mentor_id"`
	IsDataMurojaahFilled bool                    `json:"is_data_murojaah_filled"`
	JadwalPersonal       *JadwalPersonalResponse `json:"jadwal_personal,omitempty"`
}

type UpdateMahasantriRequest struct {
	Nama     *string `json:"nama,omitempty"`
	NIM      *string `json:"nim,omitempty"`
	Jurusan  *string `json:"jurusan,omitempty"`
	Gender   *string `json:"gender,omitempty"`
	MentorID *uint   `json:"mentor_id,omitempty"`
}
