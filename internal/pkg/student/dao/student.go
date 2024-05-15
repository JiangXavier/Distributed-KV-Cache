package dao

import (
	"context"
	"gorm.io/gorm"
	stuPb "leicache/api/studentpb"
	"leicache/internal/pkg/student/model"
	"leicache/utils/logger"
)

type StudentDao struct {
	*gorm.DB
}

func NewStudentDao(ctx context.Context) *StudentDao {
	return &StudentDao{NewDBClient(ctx)}
}

func (dao *StudentDao) ShowStudentInfo(req *stuPb.StudentRequest) (r *model.Student, err error) {
	err = dao.Model(&model.Student{}).Where("name=?", req.Name).First(&r).Error
	return
}

func (dao *StudentDao) CreateStudent(req *stuPb.StudentRequest) (err error) {
	var student model.Student

	student.Name = req.Name
	student.Score = float64(req.Score)
	student.Email = req.Email
	student.Grade = req.Grade
	student.PhoneNumber = req.PhoneNumber

	if err = dao.Model(&model.Student{}).Create(&student).Error; err != nil {
		logger.LogrusObj.Error("Insert User Error: ", err.Error())
		return
	}

	return
}
