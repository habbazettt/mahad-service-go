package dto

type RegisterMahasantriRequest struct {
	Nama     string `json:"nama" validate:"required"`
	NIM      string `json:"nim" validate:"required"`
	Jurusan  string `json:"jurusan" validate:"required"`
	Gender   string `json:"gender" validate:"required,oneof=L P"`
	Password string `json:"password" validate:"required,min=6"`
	MentorID uint   `json:"mentor_id" validate:"required"`
}

type RegisterMentorRequest struct {
	Nama     string `json:"nama" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Gender   string `json:"gender" validate:"required,oneof=L P"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginMahasantriRequest struct {
	NIM      string `json:"nim" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginMentorRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type ForgotPasswordRequest struct {
	NIM         string `json:"nim,omitempty"`
	Email       string `json:"email,omitempty"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
}

type UserMahasantriResponse struct {
	ID                   uint   `json:"id"`
	Nama                 string `json:"nama"`
	NIM                  string `json:"nim"`
	Jurusan              string `json:"jurusan"`
	Gender               string `json:"gender"`
	MentorID             uint   `json:"mentor_id"`
	UserType             string `json:"user_type"`
	IsDataMurojaahFilled bool   `json:"is_data_murojaah_filled"`
}

type UserMentorResponse struct {
	ID                   uint   `json:"id"`
	Nama                 string `json:"nama"`
	Email                string `json:"email"`
	Gender               string `json:"gender"`
	MahasantriCount      int    `json:"mahasantri_count"`
	UserType             string `json:"user_type"`
	IsDataMurojaahFilled bool   `json:"is_data_murojaah_filled"`
}

type UserMentorWithMahasantriResponse struct {
	ID                   uint                     `json:"id"`
	Nama                 string                   `json:"nama"`
	Email                string                   `json:"email"`
	Gender               string                   `json:"gender"`
	MahasantriCount      int                      `json:"mahasantri_count"`
	Mahasantri           []UserMahasantriResponse `json:"mahasantri"`
	UserType             string                   `json:"user_type"`
	IsDataMurojaahFilled bool                     `json:"is_data_murojaah_filled"`
}
