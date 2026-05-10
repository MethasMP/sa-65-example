package controller

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/MaeMethas/sa-65-example/entity"
	"github.com/MaeMethas/sa-65-example/service"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// LoginPayload login body
type LoginPayload struct {
	S_ID  string `json:"s_id"`
	Phone string `json:"phone"`
}

// SignUpPayload signup body
type SignUpPayload struct {
	Name  string `json:"name"`
	S_ID  string `json:"s_id"`
	Phone string `json:"phone"`
}

// LoginResponse token response
type LoginResponse struct {
	Token string `json:"token"`
	ID    uint   `json:"id"`
	S_ID  string `json:"s_id"`
}

// POST /login
func Login(c *gin.Context) {
	var payload LoginPayload
	var student entity.Student

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// ค้นหา Student ด้วย s_id ที่ผู้ใช้กรอกเข้ามา
	if err := entity.DB().Where("s_id = ?", payload.S_ID).First(&student).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// ตรวจสอบรหัสผ่าน
	err := bcrypt.CompareHashAndPassword([]byte(student.Phone), []byte(payload.Phone))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// โหลดค่าจาก environment variables
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		secretKey = "your-super-secret-jwt-key-change-this-in-production-2024"
	}
	
	issuer := os.Getenv("JWT_ISSUER")
	if issuer == "" {
		issuer = "UniversityRegistrationSystem"
	}
	
	expirationHours := 24
	if expStr := os.Getenv("JWT_EXPIRATION_HOURS"); expStr != "" {
		if exp, err := strconv.Atoi(expStr); err == nil {
			expirationHours = exp
		}
	}

	jwtWrapper := service.JwtWrapper{
		SecretKey:       secretKey,
		Issuer:          issuer,
		ExpirationHours: int64(expirationHours),
	}

	signedToken, err := jwtWrapper.GenerateToken(strconv.Itoa(int(student.ID)), student.S_ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "error signing token"})
		return
	}

	tokenResponse := LoginResponse{
		Token: signedToken,
		ID:    student.ID,
		S_ID:  student.S_ID,
	}

	c.JSON(http.StatusOK, gin.H{"data": tokenResponse})
}

// POST /signup
func CreateStudent(c *gin.Context) {
	var payload SignUpPayload
	var student entity.Student

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ตรวจสอบว่า S_ID ซ้ำหรือไม่
	var existingStudent entity.Student
	if err := entity.DB().Where("s_id = ?", payload.S_ID).First(&existingStudent).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Student ID already exists"})
		return
	}

	// ตรวจสอบว่า Phone ซ้ำหรือไม่
	if err := entity.DB().Where("phone = ?", payload.Phone).First(&existingStudent).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phone number already exists"})
		return
	}

	// เข้ารหัสลับรหัสผ่านที่ผู้ใช้กรอกก่อนบันทึกลงฐานข้อมูล
	hashPhone, err := bcrypt.GenerateFromPassword([]byte(payload.Phone), 14)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "error hashing password"})
		return
	}

	student.Name = payload.Name
	student.S_ID = payload.S_ID
	student.Phone = string(hashPhone)

	if err := entity.DB().Create(&student).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": student})
}
