package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

// 데이터베이스 핸들러
var db *sql.DB

func main() {
	// Gin 라우터 설정
	r := gin.Default()

	var err error
	dsn := os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASSWORD") +
		"@tcp(" + os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT") + ")/" + os.Getenv("DB_NAME")

	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},                      // 허용할 도메인
		AllowMethods:     []string{"GET", "POST", "OPTIONS"}, // 허용할 HTTP 메서드
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "*"},
		AllowCredentials: true, // 쿠키 허용 여부
	}))
	r.GET("/workload_info", handleGetWorkloadinfoRequest)
	r.GET("/strato", handleGetStratoRequest)
	r.POST("/submit", handleSubmitRequest)
	r.POST("/submit_resource", handleSubmitResourceRequest)

	r.Run("0.0.0.0:8080")
}
