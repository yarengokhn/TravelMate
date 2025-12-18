package database

import (
	"log"
	"travel-platform/travel-platform/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB //Pointer = Bellekteki adresi tutar (direkt nesne değil)
//Global olduğu için her yerden erişilebilir

// Initialize - Veritabanı bağlantısını kurar ve tabloları oluşturur
func ConnectDatabase() error {

	var error error

	// SQLite veritabanına bağlan
	// ":memory:" yerine "travel-platform.db" kullanıyoruz (kalıcı depolama)
	DB, error = gorm.Open(sqlite.Open("travel-platform.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if error != nil { //nil Go'da "hiçbir şey yok" demektir. Diğer dillerdeki null, None, nullptr gibi.
		log.Fatal("Failed to connect to database:", error)
	}
	log.Println("Database connection established")

	// Auto Migration - Tabloları otomatik oluştur/güncelle Struct'lara bakarak SQL tabloları oluştur
	// GORM modellerimize bakarak tabloları oluşturur
	error = DB.AutoMigrate(
		&models.User{},
		&models.Trip{},
		&models.Expense{},
		&models.Activity{},
		&models.ChatMessage{})
	if error != nil {
		log.Fatal("Failed to migrate database:", error)
	}
	log.Println("Database migrated successfully")

	return nil
}

// GetDB - Veritabanı bağlantısını döner
func GetDatabase() *gorm.DB {
	return DB
}

func CloseDatabase() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
