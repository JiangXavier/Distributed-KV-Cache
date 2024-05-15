package service

import (
	"context"
	"errors"
	"gorm.io/gorm"
	pb "leicache/api/studentpb"
	"leicache/internal/pkg/student/dao"
	"leicache/utils/logger"
	"math/rand"
	"strconv"
	"time"
)

func NewGroupManager(groupnames []string, currentPeerAddr string) map[string]*Group {
	// 为每一个 group 构造一个 Group 实例
	for i := 0; i < len(groupnames); i++ {
		g := NewGroup(groupnames[i], "lru", 100*2*20, RetrieveFunc(func(key string) ([]byte, error) {
			start := time.Now()
			dao := dao.NewStudentDao(context.Background())
			stus, err := dao.ShowStudentInfo(&pb.StudentRequest{
				Name:  key,
				Score: rand.Float32(),
			})

			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// 优化点：即使没有查询到数据，为了防止恶意攻击，put一个空置并设置合理的过期时间
					return []byte{}, gorm.ErrRecordNotFound
				} else {
					return []byte{}, err
				}
			} else {
				logger.LogrusObj.Infof("成功从后端数据库中查询到学生 %s 的分数：%v", key, stus.Score)
				logger.LogrusObj.Warnf("查询数据库总耗时：%v ms", time.Since(start).Milliseconds())
			}

			return []byte(strconv.FormatFloat(stus.Score, 'f', 2, 64)), nil
		}))
		GroupManager[groupnames[i]] = g
	}
	return GroupManager
}
