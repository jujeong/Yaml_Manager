package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

// API 요청에 사용할 데이터 구조체
type RequestData struct {
	Yaml      string                 `json:"yaml"`
	Metadata  map[string]interface{} `json:"metadata"` // 동적 JSON 필드를 처리
	Timestamp string                 `json:"timestamp"`
}

// 데이터베이스 핸들러
var db *sql.DB

func main() {
	// Gin 라우터 설정
	r := gin.Default()

	// MySQL 데이터베이스 연결
	var err error
	dsn := os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASSWORD") +
		"@tcp(" + os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT") + ")/" + os.Getenv("DB_NAME")

	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}

	// POST 엔드포인트 정의
	r.POST("/submit", handlePostRequest)

	// 서버 실행
	r.Run("0.0.0.0:8080")
}

// POST 요청 처리 함수
func handlePostRequest(c *gin.Context) {
	var requestData RequestData

	// 요청 데이터 바인딩
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 클라이언트가 timestamp를 제공하지 않는 경우 현재 시간으로 설정
	if requestData.Timestamp == "" {
		// Asia/Seoul 시간대 로드
		loc, err := time.LoadLocation("Asia/Seoul")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load location"})
			return
		}
		// 현재 시간을 Asia/Seoul 시간대로 설정
		requestData.Timestamp = time.Now().In(loc).Format("2006-01-02 15:04:05")
	}

	// metadata를 JSON 문자열로 변환
	metadataJSON, err := json.Marshal(requestData.Metadata)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize metadata"})
		return
	}

	// 요청 데이터 저장
	_, err = db.Exec("INSERT INTO workload_info (workload_name, yaml, metadata, created_timestamp) VALUES (?, ?, ?, ?)",
		requestData.Metadata["name"], requestData.Yaml, string(metadataJSON), requestData.Timestamp)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data submitted successfully"})
}
