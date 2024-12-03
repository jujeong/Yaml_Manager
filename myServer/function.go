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

var BASE_URL = "http://" + os.Getenv("KWARE_IP") + ":" + os.Getenv("KWARE_PORT") + os.Getenv("KWARE_PATH")

// GET 요청 처리 함수
func handleGetWorkloadinfoRequest(c *gin.Context) {
	var results []ys.WorkloadInfo // 구조체를 사용하여 결과 슬라이스 정의

	query := "SELECT workload_name, yaml, metadata, created_timestamp FROM workload_info"
	rows, err := db.Query(query)
	if err != nil {
		c.JSON(500, gin.H{"error": "Database query failed"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var result ys.WorkloadInfo // 공통 구조체를 사용
		if err := rows.Scan(&result.WorkloadName, &result.YAML, &result.Metadata, &result.CreatedTimestamp); err != nil {
			c.JSON(500, gin.H{"error": "Row scan failed"})
			return
		}
		results = append(results, result) // 동일한 구조체 타입을 사용
	}

	// 응답 구조 수정
	response := gin.H{
		"respond": results,
	}

	c.JSON(200, response)
}

func handleGetStratoRequest(c *gin.Context) {
	var results []ys.Strato

	query := "SELECT mlid, yaml, data FROM strato"
	rows, err := db.Query(query)
	if err != nil {
		c.JSON(500, gin.H{"error": "Database query failed"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var result ys.Strato
		if err := rows.Scan(&result.MlId, &result.YAML, &result.Data); err != nil {
			c.JSON(500, gin.H{"error": "Row scan failed"})
			return
		}
		results = append(results, result) // 동일한 구조체 타입을 사용
	}

	// 응답 구조 수정
	response := gin.H{
		"respond": results,
	}

	c.JSON(200, response)
}
// POST 요청 처리 함수
func handlePostRequest(c *gin.Context) {
	var requestData ys.RequestData

	// 요청 데이터 바인딩
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 클라이언트가 timestamp를 제공하지 않는 경우 현재 시간으로 설정
	if requestData.Timestamp == "" {
		loc, err := time.LoadLocation("Asia/Seoul")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load location"})
			return
		}
		requestData.Timestamp = time.Now().In(loc).Format("2006-01-02 15:04:05")
	}

	// metadata를 JSON 문자열로 변환
	metadataJSON, err := json.Marshal(requestData.Metadata)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize metadata"})
		return
	}

	// Resource Allocation 정보 요청
	ackBody := ReqResourceAllocInfo(BASE_URL, requestData.Yaml) // 디코딩하지 않고 원본 YAML 그대로 전달

	// 최종 워크로드 YAML 생성
	finalYaml, clusterValue := MadeFinalWorkloadYAML(ackBody, requestData.Yaml)
	// finalYaml을 YAML 문자열로 변환
	finalYamlYAML, err := yaml.Marshal(finalYaml)
	if err != nil {
		log.Fatalf("Error marshaling final YAML: %v", err)
	}
	// finalYaml을 Base64로 인코딩
	finalYamlBase64 := base64.StdEncoding.EncodeToString(finalYamlYAML)
	// POST 요청 보내기
	err = sendPostRequest(clusterValue, finalYamlBase64)
	if err != nil {
		log.Printf("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send POST request"})
		return
	}
	// POST 요청이 성공한 경우에만 요청 데이터를 DB에 저장
	_, err = db.Exec("INSERT INTO workload_info (workload_name, yaml, metadata, created_timestamp) VALUES (?, ?, ?, ?)",
		requestData.Metadata["name"], finalYamlBase64, string(metadataJSON), requestData.Timestamp)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

func sendPostRequest(clusterValue string, finalYamlBase64 string) error {
	wrapperIp := os.Getenv("WRAPPER_IP")
	wrapperPort := os.Getenv("WRAPPER_PORT")
	wrapperPath := os.Getenv("WRAPPER_PATH")
	address := "http://" + wrapperIp + ":" + wrapperPort + wrapperPath
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

	// 응답 본문을 읽고 로그로 출력
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

func ReqResourceAllocInfo(argAddr string, encodedYaml string) ys.RespResource {
	var err error
	encodedYaml = strings.TrimSpace(encodedYaml)
	// Base64 디코딩
	data, err := base64.StdEncoding.DecodeString(encodedYaml)
	if err != nil {
		log.Printf("Failed to decode base64 data: %s", err)
		return ys.RespResource{} // 에러 처리
	}

	var workflow ys.Workflow
	err = yaml.Unmarshal(data, &workflow)
	if err != nil {
		log.Printf("Failed to unmarshal YAML data: %s", err)
		return ys.RespResource{} // 에러 처리
	}

	// 요청할 리소스 JSON 객체 생성
	reqJson := ys.ReqResource{}

	uuid := "dmkim" // 사용자 ID 또는 UUID 설정
	currentTime := time.Now()
	nowTime := currentTime.Format("2006-01-02 15:04:05")

	// reqJson 구성
	reqJson.Version = "0.12"
	reqJson.Request.Name = workflow.Metadata.GenerateName
	reqJson.Request.ID = uuid
	reqJson.Request.Date = nowTime

	// 템플릿을 기반으로 컨테이너 정보를 추가
	for _, value := range workflow.Spec.Templates {
		if value.Container == nil {
			continue
		} else {
			tmpContainer := ys.Container{
				Name: value.Name,
				Resources: ys.Resources{
					Requests: ys.ResourceDetails{
						CPU:              value.Container.Resources.Requests.CPU,
						GPU:              value.Container.Resources.Requests.GPU,
						Memory:           value.Container.Resources.Requests.Memory,
						EphemeralStorage: value.Container.Resources.Requests.EphemeralStorage,
					},
					Limits: ys.ResourceDetails{
						GPU:              value.Container.Resources.Limits.NvidiaGPU,
						CPU:              value.Container.Resources.Limits.CPU,
						Memory:           value.Container.Resources.Limits.Memory,
						EphemeralStorage: value.Container.Resources.Limits.EphemeralStorage,
					},
				},
			}
			reqJson.Request.Containers = append(reqJson.Request.Containers, tmpContainer)
		}
	}

	// 기타 요청 속성 설정
	reqJson.Request.Attribute.WorkloadType = "ML"
	reqJson.Request.Attribute.IsCronJob = true
	reqJson.Request.Attribute.DevOpsType = "DEV"
	reqJson.Request.Attribute.GPUDriverVersion = 12.34
	reqJson.Request.Attribute.CudaVersion = 342.12
	reqJson.Request.Attribute.WorkloadFeature = "test" // 예시
	reqJson.Request.Attribute.UserID = uuid
	reqJson.Request.Attribute.Yaml = base64.StdEncoding.EncodeToString(data) // 다시 인코딩하여 포함

	// 리소스 요청 전송
	var ackBody ys.RespResource
	ack, body := SEND_REST_DATA(argAddr, reqJson)
	// POST 요청에 대한 상태 코드와 본문을 로깅
	log.Printf("[ReqResource] Status code: %d, Response body: %s", ack.StatusCode, body)

	if ack.StatusCode == http.StatusOK {
		// JSON 데이터를 YAML로 변환
		var jsonResponse map[string]interface{}
		err = json.Unmarshal([]byte(body), &jsonResponse)
		if err != nil {
			log.Printf("Failed to unmarshal ack body to JSON: %s", err)
			return ackBody
		}

		yamlData, err := yaml.Marshal(jsonResponse)
		if err != nil {
			log.Printf("Failed to marshal JSON to YAML: %s", err)
			return ackBody
		}

		// 변환된 YAML 데이터 출력 (원하는 방식으로 사용 가능)
		log.Printf("Converted YAML:\n%s", string(yamlData))

		err = yaml.Unmarshal(yamlData, &ackBody) // YAML을 ackBody에 unmarshal
		if err != nil {
			log.Printf("Failed to unmarshal YAML data to ackBody: %s", err)
		}
	} else {
		fmt.Printf("[ReqResource] Request failed with status: %s\n", ack.Status)
	}

	return ackBody
}

func MadeFinalWorkloadYAML(argBody ys.RespResource, inputYaml string) (map[string]interface{}, string) {
	// Base64로 인코딩된 YAML을 디코딩
	clusterValue := argBody.Response.Cluster
	yamlFile, err := base64.StdEncoding.DecodeString(inputYaml)
	if err != nil {
		log.Fatalf("Error decoding Base64 YAML data: %v", err)
	}
	// YAML 데이터를 저장할 변수
	var data map[string]interface{}
	// YAML 데이터 언마샬링
	err = yaml.Unmarshal(yamlFile, &data)
	if err != nil {
		log.Fatalf("Error unmarshalling YAML data: %v", err)
	}
	// templates 섹션에서 모든 container의 image 값을 출력하고 조건에 따라 새로운 키를 추가
	spec, ok := data["spec"].(map[interface{}]interface{})
	if ok {
		templates, ok := spec["templates"].([]interface{})
		if ok {

			for _, template := range templates {
				templateMap, ok := template.(map[interface{}]interface{})
				if ok {

					for _, val := range argBody.Response.Containers {
						if templateMap["name"] == val.Name {
							templateMap["nodeSelector"] = ys.NodeSelect{
								Node: val.Node,
							}
						}
					}

				}
			}
		}
	}
	return data, clusterValue
}
