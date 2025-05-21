package utils

// SuccessResponse contoh struktur respons sukses
type SuccessResponseSwagger struct {
	Status  string      `json:"status" example:"success"`
	Message string      `json:"message" example:"Request successful"`
	Data    interface{} `json:"data"`
}

// ErrorResponse contoh struktur respons gagal
type ErrorResponseSwagger struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Bad Request"`
	Error   string `json:"error" example:"Invalid Mahasantri ID"`
}
