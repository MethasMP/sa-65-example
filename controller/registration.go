package controller

import (
	"net/http"

	"github.com/MaeMethas/sa-65-example/entity"
	"github.com/gin-gonic/gin"
)

// POST /registrations
func CreateRegistration(c *gin.Context) {
	var registration entity.Registration

	if err := c.ShouldBindJSON(&registration); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ตรวจสอบว่า Subject มีอยู่จริง
	var subject entity.Subject
	if registration.SubjectID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SubjectID is required"})
		return
	}
	if tx := entity.DB().First(&subject, *registration.SubjectID); tx.RowsAffected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Subject not found"})
		return
	}

	// ตรวจสอบว่า State มีอยู่จริง
	var state entity.State
	if registration.StateID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "StateID is required"})
		return
	}
	if tx := entity.DB().First(&state, *registration.StateID); tx.RowsAffected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "State not found"})
		return
	}

	// ดึงข้อมูล Student จาก token
	sID, exists := c.Get("s_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var student entity.Student
	if err := entity.DB().Where("s_id = ?", sID).First(&student).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
		return
	}

	// ตั้งค่า StudentID
	registration.StudentID = &student.ID

	// ตรวจสอบว่าลงทะเบียนเรียนวิชานี้ไปแล้วหรือไม่
	var existingReg entity.Registration
	if err := entity.DB().Where("student_id = ? AND subject_id = ?", student.ID, *registration.SubjectID).First(&existingReg).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Already registered for this subject"})
		return
	}

	// บันทึกการลงทะเบียน
	if err := entity.DB().Create(&registration).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// โหลดข้อมูลที่เกี่ยวข้องกลับมา
	entity.DB().Preload("Student").Preload("Subject").Preload("State").First(&registration, registration.ID)

	c.JSON(http.StatusCreated, gin.H{"data": registration})
}

// GET /registration/:id

func GetRegistration(c *gin.Context) {

	var registration entity.Registration

	id := c.Param("id")

	//ค้นหา registration ด้วย id
	if err := entity.DB().Raw("SELECT * FROM registrations WHERE id = ?", id).Scan(&registration).Error; err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return

	}

	c.JSON(http.StatusOK, gin.H{"data": registration})

}

// List /registrations

func ListRegistration(c *gin.Context) {

	var registrations []entity.Registration

	if err := entity.DB().Preload("Student").Preload("Subject").Preload("State").Raw("SELECT * FROM registrations").Find(&registrations).Error; err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return

	}

	c.JSON(http.StatusOK, gin.H{"data": registrations})

}

// DELETE /users/:id

func DeleteRegistration(c *gin.Context) {

	id := c.Param("id")

	if tx := entity.DB().Exec("DELETE FROM registrations WHERE id = ?", id); tx.RowsAffected == 0 {

		c.JSON(http.StatusBadRequest, gin.H{"error": "registration not found"})

		return

	}

	c.JSON(http.StatusOK, gin.H{"data": id})

}

// PATCH /registrations
func UpdateRegistration(c *gin.Context) {
	var registration entity.Registration

	if err := c.ShouldBindJSON(&registration); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if tx := entity.DB().Where("id = ?", registration.ID).First(&registration); tx.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Registration not found"})
		return
	}

	// ตรวจสอบว่า Subject มีอยู่จริง (ถ้ามีการส่งมา)
	if registration.SubjectID != nil {
		var subject entity.Subject
		if tx := entity.DB().First(&subject, *registration.SubjectID); tx.RowsAffected == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Subject not found"})
			return
		}
	}

	// ตรวจสอบว่า State มีอยู่จริง (ถ้ามีการส่งมา)
	if registration.StateID != nil {
		var state entity.State
		if tx := entity.DB().First(&state, *registration.StateID); tx.RowsAffected == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "State not found"})
			return
		}
	}

	// บันทึกการอัปเดต
	if err := entity.DB().Save(&registration).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// โหลดข้อมูลที่เกี่ยวข้องกลับมา
	entity.DB().Preload("Student").Preload("Subject").Preload("State").First(&registration, registration.ID)

	c.JSON(http.StatusOK, gin.H{"data": registration})
}
