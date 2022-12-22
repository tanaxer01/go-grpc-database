package database

import (
	postgres "go.elastic.co/apm/module/apmgormv2/v2/driver/postgres"
	//"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"errors"
	"fmt"
	"log"
	"time"

	"os"
)

type ValorMoneda struct {
	Moneda string  `json:"moneda" grom:"primary_key"`
	Fecha  string  `json:"fecha"  gorm:"primary_key"`
	Valor  float32 `json:"valor"`
}

var DB *gorm.DB

func init() {
  dsn := "postgres://postgres:eea5c72fed95692a0b42bd2e3832f041@dokku-postgres-monedas-db:5432/monedas_db"
  /*
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s, sslmode=disable",
		os.Getenv("DOKKU_POSTGRES_OP_NAME"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)*/

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		log.Printf("Error while coneccting to db, %s", err)
		os.Exit(1)
	}

	//defer db.Close()
	log.Printf("[+] Database connected 5432")
	db.AutoMigrate(&ValorMoneda{})

	DB = db
}

func GetValores(moneda string, start time.Time, end time.Time) ([]ValorMoneda, error) {
	var entradas []ValorMoneda
	DB.Where("Moneda = ? AND Fecha > ? AND Fecha < ?", moneda, start, end).Find(&entradas)

	return entradas, nil
}

func AddRowsMonedas(monedas []ValorMoneda) error {
	tx := DB.Begin()

	tx.SavePoint("sp1")
	res := tx.Create(&monedas)
	if res.Error != nil {
		tx.RollbackTo("sp1")
		return errors.New("entrada duplicada")
	}

	tx.Commit()
	log.Printf("[+] Entradas agregadas: %d", len(monedas))
	return nil
}
