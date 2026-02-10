package task

import "taskrunner/pkg/utils"

type config = map[string]interface{}

var mergeMaps func(...map[string]interface{}) map[string]interface{} = utils.MergeMaps
