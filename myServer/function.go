package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	ys "main/ystruct"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/yaml.v2"
)

var BASE_URL = loadEnvVariable()

// 환경 변수 로딩 함수
func loadEnvVariable() string {
	return "http://" + os.Getenv("KWARE_IP") + ":" + os.Getenv("KWARE_PORT") + os.Getenv("KWARE_PATH")
}

// 에러 처리를 간단하게 만드는 함수
func handleError(c *gin.Context, status int, message string, err error) {
	log.Printf("%s: %v", message, err)
	c.JSON(status, gin.H{"error": message})
}

// GET 요청 처리 함수
func handleGetRequest(c *gin.Context) {
	var results []ys.WorkloadInfo
	query := "SELECT workload_name, yaml, metadata, created_timestamp FROM workload_info"

	rows, err := db.Query(query)
	if err != nil {
		handleError(c, 500, "Database query failed", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var result ys.WorkloadInfo
		if err := rows.Scan(&result.WorkloadName, &result.YAML, &result.Metadata, &result.CreatedTimestamp); err != nil {
			handleError(c, 500, "Row scan failed", err)
			return
		}
		results = append(results, result)
	}

	c.JSON(200, gin.H{"respond": results})
}

// POST 요청 처리 함수
func handlePostRequest(c *gin.Context) {
	var requestData ys.RequestData

	if err := c.ShouldBindJSON(&requestData); err != nil {
		handleError(c, http.StatusBadRequest, "Invalid request data", err)
		return
	}

	// 타임스탬프 설정
	setTimestampIfNotProvided(&requestData)

	// 메타데이터를 JSON으로 변환
	metadataJSON, err := json.Marshal(requestData.Metadata)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to serialize metadata", err)
		return
	}

	// 리소스 할당 정보 요청 및 최종 워크로드 YAML 생성
	ackBody := ReqResourceAllocInfo(BASE_URL, requestData.Yaml)
	finalYaml, clusterValue := createFinalWorkloadYAML(ackBody, requestData.Yaml)

	// YAML을 Base64로 인코딩
	finalYamlBase64 := encodeToBase64(finalYaml)

	// POST 요청 전송
	if err := sendPostRequest(clusterValue, finalYamlBase64); err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to send POST request", err)
		return
	}

	// DB에 저장
	if _, err := db.Exec("INSERT INTO workload_info (workload_name, yaml, metadata, created_timestamp) VALUES (?, ?, ?, ?)",
		requestData.Metadata["name"], finalYamlBase64, string(metadataJSON), requestData.Timestamp); err != nil {
		handleError(c, http.StatusInternalServerError, "Failed to insert into database", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

// 타임스탬프 설정 함수
func setTimestampIfNotProvided(requestData *ys.RequestData) {
	if requestData.Timestamp == "" {
		loc, err := time.LoadLocation("Asia/Seoul")
		if err != nil {
			requestData.Timestamp = time.Now().UTC().Format("2006-01-02 15:04:05")
		} else {
			requestData.Timestamp = time.Now().In(loc).Format("2006-01-02 15:04:05")
		}
	}
}

// Base64 인코딩 함수
func encodeToBase64(data interface{}) string {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		log.Fatalf("Error marshaling YAML: %v", err)
	}
	return base64.StdEncoding.EncodeToString(yamlData)
}

// POST 요청 전송 함수
func sendPostRequest(clusterValue string, finalYamlBase64 string) error {
	address := "http://" + os.Getenv("WRAPPER_IP") + ":" + os.Getenv("WRAPPER_PORT") + os.Getenv("WRAPPER_PATH")

	postData := map[string]string{
		"cluster": clusterValue,
		"yaml":    finalYamlBase64,
	}

	postJSON, err := json.Marshal(postData)
	if err != nil {
		return fmt.Errorf("failed to create JSON for POST request: %v", err)
	}

	resp, err := http.Post(address, "application/json", bytes.NewBuffer(postJSON))
	if err != nil {
		return fmt.Errorf("failed to send POST request: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	log.Printf("Response Body: %s", bodyBytes)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to submit data, status code: %d", resp.StatusCode)
	}
	return nil
}

// 리소스 할당 정보 요청 함수
func ReqResourceAllocInfo(argAddr string, encodedYaml string) ys.RespResource {
	decodedYaml, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encodedYaml))
	if err != nil {
		log.Printf("Failed to decode base64 data: %s", err)
		return ys.RespResource{}
	}

	var workflow ys.Workflow
	if err := yaml.Unmarshal(decodedYaml, &workflow); err != nil {
		log.Printf("Failed to unmarshal YAML data: %s", err)
		return ys.RespResource{}
	}

	reqJson := prepareResourceRequest(workflow)

	// 리소스 요청 전송
	ack, body := SEND_REST_DATA(argAddr, reqJson)

	log.Printf("[ReqResource] Status code: %d, Response body: %s", ack.StatusCode, body)

	if ack.StatusCode != http.StatusOK {
		fmt.Printf("[ReqResource] Request failed with status: %s\n", ack.Status)
		return ys.RespResource{}
	}

	var ackBody ys.RespResource
	if err := json.Unmarshal([]byte(body), &ackBody); err != nil {
		log.Printf("Failed to unmarshal response body: %s", err)
	}

	return ackBody
}

// 리소스 요청 JSON 준비 함수
func prepareResourceRequest(workflow ys.Workflow) ys.ReqResource {
	uuid := "dmkim"
	nowTime := time.Now().Format("2006-01-02 15:04:05")

	reqJson := ys.ReqResource{
		Version: "0.12",
		Request: ys.Request{
			Name: workflow.Metadata.GenerateName,
			ID:   uuid,
			Date: nowTime,
		},
	}

	for _, template := range workflow.Spec.Templates {
		if template.Container != nil {
			reqJson.Request.Containers = append(reqJson.Request.Containers, ys.Container{
				Name: template.Name,
				Resources: ys.Resources{
					Requests: ys.ResourceDetails{
						CPU:              template.Container.Resources.Requests.CPU,
						GPU:              template.Container.Resources.Requests.GPU,
						Memory:           template.Container.Resources.Requests.Memory,
						EphemeralStorage: template.Container.Resources.Requests.EphemeralStorage,
					},
					Limits: ys.ResourceDetails{
						CPU:              template.Container.Resources.Limits.CPU,
						GPU:              template.Container.Resources.Limits.NvidiaGPU,
						Memory:           template.Container.Resources.Limits.Memory,
						EphemeralStorage: template.Container.Resources.Limits.EphemeralStorage,
					},
				},
			})
		}
	}

	reqJson.Request.Attribute = ys.Attribute{
		WorkloadType:     "ML",
		IsCronJob:        true,
		DevOpsType:       "DEV",
		GPUDriverVersion: 12.34,
		CudaVersion:      342.12,
		WorkloadFeature:  "test",
		UserID:           uuid,
		Yaml:             base64.StdEncoding.EncodeToString([]byte(workflow.Metadata.GenerateName)),
	}

	return reqJson
}

// 최종 워크로드 YAML 생성 함수
func createFinalWorkloadYAML(argBody ys.RespResource, inputYaml string) (map[string]interface{}, string) {
	clusterValue := argBody.Response.Cluster
	yamlFile, err := base64.StdEncoding.DecodeString(inputYaml)
	if err != nil {
		log.Fatalf("Failed to decode base64 input YAML: %s", err)
	}

	var yamlData map[string]interface{}
	if err := yaml.Unmarshal(yamlFile, &yamlData); err != nil {
		log.Fatalf("Failed to unmarshal YAML: %s", err)
	}

	return yamlData, clusterValue
}
