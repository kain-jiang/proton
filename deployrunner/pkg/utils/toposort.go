package utils

import (
	"github.com/sirupsen/logrus"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

func TopoSort(tasks map[string][]string, log logrus.FieldLogger) ([]string, error) {
	log.Debug("开始拓扑排序")
	g := simple.NewDirectedGraph()
	nodeIDs := make(map[string]int64)
	idCounter := int64(1)
	for task := range tasks {
		nodeIDs[task] = idCounter
		idCounter++
	}

	for task, id := range nodeIDs {
		g.AddNode(simple.Node(id))
		log.Debugf("添加节点: %s (ID: %d)\n", task, id)
	}

	for task, deps := range tasks {
		for _, dep := range deps {
			from := nodeIDs[dep] // 依赖项
			to := nodeIDs[task]  // 当前任务
			edge := simple.Edge{F: simple.Node(from), T: simple.Node(to)}
			g.SetEdge(edge)
			log.Debugf("添加依赖: %s -> %s\n", dep, task)
		}
	}

	// 6. 执行拓扑排序
	sortedNodes, err := topo.Sort(g)
	if err != nil {
		log.Debugf("发现循环依赖:", err)
		return nil, err
	}

	// 7. 创建反向ID映射便于输出
	idToTask := make(map[int64]string)
	for task, id := range nodeIDs {
		idToTask[id] = task
	}

	result := make([]string, 0)
	// 8. 输出排序结果
	log.Debugf("\n任务执行顺序:")
	for _, node := range sortedNodes {
		log.Debugf("ID: %d, Task: %s", node.ID(), idToTask[node.ID()])
		if task, ok := idToTask[node.ID()]; ok {
			result = append(result, task)
		}
	}
	return result, nil
}
