package base

import (
	"errors"
	"strings"

	"component-manage/internal/pkg/cerr"

	k8sErrs "k8s.io/apimachinery/pkg/api/errors"
)

// MergeHelmValues 合并 helm values
func MergeHelmValues(a, b map[string]interface{}) map[string]interface{} {
	if a == nil && b == nil {
		return nil
	}

	c := make(map[string]interface{})
	for k, v := range a {
		c[k] = v
	}

	for k, v := range b {
		cv, cok := c[k].(map[string]interface{})
		bv, bok := v.(map[string]interface{})
		if cok && bok {
			c[k] = MergeHelmValues(cv, bv)
		} else {
			c[k] = v
		}
	}

	return c
}

func ImageTag(img string) string {
	// 获取最后一个 : 之后的内容
	if idx := strings.LastIndex(img, ":"); idx != -1 {
		return img[idx+1:]
	}
	//  unreachable
	panic("get image version failed")
}

func ImageName(img string) string {
	// 获取最后一个 : 之前的内容
	if idx := strings.LastIndex(img, ":"); idx != -1 {
		return img[:idx]
	}
	//  unreachable
	panic("get image name failed")
}

func DealK8sStatusError(err error) error {
	var statusError *k8sErrs.StatusError
	if errors.As(err, &statusError) {
		status := statusError.ErrStatus
		return cerr.NewError(
			cerr.OriginalProduceErrorWithStatus(status.Code),
			"produce a status error on k8s api",
			statusError.Error(),
		)
	} else {
		return err
	}
}

func DefaultString(val, dft string) string {
	if val == "" {
		return dft
	} else {
		return val
	}
}
