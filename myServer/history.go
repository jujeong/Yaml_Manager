package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// /workload_history 관련 라우터 등록
func RegisterHistoryRoutes(r *gin.Engine) {
	// HTML 페이지 제공
	r.StaticFile("/workload_history", "./workload_history.html")

	// 데이터 API 엔드포인트
	r.GET("/workload_history/data", handleGetWorkloadHistoryData)
}

// /workload_history/data 요청 처리 핸들러
func handleGetWorkloadHistoryData(c *gin.Context) {
	name := c.DefaultQuery("name", "")
	startDate := c.DefaultQuery("start_date", "")
	endDate := c.DefaultQuery("end_date", "")

	// 페이지 번호와 한 페이지당 항목 수를 쿼리 파라미터로 받음
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "20")

	// 페이지 번호와 한 페이지당 항목 수를 정수로 변환
	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		pageInt = 1
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt < 1 {
		limitInt = 20
	}

	// OFFSET 계산
	offset := (pageInt - 1) * limitInt

	// 기본 쿼리 시작
	query := "SELECT workload_name, yaml, metadata, created_timestamp FROM workload_info WHERE 1=1"
	countQuery := "SELECT COUNT(*) FROM workload_info WHERE 1=1"

	// "workload_name" 필터링 (대소문자 구분 없이 부분 일치)
	var conditions []string
	var args []interface{}

	if name != "" {
		conditions = append(conditions, "LOWER(workload_name) LIKE LOWER(?)")
		args = append(args, "%"+name+"%")
	}

	// 날짜 필터링 (시작 날짜 및 종료 날짜)
	if startDate != "" {
		conditions = append(conditions, "DATE(created_timestamp) >= ?")
		args = append(args, startDate)
	}
	if endDate != "" {
		conditions = append(conditions, "DATE(created_timestamp) <= ?")
		args = append(args, endDate)
	}

	// 조건 추가
	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
		countQuery += " AND " + strings.Join(conditions, " AND ")
	}

	// 최신 항목을 먼저 표시하려면 created_timestamp 기준으로 내림차순 정렬 추가
	query += " ORDER BY created_timestamp DESC"

	// LIMIT과 OFFSET을 쿼리에 추가
	query += " LIMIT ? OFFSET ?"
	args = append(args, limitInt, offset)

	// 총 항목 수 조회
	var totalCount int
	err = db.QueryRow(countQuery, args[:len(args)-2]...).Scan(&totalCount) // LIMIT과 OFFSET 제외
	if err != nil {
		log.Println("Error counting rows:", err)
		c.JSON(500, gin.H{"error": "Failed to count rows"})
		return
	}

	// 데이터베이스 쿼리 실행
	rows, err := db.Query(query, args...)
	if err != nil {
		log.Println("Error executing query:", err)
		c.JSON(500, gin.H{"error": "Database query failed"})
		return
	}
	defer rows.Close()

	var results []map[string]interface{}

	// 쿼리 결과 처리
	index := offset + 1 // Index는 현재 페이지의 첫 항목부터 시작
	for rows.Next() {
		var workloadName, yaml, metadata, createdTimestamp string
		if err := rows.Scan(&workloadName, &yaml, &metadata, &createdTimestamp); err != nil {
			log.Println("Error scanning row:", err)
			continue
		}

		// YAML 내용 생략 처리 (최대 20자)
		if len(yaml) > 20 {
			yaml = fmt.Sprintf("%s...", yaml[:20])
		}

		// 결과 저장
		results = append(results, map[string]interface{}{
			"index":            index,
			"workload_name":    workloadName,
			"yaml":             yaml,
			"metadata":         metadata,
			"created_timestamp": createdTimestamp,
		})
		index++
	}

	// 결과 반환
	c.JSON(200, gin.H{
		"total_count": totalCount,     // 총 항목 수
		"current_page": pageInt,      // 현재 페이지 번호
		"total_pages": (totalCount + limitInt - 1) / limitInt, // 총 페이지 수
		"results":     results,       // 검색 결과
	})
}
