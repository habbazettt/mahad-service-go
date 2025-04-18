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

// SuccessExample contoh response sukses untuk dokumentasi Swagger
type SuccessExample struct {
	Status  bool        `json:"status" example:"true"`
	Message string      `json:"message" example:"Mahasantri registered successfully"`
	Data    interface{} `json:"data"`
}

// ErrorExample contoh response error untuk dokumentasi Swagger
type ErrorExample struct {
	Status  bool        `json:"status" example:"false"`
	Message string      `json:"message" example:"Invalid request body"`
	Error   interface{} `json:"error"`
}

// SuccessMentorExample contoh response sukses untuk register mentor
type SuccessMentorExample struct {
	Status  bool        `json:"status" example:"true"`
	Message string      `json:"message" example:"Mentor registered successfully"`
	Data    interface{} `json:"data"`
}

// ErrorMentorExample contoh response error untuk register mentor
type ErrorMentorExample struct {
	Status  bool        `json:"status" example:"false"`
	Message string      `json:"message" example:"Invalid request body"`
	Error   interface{} `json:"error"`
}

// SuccessLoginMentorExample contoh response sukses untuk login mentor
type SuccessLoginMentorExample struct {
	Status  bool        `json:"status" example:"true"`
	Message string      `json:"message" example:"Login successful"`
	Data    interface{} `json:"data"`
}

// ErrorLoginMentorExample contoh response error untuk login mentor
type ErrorLoginMentorExample struct {
	Status  bool        `json:"status" example:"false"`
	Message string      `json:"message" example:"Invalid email or password"`
	Error   interface{} `json:"error"`
}

// SuccessLoginMahasantriExample contoh response sukses untuk login mahasantri
type SuccessLoginMahasantriExample struct {
	Status  bool        `json:"status" example:"true"`
	Message string      `json:"message" example:"Login successful"`
	Data    interface{} `json:"data"`
}

// ErrorLoginMahasantriExample contoh response error untuk login mahasantri
type ErrorLoginMahasantriExample struct {
	Status  bool        `json:"status" example:"false"`
	Message string      `json:"message" example:"Invalid NIM or password"`
	Error   interface{} `json:"error"`
}

// SuccessGetCurrentUserExample contoh response sukses untuk mendapatkan data user saat ini
type SuccessGetCurrentUserExample struct {
	Status  bool        `json:"status" example:"true"`
	Message string      `json:"message" example:"User data retrieved"`
	Data    interface{} `json:"data"`
}

// ErrorGetCurrentUserExample contoh response error untuk mendapatkan data user saat ini
type ErrorGetCurrentUserExample struct {
	Status  bool        `json:"status" example:"false"`
	Message string      `json:"message" example:"Unauthorized"`
	Error   interface{} `json:"error"`
}
