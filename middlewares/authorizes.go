package middlewares

import (
	"net/http"
	"os"
	"strings"

	"github.com/MaeMethas/sa-65-example/service"
	"github.com/gin-gonic/gin"
)

// validates token
func Authorizes() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientToken := c.Request.Header.Get("Authorization")
		if clientToken == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "No Authorization header provided"})
			return
		}

		extractedToken := strings.Split(clientToken, "Bearer ")

		if len(extractedToken) == 2 {
			clientToken = strings.TrimSpace(extractedToken[1])
		} else {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Incorrect Format of Authorization Token"})
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

		jwtWrapper := service.JwtWrapper{
			SecretKey: secretKey,
			Issuer:    issuer,
		}

		claims, err := jwtWrapper.ValidateToken(clientToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		
		// เก็บข้อมูล student ใน context เพื่อใช้งานต่อไป
		c.Set("student_id", claims.StudentID)
		c.Set("s_id", claims.S_ID)
		c.Next()
	}

}
