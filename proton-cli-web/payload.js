const payload = {
  apiVersion: "v1",
  cms: {},
  nvidia_device_plugin: {},
  cr: {
    local: {
      ha_ports: {
        chartmuseum: 15001,
        cr_manager: 15002,
        registry: 15000,
        rpm: 15003,
      },
      hosts: ["node-66-194"],
      ports: {
        chartmuseum: 5001,
        cr_manager: 5002,
        registry: 5000,
        rpm: 5003,
      },
      storage: "/sysvol/proton_data/cr_data",
    },
  },
  cs: {
    ipFamilies: ["IPv4"],
    cs_controller_dir: "./service-package",
    docker_data_dir: "/sysvol/proton_data/cs_docker_data",
    etcd_data_dir: "/sysvol/proton_data/cs_etcd_data",
    ha_port: 16643,
    host_network: {
      bip: "172.33.0.1/16",
      pod_network_cidr: "192.169.0.0/16",
      service_cidr: "10.96.0.0/12",
    },
    master: ["node-66-194"],
    provisioner: "local",
  },
  kafka: {
    data_path: "/sysvol/kafka/kafka_data",
    env: null,
    hosts: ["node-66-194"],
    resources: {},
  },
  nodes: [
    {
      ip4: "192.0.2.1",
      name: "node-66-194",
    },
  ],
  opensearch: {
    config: {
      hanlpRemoteextDict:
        "http://ecoconfig-private.anyshare:32128/api/ecoconfig/v1/word-list/remote_ext_dict",
      jvmOptions: "-Xmx8g -Xms8g",
    },
    data_path: "/anyshare/opensearch",
    hosts: ["node-66-194"],
    mode: "master",
    settings: {
      "cluster.routing.allocation.disk.watermark.flood_stage": "70%",
      "cluster.routing.allocation.disk.watermark.high": "65%",
      "cluster.routing.allocation.disk.watermark.low": "60%",
    },
  },
  proton_etcd: {
    data_path: "/sysvol/proton-etcd/proton-etcd_data",
    hosts: ["node-66-194"],
  },
  proton_mariadb: {
    admin_passwd: "",
    admin_user: "root",
    config: {
      innodb_buffer_pool_size: "4G",
      resource_limits_memory: "10G",
      resource_requests_memory: "10G",
    },
    data_path: "/sysvol/mariadb",
    hosts: ["node-66-194"],
  },
  proton_mongodb: {
    admin_passwd: "",
    admin_user: "root",
    data_path: "/sysvol/mongodb/mongodb_data",
    hosts: ["node-66-194"],
  },
  proton_mq_nsq: {
    data_path: "/sysvol/mq-nsq/mq-nsq_data",
    hosts: ["node-66-194"],
  },
  proton_policy_engine: {
    data_path: "/sysvol/policy-engine/policy-engine_data",
    hosts: ["node-66-194"],
  },
  proton_redis: {
    admin_passwd: "",
    admin_user: "root",
    data_path: "/sysvol/redis/redis_data",
    hosts: ["node-66-194"],
  },
  zookeeper: {
    data_path: "/sysvol/zookeeper/zookeeper_data",
    env: null,
    hosts: ["node-66-194"],
    resources: {},
  },
  prometheus: {
    hosts: ["node-66-194"],
    data_path: "/sysvol/prometheus",
  },
  grafana: {
    hosts: ["node-66-194"],
    data_path: "/sysvol/grafana",
  },
};

module.exports = {
  payload,
};
