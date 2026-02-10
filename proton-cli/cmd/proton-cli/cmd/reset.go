/*
Copyright Â© 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/cmd/proton-cli/cmd/utils"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/global"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/reset"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/prometheus"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/store"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/version"
)

var (
	ResetClusterConfigFilePath string
	ClearDataDirs              []string
	FirewallMode               string
	AssumeYes                  bool
)

// * default data directory of service storage
const (
	mariaDBDataPath      = "/sysvol/mariadb"
	mongoDBDataPath      = "/sysvol/mongodb/mongodb_data"
	redisDataPath        = "/sysvol/redis/redis_data"
	nsqDataPath          = "/sysvol/mq-nsq/mq-nsq_data"
	policyEngineDataPath = "/sysvol/policy-engine/policy-engine_data"
	protonEtcdDataPath   = "/sysvol/proton-etcd/proton-etcd_data"
	opensearchDataPath   = "/anyshare/opensearch"
	kafkaDataPath        = "/sysvol/kafka/kafka_data"
	zookeeperDataPath    = "/sysvol/zookeeper/zookeeper_data"
	orientdbDataPath     = "/sysvol/orientdb-master"
)

// resetCmd represents the reset command
var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "reset proton cluster",
	Long: `reset proton cluster. For example:
    proton-cli reset`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		logger.NewLogger().Debugf("%#v", version.Get())

		c := new(configuration.ClusterConfig)
		switch {
		case len(args) != 0:
			nodes, err := utils.NodeListFromIPList(args)
			if err != nil {
				return err
			}
			c.Nodes = nodes
			c.Cr = &configuration.Cr{Local: &configuration.LocalCR{Hosts: args}}
			c.Cs = &configuration.Cs{Provisioner: configuration.KubernetesProvisionerLocal}
			// reset nodeIP æ—¶æ¸…ç�†é»˜è®¤æ•°æ�®ç›®å½•
			setDefaultConfForClearData(c, args)
		case ResetClusterConfigFilePath != "":
			c, err = configuration.LoadFromFile(ResetClusterConfigFilePath)
			if err != nil {
				return err
			}
		default:
			_, k := client.NewK8sClient()
			if k == nil {
				return client.ErrKubernetesClientSetNil
			}
			c, err = configuration.LoadFromKubernetes(context.Background(), k)
			if err != nil {
				return err
			}
		}
		var nodeIPList []string
		for _, node := range c.Nodes {
			nodeIPList = append(nodeIPList, node.IP())
		}
		if FirewallMode != "" {
			c.Firewall.Mode = configuration.FirewallMode(FirewallMode)
		}
		if c.Firewall.Mode == "" {
			c.Firewall.Mode = configuration.FirewallFirewalld
		}
		if AssumeYes {
			return reset.Reset(c)
		}
		prompt := fmt.Sprintf("The operation will reset the cluster on %v", nodeIPList)
		if global.ClearData {
			prompt += ", and clear service storage data dir"
		}
		prompt += ". Continue? (yes/no)"
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(prompt)
		in, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		in = strings.ToLower(strings.Trim(strings.Trim(in, "\n"), " "))
		switch in {
		case "yes":
			return reset.Reset(c)
		case "no":
			return nil
		default:
			return fmt.Errorf("invalid input.")
		}
	},
	Args:               cobra.ArbitraryArgs,
	DisableSuggestions: false,
}

func init() {
	rootCmd.AddCommand(resetCmd)
	resetCmd.Flags().StringVarP(&ResetClusterConfigFilePath, "file", "f", ResetClusterConfigFilePath, "reset the nodes of the cluster config file")
	resetCmd.Flags().BoolVar(&global.ClearData, "clear-data", true, "clear service storage data")
	resetCmd.Flags().StringVar(&FirewallMode, "firewall-mode", "", "how proton manage firewall. firewalld: use firewalld as firewall. usermanaged: the firewall is managed by the user, proton doesn't modify it.")
	resetCmd.Flags().BoolVarP(&AssumeYes, "assumeyes", "y", false, "answer yes for all questions")
}

// set default data path for clear when reseting
func setDefaultConfForClearData(cfg *configuration.ClusterConfig, hosts []string) {
	// opensearch
	cfg.OpenSearch = &configuration.OpenSearch{Data_path: opensearchDataPath}
	// proton_etcd
	cfg.Proton_etcd = &configuration.ProtonDataConf{Data_path: protonEtcdDataPath}
	// proton_mariadb
	cfg.Proton_mariadb = &configuration.ProtonMariaDB{Data_path: mariaDBDataPath}
	// proton_mongodb
	cfg.Proton_mongodb = &configuration.ProtonDB{Data_path: mongoDBDataPath}
	// proton_mq_nsq
	cfg.Proton_mq_nsq = &configuration.ProtonDataConf{Data_path: nsqDataPath}
	// proton_policy_engine
	cfg.Proton_policy_engine = &configuration.ProtonDataConf{Data_path: policyEngineDataPath}
	// proton_redis
	cfg.Proton_redis = &configuration.ProtonDB{Data_path: redisDataPath}
	// zookeeper
	cfg.ZooKeeper = &configuration.ZooKeeper{Data_path: zookeeperDataPath}
	// kafka
	cfg.Kafka = &configuration.Kafka{Data_path: kafkaDataPath}
	// prometheus
	cfg.Prometheus = &configuration.Prometheus{DataPath: prometheus.DefaultDataPath}
	// package store
	cfg.PackageStore = &configuration.PackageStore{Storage: configuration.PackageStoreStorage{Path: store.DefaultStoragePath}}
	// component-manage
	cfg.ComponentManage = &configuration.ComponentManagement{}
}
