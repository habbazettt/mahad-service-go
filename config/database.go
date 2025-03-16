package config

import (
	"os"

	"github.com/habbazettt/mahad-service-go/models"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() *gorm.DB {
	err := godotenv.Load()
	if err != nil {
		logrus.Warn("⚠️  .env file tidak ditemukan, menggunakan sistem environment")
	}

	InitLogger()

	dsn := os.Getenv("LOCAL_DB_URL")
	if os.Getenv("ENV") == "production" {
		dsn = os.Getenv("PROD_DB_URL")
		logrus.Info("🔄 Menggunakan database NeonDB (Production)")
	} else {
		logrus.Info("🔄 Menggunakan database PostgreSQL (Local)")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logrus.WithError(err).Fatal("❌ Gagal terhubung ke database!")
	}

	logrus.Info("✅ Berhasil terhubung ke database!")
	DB = db

	return DB
}

func MigrateDB() {
	if DB == nil {
		logrus.Fatal("❌ Database belum terhubung! Jalankan ConnectDB() terlebih dahulu.")
	}

	err := DB.AutoMigrate(&models.Mentor{}, &models.Mahasantri{}, &models.Hafalan{}, &models.Absensi{})
	if err != nil {
		logrus.WithError(err).Fatal("❌ Gagal melakukan migrasi database!")
	}

	logrus.Info("✅ Database berhasil dimigrasi!")
}
