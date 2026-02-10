package routers

import (
	"fmt"

	v1 "component-manage/internal/routers/v1"

	"github.com/gin-gonic/gin"
)

func RegistryRouter(e *gin.Engine) error {
	if err := v1.RegistryApi(e); err != nil {
		return fmt.Errorf("registry v1 api error: %w", err)
	}
	registrySwag(e)
	return nil
}
