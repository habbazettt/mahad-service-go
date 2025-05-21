package test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/habbazettt/mahad-service-go/models"
	"github.com/habbazettt/mahad-service-go/utils"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var testResults []struct {
	Name   string
	Passed bool
}

func recordTestResult(t *testing.T, name string, passed *bool) {
	t.Cleanup(func() {
		testResults = append(testResults, struct {
			Name   string
			Passed bool
		}{Name: name, Passed: *passed})
	})
}

func TestMain(m *testing.M) {
	code := m.Run()

	println("\n============= TEST SUMMARY =============")
	for _, result := range testResults {
		status := "❌ FAIL"
		if result.Passed {
			status = "✅ PASS"
		}
		println(status, "-", result.Name)
	}
	println("========================================")

	os.Exit(code)
}

func createTestMentor(db *gorm.DB, email, password string) models.Mentor {
	hashedPass, _ := utils.HashPassword(password)
	mentor := models.Mentor{
		Nama:     "Test Mentor",
		Email:    email,
		Password: hashedPass,
		Gender:   "L",
	}
	db.Create(&mentor)
	return mentor
}

func createTestMahasantri(db *gorm.DB, nim, password string, mentorID uint) models.Mahasantri {
	hashedPass, _ := utils.HashPassword(password)
	santri := models.Mahasantri{
		Nama:     "Test Santri",
		NIM:      nim,
		Password: hashedPass,
		Jurusan:  "TI",
		Gender:   "L",
		MentorID: mentorID,
	}
	db.Create(&santri)
	return santri
}

func sendJSONRequest(app *fiber.App, method, path, payload string) (*http.Response, []byte, error) {
	req := httptest.NewRequest(method, path, strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp, body, nil
}

func TestLoginMentor_Success(t *testing.T) {
	app, db := SetupTestApp()
	createTestMentor(db, "satria@gmail.com", "satria1234")

	name := "TestLoginMentor_Success"
	passed := true
	recordTestResult(t, name, &passed)

	resp, body, err := sendJSONRequest(app, http.MethodPost, "/api/v1/auth/login/mentor", `{"email":"satria@gmail.com","password":"satria1234"}`)
	if !assert.NoError(t, err) || !assert.Equal(t, http.StatusOK, resp.StatusCode) {
		passed = false
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); !assert.NoError(t, err) {
		passed = false
		return
	}

	if !assert.True(t, result["status"].(bool)) ||
		!assert.Equal(t, "Login successful", result["message"]) {
		passed = false
		return
	}

	data := result["data"].(map[string]interface{})
	user := data["user"].(map[string]interface{})

	if !assert.NotEmpty(t, data["token"]) ||
		!assert.Equal(t, "satria@gmail.com", strings.ToLower(user["email"].(string))) ||
		!assert.Equal(t, "mentor", user["user_type"]) ||
		!assert.NotNil(t, user["mahasantri_count"]) {
		passed = false
	}
}

func TestLoginMentor_InvalidPassword(t *testing.T) {
	app, db := SetupTestApp()
	createTestMentor(db, "mentor2@example.com", "rightpass")

	name := "TestLoginMentor_InvalidPassword"
	passed := true
	recordTestResult(t, name, &passed)

	resp, body, err := sendJSONRequest(app, http.MethodPost, "/api/v1/auth/login/mentor", `{"email":"mentor2@example.com","password":"wrongpass"}`)
	if !assert.NoError(t, err) || !assert.Equal(t, http.StatusUnauthorized, resp.StatusCode) {
		passed = false
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		passed = false
		return
	}

	if !assert.False(t, result["status"].(bool)) ||
		!assert.Equal(t, "Invalid email or password", result["message"]) {
		passed = false
	}
}

func TestLoginMentor_NotRegistered(t *testing.T) {
	app, _ := SetupTestApp()

	name := "TestLoginMentor_NotRegistered"
	passed := true
	recordTestResult(t, name, &passed)

	resp, body, err := sendJSONRequest(app, http.MethodPost, "/api/v1/auth/login/mentor", `{"email":"notfound@example.com","password":"somepass"}`)
	if !assert.NoError(t, err) || !assert.Equal(t, http.StatusUnauthorized, resp.StatusCode) {
		passed = false
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		passed = false
		return
	}

	if !assert.False(t, result["status"].(bool)) ||
		!assert.Equal(t, "Invalid email or password", result["message"]) {
		passed = false
	}
}

func TestLoginMahasantri_Success(t *testing.T) {
	app, db := SetupTestApp()
	mentor := createTestMentor(db, "mentor@dummy.com", "dummy123")
	createTestMahasantri(db, "123456", "mahasantripass", mentor.ID)

	name := "TestLoginMahasantri_Success"
	passed := true
	recordTestResult(t, name, &passed)

	resp, body, err := sendJSONRequest(app, http.MethodPost, "/api/v1/auth/login/mahasantri", `{"nim":"123456","password":"mahasantripass"}`)
	if !assert.NoError(t, err) || !assert.Equal(t, http.StatusOK, resp.StatusCode) {
		passed = false
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); !assert.NoError(t, err) {
		passed = false
		return
	}

	if !assert.True(t, result["status"].(bool)) ||
		!assert.Equal(t, "Login successful", result["message"]) {
		passed = false
		return
	}

	data := result["data"].(map[string]interface{})
	user := data["user"].(map[string]interface{})

	if !assert.NotEmpty(t, data["token"]) ||
		!assert.Equal(t, "123456", user["nim"]) ||
		!assert.Equal(t, "mahasantri", user["user_type"]) ||
		!assert.NotNil(t, user["mentor_id"]) {
		passed = false
	}
}

func TestLoginMahasantri_InvalidPassword(t *testing.T) {
	app, db := SetupTestApp()
	mentor := createTestMentor(db, "x@x.com", "rightpass")
	createTestMahasantri(db, "654321", "rightpass", mentor.ID)

	name := "TestLoginMahasantri_InvalidPassword"
	passed := true
	recordTestResult(t, name, &passed)

	resp, body, err := sendJSONRequest(app, http.MethodPost, "/api/v1/auth/login/mahasantri", `{"nim":"654321","password":"wrongpass"}`)
	if !assert.NoError(t, err) || !assert.Equal(t, http.StatusUnauthorized, resp.StatusCode) {
		passed = false
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		passed = false
		return
	}

	if !assert.False(t, result["status"].(bool)) ||
		!assert.Equal(t, "Invalid NIM or password", result["message"]) {
		passed = false
	}
}

func TestLoginMahasantri_NotRegistered(t *testing.T) {
	app, _ := SetupTestApp()

	name := "TestLoginMahasantri_NotRegistered"
	passed := true
	recordTestResult(t, name, &passed)

	resp, body, err := sendJSONRequest(app, http.MethodPost, "/api/v1/auth/login/mahasantri", `{"nim":"999999","password":"somepass"}`)
	if !assert.NoError(t, err) || !assert.Equal(t, http.StatusUnauthorized, resp.StatusCode) {
		passed = false
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		passed = false
		return
	}

	if !assert.False(t, result["status"].(bool)) ||
		!assert.Equal(t, "Invalid NIM or password", result["message"]) {
		passed = false
	}
}
