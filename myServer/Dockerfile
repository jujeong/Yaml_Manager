# 1. Go 베이스 이미지
FROM golang:1.20-alpine

# 2. 작업 디렉토리 설정
WORKDIR /app

# 4. 소스 코드 복사
COPY . .

RUN go mod download

# 5. 애플리케이션 빌드
RUN go build -o main .

# 6. 애플리케이션 실행
CMD ["./main"]
