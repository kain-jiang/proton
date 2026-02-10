package backup

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/util/etcd"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/util/file"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/util/tar"

	"github.com/jhunters/goassist/arrayutil"
	jsoniter "github.com/json-iterator/go"
	"github.com/samber/lo"
	"go.etcd.io/etcd/client/v3/snapshot"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	core_v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/yaml"
)

const (
	DefaultBackupDataDir string = "/mnt/backup"
	BackupDir            string = "/opt/proton-backup"
	BackupLogDir         string = "/opt/proton-backup/logs"
	KubernetesEtcdPath   string = "/etc/kubernetes/manifests/etcd.yaml"
	BackupConfName       string = "backup.json"
	ProtonCLIConfName    string = "proton-cli.yaml"
	//k8s 的etcd快照备份目录
	EtcdSnapshotRelPath string = "k8s/etcdSnapshot"
	//proton数据服务所在命名空间
	// NamespaceResource = "resource"
	//proton-mongodb服务名
	ProtonMongoDBServiceName = "mongodb-mgmt-cluster"
	//proton-mariadb服务名
	ProtonMariaDBServiceName = "mariadb-mgmt-cluster"
	//proton-etcd nodeport服务名
	ProtonEtcdNodeServiceName = "proton-etcd-nodeport"
	//proton-etcd secret名
	ProtonEtcdSecretName = "etcdssl-secret"
	//proton-etcd secret名
	ProtonEtcdProtocol = "https"
	//proton-etcd 主机端口
	ProtonEtcdNodePort int32 = 32379
	//proton-etcd 容器端口
	ProtonEtcdPodPort int32 = 2379
	//备份数据目录剩余空间最低要求 GB
	MinimumSpace = 10
	//异步备份数据服务，循环状态接口次数
	BackupCycleCount int = 2880
	//判断备份是否完成等待间隔
	BackupSleepInterval time.Duration = 30 * time.Second
)

type BackupOpts struct {
	Ttl                int
	Resource           []string
	SkipBackupResource []string
	BackupName         string
	Id                 string
	// 多副本 MariaDB，只有此配置为 true 时才备份
	BackupMariaDB bool
	// 多副本 MongoDB，只有此配置为 true 时才备份
	BackupMongoDB bool
	// 备份路径，非空时覆盖 backupConf 的 BackupDirectory
	BackupDirectory string
}

type BackupConf struct {
	HostName        string       `json:"hostName"`
	BackupDirectory string       `json:"backupDirectory"`
	List            []BackupInfo `json:"list"`
}
type BackupInfo struct {
	Id          string          `json:"id"`
	Name        string          `json:"name"`
	CreateTime  int64           `json:"createTime"`
	EndTime     int64           `json:"endTime"`
	RunTime     int64           `json:"runTime"`
	UseSpace    int64           `json:"useSpace"`
	Ttl         int             `json:"ttl"`
	StorageType StorageTypeEnum `json:"storageType"`
	Path        string          `json:"path"`
	LogPath     string          `json:"logPath"`
	Status      bool            `json:"status"`
	IsDelete    bool            `json:"isDelete"`
	Resource    []string        `json:"resource"`
}

// MariaDB请求备份参数
type MariaDBBackupRequest struct {
	BackupDir string `json:"backup_dir"`
}

// Mongo请求备份参数
type MongoDBBackupRequest struct {
	BackupDir string `json:"backup_dir"`
}

// 存储类型
type StorageTypeEnum int32

const (
	LocalStorage   StorageTypeEnum = 0 //本地磁盘存储
	ObjectStorage  StorageTypeEnum = 1 //对象存储
	NetworkStorage StorageTypeEnum = 2 //网络存储
)

// 可备份的资源信息
type ResourceInfo struct {
	Id        int
	RestoreId int
	Name      string
	// Require   bool //true：必须备份，false：不存在跳过备份
	PathList []ResourcePathMapping
}

// 资源和备份目录的路径映射
type ResourcePathMapping struct {
	OsPath         string //资源在操作系统中的原始绝对路径
	TargetRelPath  string //备份的资源在备份文件中的相对路径
	Require        bool   //true：必须备份，false：不存在跳过备份
	ExcludeathList []string
}

// 排除备份恢复的文件或者目录
type ExcludeathPath struct {
	OsPath string //资源在操作系统中的原始绝对路径
}

// mongodb 备份包信息
type MongoDBPackageInfo struct {
	Id          string `json:"id"`
	PackageName string `json:"package_name"`
	ReleaseName string `json:"release_name"`
	ReplSetName string `json:"rs_name"`
	Status      string `json:"status"` //success, running, failed
	StorageNode string `json:"storage_node"`
	CreateTime  string `json:"create_time"`
}

// mariadb 备份包信息
type MariaDBPackageInfo struct {
	Id          string `json:"id"`
	PackageName string `json:"package_name"`
	CreateTime  string `json:"create_time"`
	Status      string `json:"status"`
	StorageNode string `json:"storage_node"`
}

var (
	Backuplog = logger.NewLogger()
	//可备份的资源集合
	ResourceCollection = []ResourceInfo{}
)

func init() {
	ResourceCollection = []ResourceInfo{
		{
			Id:        1,
			RestoreId: 1,
			Name:      "network",
			PathList: []ResourcePathMapping{
				{
					OsPath:        "/etc/sysconfig/network-scripts",
					Require:       true,
					TargetRelPath: "host/network-scripts",
				},
			},
		},
		{
			Id:        2,
			RestoreId: 2,
			Name:      "firewalld",
			PathList: []ResourcePathMapping{
				{
					OsPath:        "/etc/firewalld",
					Require:       true,
					TargetRelPath: "host/firewalld",
				},
			},
		},
		{
			Id:        3,
			Name:      "sysctl",
			RestoreId: 3,
			PathList: []ResourcePathMapping{
				{
					OsPath:        "/etc/sysctl.d",
					Require:       true,
					TargetRelPath: "host/sysctl.d",
				},
			},
		},
		{
			Id:        4,
			Name:      "dns",
			RestoreId: 4,
			PathList: []ResourcePathMapping{
				{
					OsPath:        "/etc/resolv.conf",
					Require:       true,
					TargetRelPath: "host/resolv.conf",
				},
			},
		},
		{
			Id:        5,
			Name:      "hosts",
			RestoreId: 5,
			PathList: []ResourcePathMapping{
				{
					OsPath:        "/etc/hosts",
					Require:       true,
					TargetRelPath: "host/hosts",
				},
			},
		},
		{
			Id:        6,
			RestoreId: 6,
			Name:      "kubernetes",
			PathList: []ResourcePathMapping{
				{
					OsPath:        "/etc/kubernetes",
					Require:       true,
					TargetRelPath: "k8s/kubernetes",
				},
				{
					OsPath:        "/root/.kube",
					Require:       false,
					TargetRelPath: "k8s/kube",
				},
				{
					OsPath:        "/root/.helm",
					Require:       false,
					TargetRelPath: "k8s/helm",
				},
				{
					OsPath:        "/var/lib/kubelet/config.yaml",
					Require:       false,
					TargetRelPath: "k8s/kubelet/config.yaml",
				},
				{
					OsPath:        "/var/lib/kubelet/kubeadm-flags.env",
					Require:       false,
					TargetRelPath: "k8s/kubelet/kubeadm-flags.env",
				},
				{
					OsPath:        "/var/lib/kubelet/pki",
					Require:       false,
					TargetRelPath: "k8s/kubelet/pki",
				},
			},
		},
		{
			Id:        7,
			RestoreId: 7,
			Name:      "docker",
			PathList: []ResourcePathMapping{
				{
					OsPath:        "/etc/docker/daemon.json",
					Require:       false,
					TargetRelPath: "docker/daemon.json",
				},
			},
		},
		{
			Id:        8,
			RestoreId: 8,
			Name:      "proton-slb",
			PathList: []ResourcePathMapping{
				{
					OsPath:        "/etc/slb",
					Require:       false,
					TargetRelPath: "proton-slb/slb",
				},
				{
					OsPath:        "/usr/local/haproxy/haproxy.cfg",
					Require:       false,
					TargetRelPath: "proton-slb/haproxy/haproxy.cfg",
				},
				{
					OsPath:        "/usr/local/slb-nginx",
					Require:       false,
					TargetRelPath: "proton-slb/slb-nginx",
					ExcludeathList: []string{
						"/usr/local/slb-nginx/sbin/slb-nginx",
					},
				},
				{
					OsPath:        "/etc/keepalived/keepalived.conf",
					Require:       false,
					TargetRelPath: "proton-slb/keepalived/keepalived.conf",
				},
			},
		},
		{
			Id:        9,
			RestoreId: 9,
			Name:      "proton-cr",
			PathList: []ResourcePathMapping{
				{
					OsPath:        "/etc/proton-cr/proton-cr.yaml",
					Require:       false,
					TargetRelPath: "proton-cr/proton-cr.yaml",
				},
				{
					OsPath:        "/etc/chartmuseum/config.yaml",
					Require:       false,
					TargetRelPath: "proton-cr/chartmuseum/config.yaml",
				},
				{
					OsPath:        "/etc/docker/registry/config.yml",
					Require:       false,
					TargetRelPath: "proton-cr/registry/config.yml",
				},
			},
		},
		{
			Id:        10,
			RestoreId: 10,
			Name:      "eceph",
			PathList: []ResourcePathMapping{
				{
					OsPath:        "/opt/minotaur/config/minotaur.conf",
					Require:       false,
					TargetRelPath: "eceph/minotaur.conf",
				},
				{
					OsPath:        "/var/lib/ceph/mgr",
					Require:       false,
					TargetRelPath: "eceph/mgr",
				},
				{
					OsPath:        "/etc/ceph",
					Require:       false,
					TargetRelPath: "eceph/ceph",
				},
			},
		},
		{
			Id:        11,
			RestoreId: 14,
			Name:      "proton-mariadb",
			PathList: []ResourcePathMapping{

				{
					Require:       false,
					TargetRelPath: "proton-mariadb",
				},
			},
		},
		{
			Id:        12,
			RestoreId: 13,
			Name:      "proton-mongodb",
			PathList: []ResourcePathMapping{
				{
					Require:       false,
					TargetRelPath: "proton-mongodb",
				},
			},
		},
		{
			Id:        13,
			RestoreId: 12,
			Name:      "proton-etcd",
			PathList: []ResourcePathMapping{
				{
					Require:       false,
					TargetRelPath: "proton-etcd",
				},
			},
		},
		{
			Id:        14,
			RestoreId: 11,
			Name:      "kubernetes-etcd",
			PathList: []ResourcePathMapping{
				{
					Require:       false,
					TargetRelPath: "kubernetes-etcd",
				},
			},
		},
	}

	// 用id排序,数值大的排在前
	sort.SliceStable(ResourceCollection, func(i, j int) bool {
		return ResourceCollection[i].Id > ResourceCollection[j].Id
	})

}

// 检测备份的配置文件是否存在，不存在创建
func CheckBackupConfig() error {
	exist, err := file.PathExists(BackupLogDir)
	if err != nil {
		Backuplog.Errorln("检测备份配置日志目录是否存在异常：", BackupLogDir, err)
		return err
	}
	if !exist {
		err := os.MkdirAll(BackupLogDir, os.ModePerm)
		if err != nil {
			Backuplog.Errorln("创建备份配置目录异常：", BackupLogDir, err)
			return err
		}
	}
	var backupConf = filepath.Join(BackupDir, BackupConfName)
	exist, err = file.PathExists(backupConf)
	if err != nil {
		Backuplog.Errorln("检测备份配置文件是否存在异常：", backupConf, err)
		return err
	}
	if !exist {
		f, err := os.Create(backupConf)
		f.Close()
		if err != nil {
			Backuplog.Errorln("创建备份配置文件异常：", backupConf, err)
			return err
		}
		//如果没有备份的配置文件，检测备份存储的数据目录是否存在备份的配置，存在的话把目录复制到主机的/opt/proton-backup备份配置的目录中
		proton_backup_backup_dir := filepath.Join(DefaultBackupDataDir, BackupDir)
		exist, err = file.PathExists(proton_backup_backup_dir)
		if err != nil {
			Backuplog.Errorln("检测proton_backup的备份配置目录是否存在异常：", proton_backup_backup_dir, err)
			return err
		}
		if exist {
			//当系统盘故障，可使用该备份目录还原proton-cli的备份和还原的记录以及日志
			err = file.Copy(proton_backup_backup_dir, BackupDir, Backuplog, nil)
			Backuplog.Infoln(BackupDir + "没有备份记录，目标备份数据目录:" + proton_backup_backup_dir + "下存在备份的配置文件目录时，复制到:" + BackupDir)

			if err != nil {
				return err
			}
		}
	}
	return nil
}

// 获取proton的备份配置对象
func GetBackupConf() (*BackupConf, error) {
	err := CheckBackupConfig()
	if err != nil {
		return nil, err
	}
	var backupConf = filepath.Join(BackupDir, BackupConfName)
	hostname, _ := os.Hostname()
	var conf = BackupConf{}
	content, err := os.ReadFile(backupConf)
	if err != nil {
		Backuplog.Errorln("打开备份配置文件异常：", backupConf, err)
		return nil, err
	}
	if len(content) == 0 {
		conf.HostName = hostname
		conf.BackupDirectory = DefaultBackupDataDir
	} else {
		err = json.Unmarshal([]byte(string(content)), &conf)
		if err != nil {
			return nil, err
		}
	}
	exist, err := file.PathExists(conf.BackupDirectory)
	if err != nil {
		Backuplog.Errorln("检测备份工作目录是否存在异常：", conf.BackupDirectory, err)
		return nil, err
	}
	if !exist {
		err := os.MkdirAll(conf.BackupDirectory, os.ModePerm)
		if err != nil {
			Backuplog.Errorln("创建备份工作目录异常：", conf.BackupDirectory, err)
			return nil, err
		}
	}
	return &conf, nil
}

// 判断元素是否存在数组中
func IsContain(items []string, item string) bool {
	for _, eachItem := range items {
		if strings.EqualFold(eachItem, item) {
			return true
		}
	}
	return false
}

// 数组删除重复
func removeDuplicateElement(languages []string) []string {
	result := make([]string, 0, len(languages))
	temp := map[string]struct{}{}
	for _, item := range languages {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

// 创建备份
func CreateBackUp(opt BackupOpts) error {

	conf, err := GetBackupConf()
	if err != nil {
		return err
	}

	// 来自命令行参数的 opt.BackupDirectory 覆盖来自配置文件的
	// conf.BackupDirectory
	var persistentBackupDirectory = conf.BackupDirectory
	if opt.BackupDirectory != "" {
		conf.BackupDirectory = opt.BackupDirectory
	}

	freeSpace, err := GetFreeSpaceGB(conf.BackupDirectory)
	if err != nil {
		return err
	}

	if freeSpace < MinimumSpace {
		return fmt.Errorf("backup data directory: %v, remaining free space %.2f GB,Below minimum space requirement：%d", conf.BackupDirectory, freeSpace, MinimumSpace)
	}

	var info = BackupInfo{
		Id:          opt.Id,
		Name:        opt.BackupName,
		CreateTime:  time.Now().Unix(),
		Ttl:         opt.Ttl,
		StorageType: LocalStorage,
		Status:      false,
		IsDelete:    false,
	}
	if err != nil {
		return err
	}
	var backupinfo = arrayutil.Filter(conf.List, func(s1 BackupInfo) bool { return s1.Name != info.Name })
	if len(backupinfo) > 0 {
		return errors.New("duplicate backup name:" + info.Name)
	}
	opt.Resource = removeDuplicateElement(opt.Resource)
	var workDirectory = filepath.Join(conf.BackupDirectory, opt.Id)
	err = os.MkdirAll(workDirectory, os.ModePerm)
	if err != nil {
		return err
	}
	var backError = BackUpResource(opt, conf, &info)
	if backError != nil {
		info.Status = false
		// 备份失败，删除此次的不完整备份的资源
		err = os.RemoveAll(workDirectory)
		if err != nil {
			return err
		}
		return backError
		// Backuplog.Infoln("备份失败：" + backError.Error())
	}
	info.EndTime = time.Now().Unix()
	info.RunTime = int64(time.Unix(info.EndTime, 0).Sub(time.Unix(info.CreateTime, 0)).Seconds())
	info.Status = true
	info.Path = filepath.Join(conf.BackupDirectory, opt.Id+".tar.gz")
	info.LogPath = filepath.Join(BackupLogDir, opt.Id+".log")
	// info.Resource = opt.Resource
	Backuplog.Infoln("备份成功")
	conf.List = append(conf.List, info)
	// conf 可能已经被命令行参数修改，恢复配置文件中的配置
	conf.BackupDirectory = persistentBackupDirectory
	jsonBytes, err := json.Marshal(conf)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(BackupDir, BackupConfName), jsonBytes, 0644)
	if err != nil {
		return err
	}
	//备份/opt/proton-backup目录数据到备份的数据目录下，当系统盘故障，可使用该备份目录还原proton-cli的备份和还原的记录以及日志
	err = file.Copy(BackupDir, filepath.Join(conf.BackupDirectory, BackupDir), Backuplog, nil)
	if err != nil {
		return err
	}
	Backuplog.Infoln("保存备份记录")
	return nil

}
func Check(f func() error) {
	if err := f(); err != nil {
		fmt.Println(err)
	}
}

// 备份当前节点的资源
func BackUpResource(opt BackupOpts, conf *BackupConf, fo *BackupInfo) error {
	_, k := client.NewK8sClient()
	if k == nil {
		return client.ErrKubernetesClientSetNil
	}
	clusterConf, err := configuration.LoadFromKubernetes(context.Background(), k)
	if err != nil {
		return err
	}
	if clusterConf == nil {
		return fmt.Errorf("Unable to get cluster configuration file")
	}

	//过滤出需要备份的资源
	if !IsContain(opt.Resource, "all") {
		var newResourceCollection = []ResourceInfo{}
		for _, col := range ResourceCollection {
			if IsContain(opt.Resource, col.Name) {
				newResourceCollection = append(newResourceCollection, col)
			}
		}
		ResourceCollection = newResourceCollection
	}
	//删除需要跳过的备份资源
	if len(opt.SkipBackupResource) > 0 {
		var newResourceCollection = []ResourceInfo{}
		for _, col := range ResourceCollection {
			if !IsContain(opt.SkipBackupResource, col.Name) {
				newResourceCollection = append(newResourceCollection, col)
			}
		}
		ResourceCollection = newResourceCollection
	}
	// 如果防火墙模式不是 firewalld 则不备份 firewalld
	if clusterConf.Firewall.Mode != configuration.FirewallFirewalld {
		ResourceCollection = lo.Reject(ResourceCollection, func(r ResourceInfo, _ int) bool { return r.Name == "firewalld" })
	}

	if len(ResourceCollection) == 0 {
		return errors.New("请求的备份资源无效:" + strings.Join(opt.Resource, ","))
	}
	var workDirectory = filepath.Join(conf.BackupDirectory, opt.Id)
	defer func() {
		Backuplog.Infof("delete working directory %s", workDirectory)
		err := os.RemoveAll(workDirectory)
		if err != nil {
			Backuplog.Errorf("delete working directory %s failed: %v", workDirectory, err)
		}
	}()

	//master节点和work节点都可以执行备份

	//备份proton-cli集群配置文件
	Backuplog.Infoln("create proton-cli config backup")
	var NamespaceResource = configuration.GetProtonResourceNSFromFile()
	var protoncliPath = filepath.Join(workDirectory, ProtonCLIConfName)
	jsonBytes, err := yaml.Marshal(clusterConf)
	if err != nil {
		return err
	}
	err = os.WriteFile(protoncliPath, jsonBytes, 0644)
	if err != nil {
		return err
	}
	for _, info := range ResourceCollection {
		Backuplog.Infoln("备份" + info.Name)
		var skipbackup = false
		if info.Name == "proton-mariadb" {
			for _, p := range info.PathList {
				Backuplog.Infoln("检测服务是否可备份：" + info.Name)
				if clusterConf == nil {
					return fmt.Errorf("Unable to get cluster configuration file")
				} else if clusterConf.Proton_mariadb == nil {
					if p.Require {
						return fmt.Errorf("Unable to get Proton-mariadb configuration")
					}
					Backuplog.Infoln("Unable to get Proton-mariadb configuration,skip backup Proton-mariadb")
					skipbackup = true
					continue
				} else if clusterConf.Proton_mariadb.Hosts == nil {
					if p.Require {
						return fmt.Errorf("Unable to get Proton-mariadb Hosts configuration")
					}
					Backuplog.Infoln("Unable to get Proton-mariadb Hosts configuration,skip backup Proton-mariadb")
					skipbackup = true
					continue
				} else if len(clusterConf.Proton_mariadb.Hosts) == 0 || (len(clusterConf.Proton_mariadb.Hosts) > 1 && !opt.BackupMariaDB) {
					if p.Require {
						return fmt.Errorf("Proton-mariadb the number of copies is %d, and mariadb data cannot be backed up", len(clusterConf.Proton_mariadb.Hosts))
					}
					Backuplog.Infof("Proton-mariadb the number of copies is %d, and mariadb data cannot be backed up,skip backup Proton-mariadb", len(clusterConf.Proton_mariadb.Hosts))
					skipbackup = true
					continue
				} else if !slices.Contains(clusterConf.Proton_mariadb.Hosts, conf.HostName) {
					if p.Require {
						return fmt.Errorf("The current node not have proton-mariadb")
					}
					Backuplog.Infoln("The current node not have proton-mariadb,skip backup Proton-mariadb")
					skipbackup = true
					continue
				} else if clusterConf.Proton_mariadb.Admin_user == "" || clusterConf.Proton_mariadb.Admin_passwd == "" {
					return fmt.Errorf("Proton-mariadb username or password cannot be empty")
				}

				svc, err := k.CoreV1().Services(NamespaceResource).Get(context.Background(), ProtonMariaDBServiceName, metav1.GetOptions{})
				if err != nil {
					return err
				} else if svc == nil {
					return fmt.Errorf("unable to get mariadb service:" + ProtonMariaDBServiceName)
				}
				var ip = svc.Spec.ClusterIP
				if ip == "" {
					return errors.New("Unable to obtain the management terminal address of mariadb")
				}
				headers := map[string]string{
					"admin-key": base64.StdEncoding.EncodeToString([]byte(clusterConf.Proton_mariadb.Admin_user + ":" + clusterConf.Proton_mariadb.Admin_passwd)),
				}
				var httpclient = client.NewHttpClient(30)

				// get mariadb backup size and compare target dirctory free size
				// if bigger, then return error
				Backuplog.Infoln("check mariadb size and backup directory size")
				if code, resp, err := httpclient.Get(fmt.Sprintf("http://%s:8888/api/proton-rds-mgmt/v2/backup_size", ip), headers); err != nil {
					return fmt.Errorf("unable to get mariadb backup directory size, err: %w", err)
				} else if code != http.StatusOK {
					return fmt.Errorf("unable to get mariadb backup directory size, http status code: %d, response body: %v", code, resp)
				} else {
					freeSpace, err := GetFreeSpaceGB(conf.BackupDirectory)
					if err != nil {
						return fmt.Errorf("unable to get backup directory free size, error: %w", err)
					}
					// dbBackupSize unit is KB
					dbBackupSize := resp.(float64) / 1024 / 1024
					if dbBackupSize >= freeSpace {
						return fmt.Errorf("unable to create mariadb backup, mariadb data size: %f bigger than backup directory free size: %f", dbBackupSize, freeSpace)
					}
				}

				var mariadbRequest = MariaDBBackupRequest{}
				mariadbRequest.BackupDir = filepath.Join(workDirectory, p.TargetRelPath)
				if s, b, err := httpclient.Post(fmt.Sprintf("http://%s:8888/api/proton-rds-mgmt/v2/backups", ip), headers, mariadbRequest); err != nil {
					return fmt.Errorf("unable to create mariadb backup, error: %w", err)
				} else if s != http.StatusAccepted {
					return fmt.Errorf("unable to create mariadb backup, http status code: %d, response body: %v", s, b)
				} else {
					time.Sleep(BackupSleepInterval)
					Backuplog.Infof("create mariadb backup, http status code: %d, response body: %v \n", s, b)
					responseByte, err := jsoniter.Marshal(b)
					if err != nil {
						Backuplog.Infoln("unable to create mariadb backup,The returned data cannot be converted to byte")
						return err
					}
					info := MariaDBPackageInfo{}
					err = jsoniter.Unmarshal(responseByte, &info)
					if err != nil {
						Backuplog.Infoln("unable to create mariadb backup,The returned data cannot be converted to MariaDBPackageInfo")
						return err
					}
					for i := 1; i <= BackupCycleCount; i++ {
						if s, b, err := httpclient.Get(fmt.Sprintf("http://%s:8888/api/proton-rds-mgmt/v2/backups?id=%s", ip, info.Id), headers); err != nil {
							Backuplog.Errorf("unable to get mariadb backup message, error: %v, retry", err)
						} else if s != http.StatusOK {
							return fmt.Errorf("unable to get mariadb backup message, http status code: %d, response body: %v", s, b)
						} else {
							Backuplog.Infof("get mariadb backup message, http status code: %d, response body: %v \n", s, b)
							responseByte, err := jsoniter.Marshal(b)
							if err != nil {
								Backuplog.Infoln("unable to get mariadb backup message,The returned data cannot be converted to byte")
								return err
							}
							desc := MariaDBPackageInfo{}
							err = jsoniter.Unmarshal(responseByte, &desc)
							if err != nil {
								Backuplog.Infoln("unable to get mariadb backup message,The returned data cannot be converted to MariaDBBackupList")
								return err
							}
							if desc.Status == "success" {
								goto MariadbSuccess
							}
							if desc.Status == "failed" {
								return fmt.Errorf("backup mariadb failed, http status code: %d, response body: %v", s, b)
							}
						}
						time.Sleep(BackupSleepInterval)
					}
					return fmt.Errorf("unable to get mariadb backup message：checked %d count", BackupCycleCount)
				}
			MariadbSuccess:
				Backuplog.Infoln("mariadb backup success")
			}
		} else if info.Name == "proton-mongodb" {

			for _, p := range info.PathList {
				if clusterConf == nil {
					return fmt.Errorf("Unable to get cluster configuration file")
				} else if clusterConf.Proton_mongodb == nil {
					if p.Require {
						return fmt.Errorf("Unable to get Proton-mongodb configuration")
					}
					Backuplog.Infoln("Unable to get Proton-mongodb configuration,skip backup Proton-mongodb")
					skipbackup = true
					continue
				} else if clusterConf.Proton_mongodb.Hosts == nil {
					if p.Require {
						return fmt.Errorf("Unable to get Proton-mongodb Hosts configuration")
					}
					Backuplog.Infoln("Unable to get Proton-mongodb Hosts configuration,skip backup Proton-mongodb")
					skipbackup = true
					continue
				} else if len(clusterConf.Proton_mongodb.Hosts) == 0 || (len(clusterConf.Proton_mongodb.Hosts) > 1 && !opt.BackupMongoDB) {
					if p.Require {
						return fmt.Errorf("Proton-mongodb the number of copies is %d, and mongodb data cannot be backed up", len(clusterConf.Proton_mongodb.Hosts))
					}
					Backuplog.Infof("Proton-mongodb the number of copies is %d, and mongodb data cannot be backed up,skip backup Proton-mongodb", len(clusterConf.Proton_mongodb.Hosts))
					skipbackup = true
					continue
				} else if !slices.Contains(clusterConf.Proton_mongodb.Hosts, conf.HostName) {
					if p.Require {
						return fmt.Errorf("The current node not have proton-mongodb")
					}
					Backuplog.Infoln("The current node not have proton-mongodb,skip backup Proton-mongodb")
					skipbackup = true
					continue
				} else if clusterConf.Proton_mongodb.Admin_user == "" || clusterConf.Proton_mongodb.Admin_passwd == "" {
					return fmt.Errorf("Proton-mongodb username or password cannot be empty")
				}
				svc, err := k.CoreV1().Services(NamespaceResource).Get(context.Background(), ProtonMongoDBServiceName, metav1.GetOptions{})
				if err != nil {
					return err
				} else if svc == nil {
					return fmt.Errorf("unable to get mongodb service:" + ProtonMongoDBServiceName)
				}
				var ip = svc.Spec.ClusterIP
				if ip == "" {
					return errors.New("Unable to obtain the management terminal address of mongodb")
				}
				headers := map[string]string{
					"admin-key": base64.StdEncoding.EncodeToString([]byte(clusterConf.Proton_mongodb.Admin_user + ":" + clusterConf.Proton_mongodb.Admin_passwd)),
				}
				var httpclient = client.NewHttpClient(30)

				// get mongodb backup size and compare target dirctory free size
				// if bigger, then return error
				Backuplog.Infoln("check mongodb size and backup directory size")
				if code, resp, err := httpclient.Get(fmt.Sprintf("http://%s:30281/api/proton-mongodb-mgmt/v2/backup_size", ip), headers); err != nil {
					return fmt.Errorf("unable to get mongodb backup directory size, err: %w", err)
				} else if code != http.StatusOK {
					return fmt.Errorf("unable to get mongodb backup directory size, http status code: %d, response body: %v", code, resp)
				} else {
					freeSpace, err := GetFreeSpaceGB(conf.BackupDirectory)
					if err != nil {
						return fmt.Errorf("unable to get backup directory free size, error: %w", err)
					}
					// dbSize unit is KB
					dbSize := resp.(float64) / 1024 / 1024
					if dbSize >= freeSpace {
						return fmt.Errorf("unable to create mongodb backup, mongodb data size: %f bigger than backup directory free size: %f", dbSize, freeSpace)
					}
				}

				var mongodbRequest = MongoDBBackupRequest{}
				mongodbRequest.BackupDir = filepath.Join(workDirectory, p.TargetRelPath)
				if s, b, err := httpclient.Post(fmt.Sprintf("http://%s:30281/api/proton-mongodb-mgmt/v2/backups", ip), headers, mongodbRequest); err != nil {
					return fmt.Errorf("unable to create mongodb backup, error: %w", err)
				} else if s != http.StatusOK {
					return fmt.Errorf("unable to create mongodb backup, http status code: %d, response body: %v", s, b)
				} else {
					time.Sleep(BackupSleepInterval)
					Backuplog.Infof("create mongodb backup, http status code: %d, response body: %v \n", s, b)
					responseByte, err := jsoniter.Marshal(b)
					if err != nil {
						Backuplog.Infoln("unable to create mongodb backup,The returned data cannot be converted to byte")
						return err
					}
					info := MongoDBPackageInfo{}
					err = jsoniter.Unmarshal(responseByte, &info)
					if err != nil {
						Backuplog.Infoln("unable to create mongodb backup,The returned data cannot be converted to MongoDBPackageInfo")
						return err
					}
					for i := 1; i <= BackupCycleCount; i++ {
						if s, b, err := httpclient.Get(fmt.Sprintf("http://%s:30281/api/proton-mongodb-mgmt/v2/backups/%s", ip, info.Id), headers); err != nil {
							Backuplog.Errorf("unable to get mongodb backup message, error: %v, retry", err)
						} else if s != http.StatusOK {
							Backuplog.Errorf("unable to get mongodb backup message, http status code: %d, response body: %v, retry", s, b)
						} else {
							Backuplog.Infof("get mongodb backup message, http status code: %d, response body: %v \n", s, b)
							responseByte, err := jsoniter.Marshal(b)
							if err != nil {
								Backuplog.Infoln("unable to get mongodb backup message,The returned data cannot be converted to byte")
								return err
							}
							list := []MongoDBPackageInfo{}
							err = jsoniter.Unmarshal(responseByte, &list)
							if err != nil {
								Backuplog.Infoln("unable to get mongodb backup message,The returned data cannot be converted to MongoDBBackupList")
								return err
							}
							if len(list) > 1 || len(list) == 0 {
								return fmt.Errorf("unable to get mongodb backup message, count: %d", len(list))
							}
							if list[0].Status == "success" {
								goto MongodbSuccess
							}
							if list[0].Status == "failed" {
								return fmt.Errorf("backup mongodb failed, http status code: %d, response body: %v", s, b)
							}
						}
						time.Sleep(BackupSleepInterval)
					}
					return fmt.Errorf("unable to get mongodb backup message：checked %d count", BackupCycleCount)
				}
			MongodbSuccess:
				Backuplog.Infoln("mongodb backup success")
			}
		} else if info.Name == "proton-etcd" {

			for _, p := range info.PathList {
				if clusterConf == nil {
					return fmt.Errorf("Unable to get cluster configuration file")
				} else if clusterConf.Proton_etcd == nil {
					if p.Require {
						return fmt.Errorf("Unable to get Proton-etcd configuration")
					}
					Backuplog.Infoln("Unable to get Proton-etcd configuration,skip")
					skipbackup = true
					continue
				} else if len(clusterConf.Proton_etcd.Hosts) == 0 {
					if p.Require {
						return fmt.Errorf("Proton-etcd the number of copies is %d, and etcd data cannot be backed up", len(clusterConf.Proton_etcd.Hosts))
					}
					Backuplog.Infof("Proton-etcd the number of copies is %d, and etcd data cannot be backed up", len(clusterConf.Proton_etcd.Hosts))
					skipbackup = true
					continue
				} else if !slices.Contains(clusterConf.Proton_etcd.Hosts, conf.HostName) {
					if p.Require {
						return fmt.Errorf("The current node not have proton-etcd")
					}
					Backuplog.Infoln("The current node not have proton-etcd,skip backup Proton-etcd")
					skipbackup = true
					continue
				}
				secret, err := k.CoreV1().Secrets(NamespaceResource).Get(context.Background(), ProtonEtcdSecretName, metav1.GetOptions{})
				if err != nil {
					return err
				} else if secret == nil {
					return fmt.Errorf("unable to get proton-etcd serret:" + ProtonEtcdSecretName)
				}
				var etcdSnapshotPath = filepath.Join(workDirectory, p.TargetRelPath)
				err = os.MkdirAll(etcdSnapshotPath, os.ModePerm)
				if err != nil {
					return err
				}
				var caCrt = secret.Data["ca.crt"]
				var peerCrt = secret.Data["peer.crt"]
				var peerkey = secret.Data["peer.key"]
				// 创建证书文件
				caFile, err := os.Create(filepath.Join(etcdSnapshotPath, "ca.crt"))
				defer Check(caFile.Close)
				if err != nil {
					return err
				}
				_, err = caFile.Write(caCrt)
				if err != nil {
					return err
				}

				peerCrtFile, err := os.Create(filepath.Join(etcdSnapshotPath, "peer.crt"))
				if err != nil {
					return err
				}
				defer Check(peerCrtFile.Close)

				_, err = peerCrtFile.Write(peerCrt)
				if err != nil {
					return err
				}
				peerKeyFile, err := os.Create(filepath.Join(etcdSnapshotPath, "peer.key"))
				defer Check(peerKeyFile.Close)

				if err != nil {
					return err
				}
				_, err = peerKeyFile.Write(peerkey)
				if err != nil {
					return err
				}
				proton_etcd_pods, err := k.CoreV1().Pods(NamespaceResource).List(context.Background(), metav1.ListOptions{
					FieldSelector: "spec.nodeName=" + conf.HostName,
					LabelSelector: "app.kubernetes.io/instance=proton-etcd",
				})
				if err != nil {
					return err
				}
				if proton_etcd_pods == nil && len(proton_etcd_pods.Items) == 0 {
					return fmt.Errorf("unable to get proton-etcd pod from node: %s", conf.HostName)
				}
				for _, etcd_pod := range proton_etcd_pods.Items {
					var proton_etcd_yaml_Path = filepath.Join(etcdSnapshotPath, etcd_pod.Name+".yaml")
					jsonBytes, err := yaml.Marshal(etcd_pod)
					if err != nil {
						return err
					}
					err = os.WriteFile(proton_etcd_yaml_Path, jsonBytes, 0644)
					if err != nil {
						return err
					}
				}
				_, err = k.CoreV1().Services(NamespaceResource).Get(context.Background(), ProtonEtcdNodeServiceName, metav1.GetOptions{})
				if err != nil {
					if !k8serrors.IsNotFound(err) {
						return err
					}
					var labels = labels.Set{
						"app.kubernetes.io/instance": "proton-etcd",
						"app.kubernetes.io/name":     "proton-etcd",
					}
					var ProtonEtcdNodePortService = core_v1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Name:      ProtonEtcdNodeServiceName,
							Namespace: NamespaceResource,
							Labels:    labels,
						},
						Spec: core_v1.ServiceSpec{
							Ports: []core_v1.ServicePort{
								{
									Name:       "client",
									Port:       ProtonEtcdPodPort,
									TargetPort: intstr.FromString("client"),
									NodePort:   ProtonEtcdNodePort,
								},
							},
							Selector: labels,
							Type:     core_v1.ServiceTypeNodePort,
						},
					}
					if _, err := k.CoreV1().Services(NamespaceResource).Create(context.Background(), &ProtonEtcdNodePortService, metav1.CreateOptions{}); err != nil {
						return err
					}
				}
				ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
				defer cancel()
				var etcdUrl = fmt.Sprintf("%s://%s:%d", ProtonEtcdProtocol, "127.0.0.1", ProtonEtcdNodePort)
				if cfg, err := etcd.EtcdClientConfig(etcdUrl, filepath.Join(etcdSnapshotPath, "peer.crt"), filepath.Join(etcdSnapshotPath, "peer.key"), filepath.Join(etcdSnapshotPath, "ca.crt")); err != nil {
					return err
				} else if err := snapshot.Save(ctx, zap.NewExample(), cfg, filepath.Join(etcdSnapshotPath, "proton-etcd-snapshot.db")); err != nil {
					return err
				}

				defer func() {
					// delete etcd nodeport service
					if err := k.CoreV1().Services(NamespaceResource).Delete(context.Background(), ProtonEtcdNodeServiceName, metav1.DeleteOptions{}); err != nil {
						Backuplog.Infoln("delete etcd nodeport service error:", err)
					} else {
						Backuplog.Infoln("delete etcd nodeport service success")
					}
				}()
			}
		} else if info.Name == "kubernetes-etcd" {
			for _, p := range info.PathList {
				if clusterConf == nil {
					return fmt.Errorf("Unable to get cluster configuration file")
				} else if clusterConf.Cs == nil {
					if p.Require {
						return fmt.Errorf("Unable to get cs configuration")
					}
					Backuplog.Infoln("Unable to get cs configuration,skip create kubernetes-etcd Snapshop")
					skipbackup = true
					continue
				} else if !slices.Contains(clusterConf.Cs.Master, conf.HostName) {
					if p.Require {
						return fmt.Errorf("The current node is not the master")
					}
					Backuplog.Infoln("The current node is not the master,skip create kubernetes-etcd Snapshop")
					skipbackup = true
					continue
				}
				Backuplog.Infoln("create kubernetes-etcd Snapsho")
				var etcdSnapshotPath = filepath.Join(workDirectory, p.TargetRelPath)
				err = os.MkdirAll(etcdSnapshotPath, os.ModePerm)
				if err != nil {
					return err
				}
				ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
				defer cancel()
				if cfg, err := etcd.EtcdClientConfig(etcd.EtcdEndpoint, etcd.APIServerEtcdClientCertPath, etcd.APIServerEtcdClientKeyPath, etcd.EtcdCACertPath); err != nil {
					return err
				} else if err := snapshot.Save(ctx, zap.NewExample(), cfg, filepath.Join(etcdSnapshotPath, etcd.EtcdSnapshotFileName)); err != nil {
					return err
				}

				var errList []error
				if err := file.Copy(etcd.APIServerEtcdClientCertPath, filepath.Join(etcdSnapshotPath, etcd.APIServerEtcdClientCertName), Backuplog, nil); err != nil {
					errList = append(errList, fmt.Errorf("%s: %w", info.Name, err))
				}
				if err := file.Copy(etcd.APIServerEtcdClientKeyPath, filepath.Join(etcdSnapshotPath, etcd.APIServerEtcdClientKeyName), Backuplog, nil); err != nil {
					errList = append(errList, fmt.Errorf("%s: %w", info.Name, err))
				}
				if err := file.Copy(etcd.EtcdCACertPath, filepath.Join(etcdSnapshotPath, "ca.crt"), Backuplog, nil); err != nil {
					errList = append(errList, fmt.Errorf("%s: %w", info.Name, err))
				}

				if err := file.Copy(KubernetesEtcdPath, filepath.Join(etcdSnapshotPath, "etcd.yaml"), Backuplog, nil); err != nil {
					errList = append(errList, fmt.Errorf("%s: %w", info.Name, err))
				}

				if len(errList) > 0 {
					return utilerrors.NewAggregate(errList)
				}
			}
		} else {
			for _, p := range info.PathList {
				if !p.Require {
					if exist, err := file.PathExists(p.OsPath); err != nil {
						return err
					} else if !exist {
						continue
					}
				}
				if p.OsPath == "" {
					continue
				}
				err := file.Copy(p.OsPath, filepath.Join(workDirectory, p.TargetRelPath), Backuplog, p.ExcludeathList)
				if err != nil {
					return err
				}
			}
		}
		if !skipbackup {
			fo.Resource = append(fo.Resource, info.Name)
		}
	}

	Backuplog.Infoln("压缩备份目录")
	if err := tar.CreateTarball(filepath.Join(conf.BackupDirectory, opt.Id+".tar.gz"), conf.BackupDirectory, opt.Id); err != nil {
		return err
	} else {
		fi, err := os.Stat(filepath.Join(conf.BackupDirectory, opt.Id+".tar.gz"))
		if err != nil {
			return err
		}
		fo.UseSpace = fi.Size()
		err = os.RemoveAll(workDirectory)
		if err != nil {
			return err
		}
	}

	return nil
}

// 更新备份的存储目录
func UpgradeBackUpPath(backupDirectory string) error {
	conf, err := GetBackupConf()
	// 当前备份配置目录没有备份记录，但是修改的目标备份数据目录下存在备份的配置文件目录时，复制到当前备份配置目录中，重新获取备份配置信息
	if len(conf.List) == 0 {
		//如果没有备份的配置文件，检测备份存储的数据目录是否存在备份的配置，存在的话把目录复制到主机的/opt/proton-backup备份配置的目录中
		proton_backup_backup_dir := filepath.Join(backupDirectory, BackupDir)
		exist, err := file.PathExists(proton_backup_backup_dir)
		if err != nil {
			Backuplog.Errorln("检测proton_backup的备份配置目录是否存在异常：", proton_backup_backup_dir, err)
			return err
		}
		if exist {
			//当系统盘故障，可使用该备份目录还原proton-cli的备份和还原的记录以及日志
			err = file.Copy(proton_backup_backup_dir, BackupDir, Backuplog, nil)
			Backuplog.Infoln(conf.BackupDirectory + "没有备份记录，目标备份数据目录:" + proton_backup_backup_dir + "下存在备份的配置文件目录时，复制到:" + conf.BackupDirectory)
			if err != nil {
				return err
			}
			conf, err = GetBackupConf()
			if err != nil {
				return err
			}
		}
	}

	if err != nil {
		return err
	}
	if conf != nil && conf.BackupDirectory != "" {
		exist, err := file.PathExists(backupDirectory)
		if err != nil {
			return err
		}
		if !exist {
			Backuplog.Infoln("path does not exist:" + backupDirectory)
		} else {

			conf.BackupDirectory = backupDirectory
			jsonBytes, err := json.Marshal(conf)
			if err != nil {
				return err
			}
			err = os.WriteFile(filepath.Join(BackupDir, BackupConfName), jsonBytes, 0644)
			if err != nil {
				return err
			}
			Backuplog.Infoln("保存成功")
		}
	} else {
		Backuplog.Infoln("no backup config")
	}
	return nil
}

// 清理过期的备份资源包。
// 成功备份的清理规则：清理超过有效时间并且不在最近三次成功的备份记录和资源
// 失败备份的清理规则：超过有效时间删除备份记录和资源
func CleanupExpiredBackUp() error {
	// clean previous release mariadb/mongodb history backup job
	// get all mariadb/mongodb backups and delete them
	err := DeleteMariadbHistoryBackupJob()
	if err != nil {
		Backuplog.Infof("unable to delete mariadb history backup job, error: %v", err)
	}

	err = DeleteMongodbHistoryBackupJob()
	if err != nil {
		Backuplog.Infof("unable to delete mongodb history backup job, error: %v", err)
	}

	conf, err := GetBackupConf()
	if err != nil {
		return err
	}
	if conf != nil && len(conf.List) > 0 {
		var validSet = []BackupInfo{} //有效的备份记录
		//最近三次成功的备份记录
		var lasterSuccessSet = arrayutil.Filter(conf.List, func(s1 BackupInfo) bool { return !s1.Status })

		// 用id排序,数值小的排在前
		sort.SliceStable(lasterSuccessSet, func(i, j int) bool {
			return conf.List[i].CreateTime > conf.List[j].CreateTime
		})
		if len(lasterSuccessSet) > 3 {
			lasterSuccessSet = lasterSuccessSet[0:3] //最近三次成功的备份记录
		}
		for _, info := range conf.List {
			var ttl float64
			var expirationDate = time.Unix(info.CreateTime, 0).Add(time.Hour * 24 * time.Duration(info.Ttl))
			if expirationDate.Before(time.Now()) {
				ttl = 0
			} else {
				ttl = time.Until(expirationDate).Hours()
			}
			ttlHours := ttl
			if ttlHours > 0 {
				validSet = append(validSet, info)
			} else {
				if info.Status {
					var exitsArray = arrayutil.Filter(lasterSuccessSet, func(s1 BackupInfo) bool { return s1.Id != info.Id })
					if len(exitsArray) > 0 {
						validSet = append(validSet, info)
					} else {
						os.Remove(info.Path)
						os.Remove(info.LogPath)
						if err != nil {
							return err
						}
					}
				} else {
					os.Remove(info.Path)
					os.Remove(info.LogPath)
					if err != nil {
						return err
					}
				}
			}
		}
		conf.List = validSet
		jsonBytes, err := json.Marshal(conf)
		if err != nil {
			return err
		}
		err = os.WriteFile(filepath.Join(BackupDir, BackupConfName), jsonBytes, 0644)
		if err != nil {
			return err
		}
		Backuplog.Infoln("过期备份资源清理成功")
	}
	return nil
}

// 获取指定目录可用空间的大小-GB单位
func GetFreeSpaceGB(path string) (float64, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return 0, err
	}
	size := float64(stat.Bavail*uint64(stat.Bsize)) / 1024 / 1024 / 1024 // GB
	return size, nil
}

func DeleteMariadbHistoryBackupJob() error {
	var resourceNamespace = configuration.GetProtonResourceNSFromFile()
	var clusterConf *configuration.ClusterConfig
	_, k := client.NewK8sClient()
	if k == nil {
		return client.ErrKubernetesClientSetNil
	}
	clusterConf, err := configuration.LoadFromKubernetes(context.Background(), k)
	if err != nil {
		return err
	}
	if clusterConf == nil {
		return fmt.Errorf("Unable to get cluster configuration file")
	}
	svc, err := k.CoreV1().Services(resourceNamespace).Get(context.Background(), ProtonMariaDBServiceName, metav1.GetOptions{})
	if err != nil {
		return err
	} else if svc == nil {
		return fmt.Errorf("unable to get mariadb service:" + ProtonMariaDBServiceName)
	}
	var ip = svc.Spec.ClusterIP
	if ip == "" {
		return errors.New("Unable to obtain the management terminal address of mariadb")
	}
	headers := map[string]string{
		"admin-key": base64.StdEncoding.EncodeToString([]byte(clusterConf.Proton_mariadb.Admin_user + ":" + clusterConf.Proton_mariadb.Admin_passwd)),
	}
	var httpclient = client.NewHttpClient(30)
	if s, b, err := httpclient.Get(fmt.Sprintf("http://%s:8888/api/proton-rds-mgmt/v2/backups", ip), headers); err != nil {
		return fmt.Errorf("unable to get mariadb backup, error: %w", err)
	} else if s != http.StatusOK {
		return fmt.Errorf("unable to get mariadb backup, http status code: %d, response body: %v", s, b)
	} else {
		Backuplog.Infof("get mariadb backup message, http status code: %d, response body: %v \n", s, b)
		responseByte, err := jsoniter.Marshal(b)
		if err != nil {
			Backuplog.Infoln("unable to get mariadb backup message,The returned data cannot be converted to byte")
			return err
		}
		jobs := []MariaDBPackageInfo{}
		err = jsoniter.Unmarshal(responseByte, &jobs)
		if err != nil {
			Backuplog.Infoln("unable to get mariadb backup message,The returned data cannot be converted to MariaDBBackupList")
			return err
		}
		for _, job := range jobs {
			for i := 1; i <= 5; i++ {
				Backuplog.Infof("delete mariadb history backup job: %s", job.Id)
				if _, err = httpclient.Delete(fmt.Sprintf("http://%s:8888/api/proton-rds-mgmt/v2/backups/%s", ip, job.Id), headers); err != nil {
					return fmt.Errorf("unable to delete mariadb backup job, error: %w", err)
				}
				time.Sleep(2 * time.Second)
				if s, b, err := httpclient.Get(fmt.Sprintf("http://%s:8888/api/proton-rds-mgmt/v2/backups?id=%s", ip, job.Id), headers); err != nil {
					return fmt.Errorf("unable to get mariadb backup, error: %w", err)
				} else if s != http.StatusOK {
					return fmt.Errorf("unable to get mariadb backup, http status code: %d, response body: %v", s, b)
				} else {
					Backuplog.Infof("get mariadb backup message, http status code: %d, response body: %v , continue waiting\n", s, b)
					time.Sleep(2 * time.Second)
				}
			}
		}
	}
	return nil
}

func DeleteMongodbHistoryBackupJob() error {
	var resourceNamespace = configuration.GetProtonResourceNSFromFile()
	var clusterConf *configuration.ClusterConfig
	_, k := client.NewK8sClient()
	if k == nil {
		return client.ErrKubernetesClientSetNil
	}
	clusterConf, err := configuration.LoadFromKubernetes(context.Background(), k)
	if err != nil {
		return err
	}
	if clusterConf == nil {
		return fmt.Errorf("Unable to get cluster configuration file")
	}
	svc, err := k.CoreV1().Services(resourceNamespace).Get(context.Background(), ProtonMongoDBServiceName, metav1.GetOptions{})
	if err != nil {
		return err
	} else if svc == nil {
		return fmt.Errorf("unable to get mongodb service:" + ProtonMongoDBServiceName)
	}
	var ip = svc.Spec.ClusterIP
	if ip == "" {
		return errors.New("Unable to obtain the management terminal address of mongodb")
	}
	headers := map[string]string{
		"admin-key": base64.StdEncoding.EncodeToString([]byte(clusterConf.Proton_mongodb.Admin_user + ":" + clusterConf.Proton_mongodb.Admin_passwd)),
	}
	var httpclient = client.NewHttpClient(30)
	if s, b, err := httpclient.Get(fmt.Sprintf("http://%s:30281/api/proton-mongodb-mgmt/v2/backups", ip), headers); err != nil {
		return fmt.Errorf("unable to get mongodb backup, error: %w", err)
	} else if s != http.StatusOK {
		return fmt.Errorf("unable to get mongodb backup, http status code: %d, response body: %v", s, b)
	} else {
		Backuplog.Infof("get mongodb backup message, http status code: %d, response body: %v \n", s, b)
		responseByte, err := jsoniter.Marshal(b)
		if err != nil {
			Backuplog.Infoln("unable to get mongodb backup message,The returned data cannot be converted to byte")
			return err
		}
		jobs := []MongoDBPackageInfo{}
		err = jsoniter.Unmarshal(responseByte, &jobs)
		if err != nil {
			Backuplog.Infoln("unable to get mongodb backup message,The returned data cannot be converted to MongoDBBackupList")
			return err
		}
		for _, job := range jobs {
			for i := 1; i <= 5; i++ {
				Backuplog.Infof("delete mongodb backup job: %s", job.Id)
				if _, err = httpclient.Delete(fmt.Sprintf("http://%s:30281/api/proton-mongodb-mgmt/v2/backups/%s", ip, job.Id), headers); err != nil {
					return fmt.Errorf("unable to delete mongodb backup job, error: %w", err)
				}
				time.Sleep(2 * time.Second)
				if s, b, err := httpclient.Get(fmt.Sprintf("http://%s:30281/api/proton-mongodb-mgmt/v2/backups?id=%s", ip, job.Id), headers); err != nil {
					return fmt.Errorf("unable to get mongodb backup, error: %w", err)
				} else if s != http.StatusOK {
					return fmt.Errorf("unable to get mongodb backup, http status code: %d, response body: %v", s, b)
				} else {
					Backuplog.Infof("get mongodb backup message, http status code: %d, response body: %v , continue waiting\n", s, b)
					time.Sleep(2 * time.Second)
				}
			}
		}
	}
	return nil
}
