package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/habbazettt/mahad-service-go/config"
	"github.com/habbazettt/mahad-service-go/models"
	"github.com/habbazettt/mahad-service-go/utils"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type MahasantriSeedData struct {
	Nama     string `json:"nama"`
	NIM      string `json:"nim"`
	Jurusan  string `json:"jurusan"`
	Gender   string `json:"gender"`
	Password string `json:"password"`
	MentorID uint   `json:"mentor_id"`
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error memuat file .env: %v. Pastikan Anda menjalankan perintah dari folder root 'absensi-service'.", err)
	}

	db := config.ConnectDB()
	fmt.Println("✅ Berhasil terhubung ke database di VM GCP untuk seeding...")

	// Seed Admin pertama
	err = seedDefaultAdmin(db)
	if err != nil {
		log.Fatalf("❌ Gagal melakukan seeding data Admin: %v", err)
	}

	err = seedMahasantriFromJSON(db, "./cmd/seed/mahasantri_data.json")
	if err != nil {
		log.Fatalf("❌ Gagal melakukan seeding data Mahasantri dari JSON: %v", err)
	}

	fmt.Println("\n✅ Seeding data berhasil diselesaikan!")
}

// seedDefaultAdmin membuat admin default jika belum ada
func seedDefaultAdmin(db *gorm.DB) error {
	fmt.Println("\n--- Memulai Seeding Data Admin Default ---")

	// Cek apakah admin sudah ada
	var existingAdmin models.Admin
	result := db.Where("email = ?", "admin@mtadigital.com").First(&existingAdmin)

	if result.Error == gorm.ErrRecordNotFound {
		// Hash password
		hashedPassword, err := utils.HashPassword("admin123")
		if err != nil {
			return fmt.Errorf("gagal hash password admin: %w", err)
		}

		// Buat admin default
		admin := models.Admin{
			Nama:     "Administrator",
			Email:    "admin@mtadigital.com",
			Password: hashedPassword,
		}

		if err := db.Create(&admin).Error; err != nil {
			return fmt.Errorf("gagal membuat admin: %w", err)
		}

		fmt.Printf("👍 Berhasil membuat admin default: %s (%s)\n", admin.Nama, admin.Email)
		fmt.Println("   ⚠️  Password default: admin123 - Segera ganti password ini!")
	} else if result.Error != nil {
		return result.Error
	} else {
		fmt.Printf("ℹ️  Admin dengan email %s sudah ada, dilewati.\n", existingAdmin.Email)
	}

	return nil
}

func seedMahasantriFromJSON(db *gorm.DB, jsonPath string) error {
	fmt.Println("\n--- Memulai Seeding Data Mahasantri dari File JSON ---")

	// 1. Baca file JSON
	byteValue, err := ioutil.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("gagal membaca file JSON: %w", err)
	}

	// 2. Unmarshal (parse) data JSON ke dalam slice struct MahasantriSeedData
	var mahasantriToSeed []MahasantriSeedData
	if err := json.Unmarshal(byteValue, &mahasantriToSeed); err != nil {
		return fmt.Errorf("gagal unmarshal data JSON: %w", err)
	}

	fmt.Printf("ℹ️ Ditemukan %d data mahasantri di file JSON untuk di-seed.\n", len(mahasantriToSeed))

	// 3. Looping dan masukkan data
	for _, santriData := range mahasantriToSeed {
		// Cek dulu apakah mentor ID ada di database
		var mentor models.Mentor
		if err := db.First(&mentor, santriData.MentorID).Error; err != nil {
			fmt.Printf("⚠️  Peringatan: Melewati mahasantri '%s' karena MentorID %d tidak ditemukan. Silakan perbaiki data di JSON.\n", santriData.Nama, santriData.MentorID)
			continue // Lanjut ke data berikutnya
		}

		// Cek apakah NIM sudah ada
		var existingSantri models.Mahasantri
		result := db.Where("nim = ?", santriData.NIM).First(&existingSantri)

		if result.Error == gorm.ErrRecordNotFound {
			// Hash password dari JSON
			hashedPassword, err := utils.HashPassword(santriData.Password)
			if err != nil {
				fmt.Printf("❌ Gagal hash password untuk NIM %s: %v\n", santriData.NIM, err)
				continue
			}

			// Buat record baru
			newSantri := models.Mahasantri{
				Nama:     santriData.Nama,
				NIM:      santriData.NIM,
				Jurusan:  santriData.Jurusan,
				Gender:   santriData.Gender,
				Password: hashedPassword,
				MentorID: santriData.MentorID,
			}

			if err := db.Create(&newSantri).Error; err != nil {
				fmt.Printf("❌ Gagal membuat mahasantri %s: %v\n", santriData.NIM, err)
			} else {
				fmt.Printf("👍 Berhasil seeding mahasantri: %s - %s (Mentor: %s)\n", newSantri.NIM, newSantri.Nama, mentor.Nama)
			}
		} else if result.Error != nil {
			// Handle error lain
			return result.Error
		} else {
			// Jika sudah ada
			fmt.Printf("ℹ️  Mahasantri dengan NIM %s sudah ada, dilewati.\n", santriData.NIM)
		}
	}
	return nil
}
