# 1. Go 베이스 이미지
FROM golang:1.20-alpine

# 2. 작업 디렉토리 설정
WORKDIR /app

# 3. Go 모듈 파일 복사 및 종속성 설치
COPY go.mod go.sum ./
RUN go mod download

# 4. 소스 코드 복사
COPY . .

# 5. 애플리케이션 빌드
RUN go build -o main .

# 6. 애플리케이션 실행
CMD ["./main"]
