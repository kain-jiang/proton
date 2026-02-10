package componentmanage

import (
	"context"
	"errors"
	"net"
	"strconv"
	"time"

	core_v1 "k8s.io/api/core/v1"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/clock"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/rds/mgmt/v1alpha1"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
)

func (m *Applier) RDS_MGMTClient(kube client.Client) (v1alpha1.Interface, error) {
	if m.NewCfg.Proton_mariadb == nil {
		return nil, errors.New("won't install internal mariadb")
	}
	ctx := context.TODO()
	log := logger.NewLogger()
	clk := new(clock.RealClock)
	name := types.NamespacedName{Namespace: configuration.GetProtonResourceNSFromFile(), Name: mariadbManagementServiceName}
	svc := new(core_v1.Service)
	err := kube.Get(ctx, name, svc)
	for i := 0; i < 8 && (api_errors.IsNotFound(err) || svc.Spec.ClusterIP == ""); i++ {
		log.WithError(err).WithField("name", name).WithField("clusterIP", svc.Spec.ClusterIP).Debug("retry getting rds mgmt service")
		clk.Sleep(time.Second << i)
		err = kube.Get(ctx, name, svc)
	}
	if err == nil && svc.Spec.ClusterIP == "" {
		err = errors.New("spec.clusterIP is missing")
	}
	if err != nil {
		return nil, err
	}
	return v1alpha1.ClientFor(&v1alpha1.Config{
		Host:     net.JoinHostPort(svc.Spec.ClusterIP, strconv.Itoa(mariadbManagementServicePort)),
		Username: m.NewCfg.Proton_mariadb.Admin_user,
		Password: m.NewCfg.Proton_mariadb.Admin_passwd,
	})
}
