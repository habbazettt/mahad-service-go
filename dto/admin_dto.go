package dto

// LoginAdminRequest adalah request untuk login admin
type LoginAdminRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// RegisterAdminRequest adalah request untuk registrasi admin baru (hanya admin yang bisa akses)
type RegisterAdminRequest struct {
	Nama     string `json:"nama" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// UpdateAdminRequest adalah request untuk update data admin
type UpdateAdminRequest struct {
	Nama  string `json:"nama,omitempty"`
	Email string `json:"email,omitempty"`
}

// UserAdminResponse adalah response data admin
type UserAdminResponse struct {
	ID       uint   `json:"id"`
	Nama     string `json:"nama"`
	Email    string `json:"email"`
	UserType string `json:"user_type"`
}

// AdminDashboardResponse adalah response untuk dashboard admin
type AdminDashboardResponse struct {
	TotalMentor     int64 `json:"total_mentor"`
	TotalMahasantri int64 `json:"total_mahasantri"`
	TotalHafalan    int64 `json:"total_hafalan"`
	TotalAbsensi    int64 `json:"total_absensi"`
}

// AdminMentorResponse adalah response mentor untuk admin (dengan statistik)
type AdminMentorResponse struct {
	ID              uint   `json:"id"`
	Nama            string `json:"nama"`
	Email           string `json:"email"`
	Gender          string `json:"gender"`
	MahasantriCount int64  `json:"mahasantri_count"`
}

// AdminMahasantriResponse adalah response mahasantri untuk admin
type AdminMahasantriResponse struct {
	ID         uint   `json:"id"`
	Nama       string `json:"nama"`
	NIM        string `json:"nim"`
	Jurusan    string `json:"jurusan"`
	Gender     string `json:"gender"`
	MentorID   uint   `json:"mentor_id"`
	MentorNama string `json:"mentor_nama"`
}

// AdminCreateMentorRequest adalah request untuk admin membuat mentor baru
type AdminCreateMentorRequest struct {
	Nama     string `json:"nama" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Gender   string `json:"gender" validate:"required,oneof=L P"`
	Password string `json:"password" validate:"required,min=6"`
}

// AdminUpdateMentorRequest adalah request untuk admin update mentor
type AdminUpdateMentorRequest struct {
	Nama   string `json:"nama,omitempty"`
	Email  string `json:"email,omitempty"`
	Gender string `json:"gender,omitempty"`
}

// AdminCreateMahasantriRequest adalah request untuk admin membuat mahasantri baru
type AdminCreateMahasantriRequest struct {
	Nama     string `json:"nama" validate:"required"`
	NIM      string `json:"nim" validate:"required"`
	Jurusan  string `json:"jurusan" validate:"required"`
	Gender   string `json:"gender" validate:"required,oneof=L P"`
	Password string `json:"password" validate:"required,min=6"`
	MentorID uint   `json:"mentor_id" validate:"required"`
}

// AdminUpdateMahasantriRequest adalah request untuk admin update mahasantri
type AdminUpdateMahasantriRequest struct {
	Nama     string `json:"nama,omitempty"`
	NIM      string `json:"nim,omitempty"`
	Jurusan  string `json:"jurusan,omitempty"`
	Gender   string `json:"gender,omitempty"`
	MentorID uint   `json:"mentor_id,omitempty"`
}
