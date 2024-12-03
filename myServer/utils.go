package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v2"
)

func check(argErr error) {
	if argErr != nil {
		log.Printf("Error: %v", argErr)
	}
}

func SEND_REST_DATA(argAddr string, argJsonData interface{}) (*http.Response, string) {

	// JSON으로 변환
	jsonData, err := json.Marshal(&argJsonData)
	check(err)

	// 변환된 JSON 데이터를 로그에 출력
	log.Printf("Sending POST request to %s with the following JSON data:\n%s", argAddr, string(jsonData))

	// POST 요청에서 Content-Type을 application/json으로 설정
	resp, err := http.Post(argAddr, "application/json", bytes.NewBuffer(jsonData))
	check(err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	check(err)

	return resp, string(body)
}

func MakeYamlFile(argData interface{}, argPath string) {

	// Write the YAML data to a file
	file, err := os.Create(argPath)
	if err != nil {
		fmt.Printf("Error while creating file: %v\n", err)
		return
	}
	defer file.Close()

	// YAML로 직렬화 (serialize)하고 파일에 저장
	encoder := yaml.NewEncoder(file)
	// encoder.SetIndent(2) // YAML 파일의 가독성을 위해 인덴트를 설정합니다.
	err = encoder.Encode(argData)
	if err != nil {
		log.Fatalf("Error encoding YAML to file: %v", err)
	}

	fmt.Println("YAML file created successfully.")
}
