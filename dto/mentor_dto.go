package dto

type MentorResponse struct {
	ID              uint                    `json:"id"`
	Nama            string                  `json:"nama"`
	Email           string                  `json:"email"`
	Gender          string                  `json:"gender"`
	MahasantriCount int                     `json:"mahasantri_count"`
	Mahasantri      []MahasantriResponse    `json:"mahasantri,omitempty"`
	JadwalPersonal  *JadwalPersonalResponse `json:"jadwal_personal,omitempty"`
}

type UpdateMentorRequest struct {
	Nama   *string `json:"nama,omitempty"`
	Email  *string `json:"email,omitempty"`
	Gender *string `json:"gender,omitempty"`
}
