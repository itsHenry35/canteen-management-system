package services

import (
	"fmt"
	"log"
	"sync"

	"github.com/itsHenry35/canteen-management-system/models"
	"github.com/itsHenry35/canteen-management-system/utils"
)

// RebuildParentStudentMapping 重建所有家长-学生映射关系
func RebuildParentStudentMapping() error {
	// 获取所有班级ID
	classIDs, err := utils.GetAllClassIDs()
	if err != nil {
		return fmt.Errorf("获取班级列表失败: %v", err)
	}

	// 清空现有的映射关系
	if err := models.ClearAllParentStudentRelations(); err != nil {
		return fmt.Errorf("清空映射关系失败: %v", err)
	}

	log.Printf("开始重建家长-学生映射关系，共有 %d 个班级需要处理", len(classIDs))

	// 使用等待组但有限制并发数
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 2) // 最多2个并发请求，避免触发QPS限制

	// 记录处理结果
	successCount := 0
	failCount := 0
	relationCount := 0
	var resultMutex sync.Mutex

	// 遍历所有班级并获取家长-学生关系
	for i, classID := range classIDs {
		wg.Add(1)
		semaphore <- struct{}{} // 占用一个并发槽

		go func(cid string, index int) {
			defer wg.Done()
			defer func() { <-semaphore }() // 释放并发槽

			// 获取该班级所有的家长-学生关系
			relations, err := utils.GetClassParentStudentRelations(cid)
			if err != nil {
				log.Printf("获取班级 %s 的关系失败: %v", cid, err)
				resultMutex.Lock()
				failCount++
				resultMutex.Unlock()
				return
			}

			if relations == nil || len(relations) == 0 {
				log.Printf("班级 %s (%d/%d) 没有家长-学生关系", cid, index+1, len(classIDs))
				return
			}

			// 保存本班级获取到的计数
			localRelationCount := 0

			// 保存获取到的关系
			for _, rel := range relations {

				// 保存关系到数据库
				err := models.SaveParentStudentRelation(rel.GuardianUserID, rel.StudentUserId)
				if err != nil {
					log.Printf("保存关系失败 (家长: %s, 学生: %d): %v",
						rel.GuardianUserID, rel.StudentUserId, err)
				} else {
					localRelationCount++
				}
			}

			resultMutex.Lock()
			successCount++
			relationCount += localRelationCount
			resultMutex.Unlock()

			log.Printf("班级 %s (%d/%d) 处理完成，保存了 %d 个关系",
				cid, index+1, len(classIDs), localRelationCount)
		}(classID, i)
	}

	// 等待所有操作完成
	wg.Wait()

	log.Printf("家长-学生映射关系重建完成。成功: %d 班级, 失败: %d 班级, 共创建 %d 个关系",
		successCount, failCount, relationCount)

	if failCount > 0 {
		return fmt.Errorf("部分班级(%d/%d)处理失败", failCount, len(classIDs))
	}

	return nil
}
