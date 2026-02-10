package logic

import (
	"fmt"
	"sort"

	"component-manage/internal/global"
	"component-manage/pkg/models/types"
)

func ListAllComponents() ([]*types.ComponentObject, error) {
	// 检查组件已存在
	cpts, err := global.Persist.GetAllComponentObject()
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}

	result := make([]*types.ComponentObject, 0)

	for _, cpt := range cpts {
		result = append(result, cpt)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}
