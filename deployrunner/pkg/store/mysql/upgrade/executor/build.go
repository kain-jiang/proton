package executor

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	driver "taskrunner/pkg/sql-driver"
	"taskrunner/pkg/store/mysql/upgrade/trait"
	"taskrunner/pkg/utils"
	ttrait "taskrunner/trait"

	"github.com/ghodss/yaml"
)

// TraverseDirectory 遍历目录，返回所有非隐藏文件的路径，递归处理子目录
func TraverseDirectory(root string, svcName string, stage int) ([]trait.Plan, trait.Error) {
	res := []trait.Plan{}
	_, err := os.Stat(root)
	if os.IsNotExist(err) {
		return nil, nil
	}

	dealFile := func(path string, info os.FileInfo, _ error) error {
		fname := filepath.Base(path)
		parts := strings.Split(fname, "_")
		if len(parts) < 3 {
			return errors.Join(trait.ErrUnknowFile, fmt.Errorf("%s is not suuport file name", fname))
		}
		dateID, err := strconv.Atoi(parts[0])
		if err != nil {
			return errors.Join(trait.ErrUnknowFile, err, fmt.Errorf("%s is not suuport file name", fname))
		}
		epoch, err := strconv.Atoi(parts[1])
		if err != nil {
			return errors.Join(trait.ErrUnknowFile, err, fmt.Errorf("%s is not suuport file name", fname))
		}
		p := &trait.Plan{
			PlanMeta: trait.PlanMeta{
				ServiceName: svcName,
				Stage:       stage,
				DateID:      dateID,
				Epoch:       epoch,
			},
		}

		bs, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		err = yaml.Unmarshal(bs, &p.Operators)
		if err != nil {
			return errors.Join(trait.ErrUnknowFile, err, fmt.Errorf("%s is not suuport file name", fname))
		}
		res = append(res, *p)
		return nil
	}

	err = utils.TraverseDirectory(root, dealFile)
	if err != nil {
		return res, &ttrait.Error{
			Internal: ttrait.ECNULL,
			Detail:   fmt.Sprintf("遍历目录%s", root),
			Err:      err,
		}
	}
	return res, nil
}

// BuildSvc 以目录为服务计划文件构建服务的多stage计划
func BuildSvc(root string, svcName string, dbtype string) (trait.StagePlan, trait.Error) {
	stags := []string{trait.PreStage, trait.PostStage, trait.InitStage}
	p := trait.StagePlan{
		SvcName: svcName,
	}
	dbtype0 := driver.ConvertDBType(dbtype)
	if dbtype0 == "MARIADB" {
		dbtype = "mysql"
	} else {
		dbtype = strings.ToLower(dbtype)
	}
	for j, i := range stags {
		pl, err := TraverseDirectory(filepath.Join(root, dbtype, i), svcName, j)
		if err != nil {
			return p, err
		}
		p.Pre[j] = pl
	}
	return p, nil
}

// BuildMultiSvcFromDir 以目录下获取的子目录为服务升级计划文件构建多个服务的计划列表
func BuildMultiSvcFromDir(root string, dbtype string, excludeSvc ...string) (ps trait.Plans, err trait.Error) {
	root0, rerr := filepath.Abs(root)
	if rerr != nil {
		err = &ttrait.Error{
			Internal: ttrait.ECNULL,
			Detail:   fmt.Sprintf("路径错误%s", root),
			Err:      rerr,
		}
		return
	}
	root = root0
	files, rerr := os.ReadDir(root)
	if rerr != nil {
		fmt.Println("读取目录失败:", rerr)
		return ps, &ttrait.Error{
			Internal: ttrait.ECNULL,
			Detail:   fmt.Sprintf("读取目录失败%s", root),
			Err:      rerr,
		}
	}

	index := map[string]bool{}
	for _, s := range excludeSvc {
		index[s] = true
	}
	ps.Plans = make(map[string]trait.StagePlan)

	for _, file := range files {
		// 如果是隐藏文件或目录，跳过
		if file.Name()[0] == '.' {
			continue
		}
		// 只处理文件，不处理子目录
		if file.IsDir() {
			svcName := file.Name()
			if _, ok := index[svcName]; ok {
				// 忽略排除模块
				continue
			}
			p, err := BuildSvc(filepath.Join(root, svcName), svcName, dbtype)
			if err != nil {
				return ps, err
			}
			ps.Plans[svcName] = p
		}
	}
	return ps, nil
}
