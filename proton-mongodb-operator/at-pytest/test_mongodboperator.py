import re
import pytest
import allure
import common
import json
import time
import requests

SSH_CLIENT = None
MONGO_URL="mongodb://USERNAME:PASSWORD@mongodboperator-mongodb-0.mongodboperator-mongodb.default.svc.cluster.local:28000,mongodbopeartor-mongodb-1.mongodboperator-mongodb.default.svc.cluster.local:28000,mongodboperator-mongodb-2.mongodboperator-mongodb.default.svc.cluster.local:28000/?replicaSet=rs0"

@allure.feature("MongoOperator")
class Test_MongoOperator:
    def wait_cr_ready(self):
        for i in range(60):
            stdout, stderr = SSH_CLIENT.exec_cmd(
                cmd_str="kubectl get pod | grep mongodb | grep 0/1",
                timeout=30
            )

            assert not re.search(r'.*CrashLoopBackOff.*', stdout)
            if stdout == "":
                break
            time.sleep(5)
        stdout, stderr = SSH_CLIENT.exec_cmd(
            cmd_str="kubectl get pod | grep mongodb | grep 0/1",
            timeout=30
            )
        assert stdout == ""

    def generate_cr_json(self, host, n):
        volume = []
        for i in range(n):
            volume.append(
                {
                    "host": host,
                    "path": f"/data/pv-{i}"
                }
            )
        data = {
            "apiVersion": "mongodb.proton.aishu.cn/v1",
            "kind": "MongodbOperator",
            "metadata": {
                "name": "mongodboperator",
                "namespace": "default",
            },
            "spec":{
                "secretname": "mongo-secret",
                "logrotate": {
                    "image": "acr.aishu.cn/proton/logrotate:1.0.0",
                    "imagePullPolicy": "IfNotPresent",
                    "logsize": "2M",
                    "schedule": "*/2 * * * *",
                    "logcount": 5,
                },
                "exporter": {
                    "image": "acr.aishu.cn/proton/mongodb-exporter:2.1.0-develop",
                    "imagePullPolicy": "IfNotPresent",
                },
                "mgmt": {
                    "image": "acr.aishu.cn/proton/proton-mongodb-mgmt:2.1.0-develop",
                    "imagePullPolicy": "IfNotPresent",
                    "useEncryption": False,
                    "logLevel": "info",
                    "service": {
                        "type": "NodePort",
                        "enableDualStack": False,
                        "port": 30281,
                    },
                },
                "mongodb": {
                    "replicas": n,
                    "replset": {
                        "name": "rs0",
                    },
                    "image": "acr.aishu.cn/proton/mongodb:2.0.0-develop",
                    "imagePullPolicy": "IfNotPresent",
                    "conf": {
                        "wiredTigerCacheSizeGB": 4,
                        "tls":{
                            "enabled": False
                        }
                    },
                    "service": {
                        "type": "NodePort",
                        "enableDualStack": False,
                        "port": 30280,
                    },
                    "debug": "0",
                    "storage": {
                        "capacity": "10Gi",
                        "storageClassName": "",
                        "volume": volume,
                    },
                    "resources": {},
                },
            }
        }
        return data

    @allure.title("准备环境")
    def setup_class(self):
        pass

    @allure.title("清理环境")
    def teardown_class(self):
        SSH_CLIENT.exec_cmd(
            cmd_str="kubectl delete crd/mongodboperators.mongodb.proton.aishu.cn",
            timeout=120
        )
        SSH_CLIENT.exec_cmd(
            cmd_str="helm delete --purge mongodb-operator",
            timeout=120
        )
        SSH_CLIENT.exec_cmd(
            cmd_str="kubectl delete secret/mongo-secret",
            timeout=30
        )
        SSH_CLIENT.exec_cmd(
            cmd_str="rm -rf /data/pv-{0,1,2}",
            timeout=30
        )

    @allure.title("安装MongodbOperator成功")
    def test_0(self,host,user,password):
        global SSH_CLIENT
        SSH_CLIENT = common.SSHClient(host, 22, user, password)
        stdout, stderr = SSH_CLIENT.exec_cmd(
            cmd_str=f"helm install proton/mongodb-operator --version=1.0.0-master --name mongodb-operator",
            timeout=120
        )
        time.sleep(30)
        assert re.search(r'.*DEPLOYED*', stdout)

        stdout, stderr = SSH_CLIENT.exec_cmd(
            cmd_str="kubectl get pod -n proton-mongodb-operator-system | grep mongodb",
            timeout=30
        )

        assert re.search(r'.*2/2.*', stdout)

    @allure.title("安装1副本cr成功")
    def test_1(self, host):
        stdout, stderr = SSH_CLIENT.exec_cmd(
            cmd_str="hostname",
            timeout=30
        )
        hostname = stdout.replace("\n", "")

        SSH_CLIENT.exec_cmd(
            cmd_str="mkdir -p /data/pv-{0,1,2}",
            timeout=30
        )
        SSH_CLIENT.exec_cmd(
            cmd_str="kubectl create secret generic mongo-secret --from-literal=username=USERNAME --from-literal=password=UEFTU1dPUkQ=",
            timeout=30
        )

        data = self.generate_cr_json(hostname, 1)
        with open('cr.json', 'w') as f:
            json.dump(data, f)
        SSH_CLIENT.scp('cr.json', '/tmp/cr.json')

        stdout, stderr = SSH_CLIENT.exec_cmd(
            cmd_str="kubectl create -f /tmp/cr.json",
            timeout=30
        )
        assert re.search(r'.*created.*', stdout)

        time.sleep(30)
        self.wait_cr_ready()

        stdout, stderr = SSH_CLIENT.exec_cmd(
            cmd_str=f"kubectl exec -it pod/mongodboperator-mongodb-0 -- mongo {MONGO_URL} --quiet --eval \"version()\"",
            timeout=30
        )
        assert re.search(r'.*4.2.23-23.*', stdout)

    @allure.title("1副本扩容3副本成功/3副本缩容1副本成功")
    def test_2(self, host):
        stdout, stderr = SSH_CLIENT.exec_cmd(
            cmd_str=f"kubectl exec -i pod/mongodboperator-mongodb-0 -- mongo {MONGO_URL} --quiet --eval \"db.getSiblingDB('test').tb_test1.insertOne({{'name':'test'}})\"",
            timeout=30
        )
        assert not stderr

        stdout, stderr = SSH_CLIENT.exec_cmd(
            cmd_str=f"kubectl exec -i pod/mongodboperator-mongodb-0 -- mongo {MONGO_URL} --quiet --eval \"db.getSiblingDB('test').tb_test1.findOne()\"",
            timeout=30
        )
        assert not stderr

        stdout, stderr = SSH_CLIENT.exec_cmd(
            cmd_str="hostname",
            timeout=30
        )
        hostname = stdout.replace("\n", "")

        data = self.generate_cr_json(hostname, 3)
        with open('cr.json', 'w') as f:
            json.dump(data, f)
        SSH_CLIENT.scp('cr.json', '/tmp/cr.json')

        stdout, stderr = SSH_CLIENT.exec_cmd(
            cmd_str="kubectl apply -f /tmp/cr.json",
            timeout=30
        )
        assert re.search(r'.*configured.*', stdout)

        time.sleep(60)
        self.wait_cr_ready()
        stdout, stderr = SSH_CLIENT.exec_cmd(
            cmd_str=f"kubectl exec -i pod/mongodboperator-mongodb-0 -- mongo {MONGO_URL} --quiet --eval \"db.getSiblingDB('test').tb_test1.findOne()\"",
            timeout=30
        )
        assert not stderr

        data = self.generate_cr_json(hostname, 1)
        with open('cr.json', 'w') as f:
            json.dump(data, f)
        SSH_CLIENT.scp('cr.json', '/tmp/cr.json')
        stdout, stderr = SSH_CLIENT.exec_cmd(
            cmd_str="kubectl apply -f /tmp/cr.json",
            timeout=30
        )
        assert re.search(r'.*configured.*', stdout)

        time.sleep(60)
        self.wait_cr_ready()
        stdout, stderr = SSH_CLIENT.exec_cmd(
            cmd_str=f"kubectl exec -i pod/mongodboperator-mongodb-0 -- mongo {MONGO_URL} --quiet --eval \"db.getSiblingDB('test').tb_test1.findOne()\"",
            timeout=30
        )
        assert not stderr

        data = self.generate_cr_json(hostname, 3)
        with open('cr.json', 'w') as f:
            json.dump(data, f)
        SSH_CLIENT.scp('cr.json', '/tmp/cr.json')
        stdout, stderr = SSH_CLIENT.exec_cmd(
            cmd_str="kubectl apply -f /tmp/cr.json",
            timeout=30
        )
        assert re.search(r'.*configured.*', stdout)

        time.sleep(60)
        self.wait_cr_ready()
        stdout, stderr = SSH_CLIENT.exec_cmd(
            cmd_str=f"kubectl exec -i pod/mongodboperator-mongodb-0 -- mongo {MONGO_URL} --quiet --eval \"db.getSiblingDB('test').tb_test1.findOne()\"",
            timeout=30
        )
        assert not stderr

    @allure.title("1副本停机服务可用, 可以自动恢复")
    def test_3(self, host):
        stdout, stderr = SSH_CLIENT.exec_cmd(
            cmd_str="kubectl delete pod/mongodboperator-mongodb-2",
            timeout=30
        )

        stdout, stderr = SSH_CLIENT.exec_cmd(
            cmd_str=f"kubectl exec -i pod/mongodboperator-mongodb-0 -- mongo {MONGO_URL} --quiet --eval \"db.getSiblingDB('test').tb_test1.findOne()\"",
            timeout=30
        )
        assert not stderr
        self.wait_cr_ready()


    @allure.title("3副本停机,自动恢复后服务可用")
    def test_4(self, host):
        stdout, stderr = SSH_CLIENT.exec_cmd(
            cmd_str="kubectl delete pod/mongodboperator-mongodb-0 pod/mongodboperator-mongodb-1 pod/mongodboperator-mongodb-2",
            timeout=60
        )
        time.sleep(60)
        self.wait_cr_ready()
        stdout, stderr = SSH_CLIENT.exec_cmd(
            cmd_str=f"kubectl exec -i pod/mongodboperator-mongodb-0 -- mongo {MONGO_URL} --quiet --eval \"db.getSiblingDB('test').tb_test1.findOne()\"",
            timeout=30
        )
        assert not stderr

    @allure.title("数据库管理:创建/删除db")
    def test_5(self, host):
        SSH_CLIENT.exec_cmd(
            cmd_str="firewall-cmd --add-port=30281/tcp",
            timeout=30
        )

        resp = requests.put(
            url=f'http://{host}:30281/api/proton-mongodb-mgmt/v2/dbs/at_db',
            headers={
                "Content-Type": "application/json",
                "admin-key":"",
            },
            data=json.dumps(
                {
                "collection_name":"as_collection_1"
                }
            )
        )
        assert resp.status_code == 400

        resp = requests.put(
            url=f'http://{host}:30281/api/proton-mongodb-mgmt/v2/dbs/at_db',
            headers={
                "Content-Type": "application/json",
                "admin-key":"VVNFUk5BTUU6UEFTU1dPUkQ=",
            },
            data=json.dumps(
                {
                "collection_name":"as_collection_1"
                }
            )
        )
        assert resp.status_code == 201

        resp = requests.put(
            url=f'http://{host}:30281/api/proton-mongodb-mgmt/v2/dbs/at_db',
            headers={
                "Content-Type": "application/json",
                "admin-key":"VVNFUk5BTUU6UEFTU1dPUkQ=",
            },
            data=json.dumps(
                {
                "collection_name":"as_collection_1"
                }
            )
        )
        assert resp.status_code == 403

        resp = requests.get(
            url=f'http://{host}:30281/api/proton-mongodb-mgmt/v2/dbs',
            headers={
                "Content-Type": "application/json",
                "admin-key":"VVNFUk5BTUU6UEFTU1dPUkQ=",
            },
        )
        assert resp.status_code == 200
        dbs = [v['db_name'] for v in resp.json()]
        assert 'at_db' in dbs

        resp = requests.delete(
            url=f'http://{host}:30281/api/proton-mongodb-mgmt/v2/dbs/at_db',
            headers={
                "Content-Type": "application/json",
                "admin-key":"VVNFUk5BTUU6UEFTU1dPUkQ=",
            },
        )
        assert resp.status_code == 204

        resp = requests.get(
            url=f'http://{host}:30281/api/proton-mongodb-mgmt/v2/dbs',
            headers={
                "Content-Type": "application/json",
                "admin-key":"VVNFUk5BTUU6UEFTU1dPUkQ=",
            },
        )
        dbs = [v['db_name'] for v in resp.json()]
        assert 'at_db' not in dbs

    @allure.title("数据库账户管理:创建/删除db用户")
    def test_6(self, host):
        resp = requests.put(
            url=f'http://{host}:30281/api/proton-mongodb-mgmt/v2/dbs/at_db',
            headers={
                "Content-Type": "application/json",
                "admin-key":"VVNFUk5BTUU6UEFTU1dPUkQ=",
            },
            data=json.dumps(
                {
                "collection_name":"as_collection_1"
                }
            )
        )
        assert resp.status_code == 201

        resp = requests.put(
            url=f'http://{host}:30281/api/proton-mongodb-mgmt/v2/users/at_db/as',
            headers={
                "Content-Type": "application/json",
                "admin-key":"VVNFUk5BTUU6UEFTU1dPUkQ=",
            },
            data=json.dumps(
                {
                "password":"UEFTU1dPUkQ=" # cSpell:ignore UEFTU1dPUkQ
                }
            )
        )
        assert resp.status_code == 201

        resp = requests.get(
            url=f'http://{host}:30281/api/proton-mongodb-mgmt/v2/users',
            headers={
                "Content-Type": "application/json",
                "admin-key":"VVNFUk5BTUU6UEFTU1dPUkQ=",
            },
        )
        users = [v['username'] for v in resp.json()]
        assert 'as' in users

        resp = requests.delete(
            url=f'http://{host}:30281/api/proton-mongodb-mgmt/v2/users/at_db/as',
            headers={
                "Content-Type": "application/json",
                "admin-key":"VVNFUk5BTUU6UEFTU1dPUkQ=",
            },
        )
        assert resp.status_code == 204

        resp = requests.get(
            url=f'http://{host}:30281/api/proton-mongodb-mgmt/v2/users',
            headers={
                "Content-Type": "application/json",
                "admin-key":"VVNFUk5BTUU6UEFTU1dPUkQ=",  # cSpell:ignore VVNFUk5BTUU6UEFTU1dPUkQ
            },
        )
        assert resp.json() == None

    @allure.title("数据库账户权限管理:创建/删除/修改用户权限")
    def test_7(self, host):
        resp = requests.put(
            url=f'http://{host}:30281/api/proton-mongodb-mgmt/v2/users/at_db/as',
            headers={
                "Content-Type": "application/json",
                "admin-key":"VVNFUk5BTUU6UEFTU1dPUkQ=",
            },
            data=json.dumps(
                {
                "password":"UEFTU1dPUkQ=" # cSpell:ignore UEFTU1dPUkQ
                }
            )
        )
        assert resp.status_code == 201

        resp = requests.get(
            url=f'http://{host}:30281/api/proton-mongodb-mgmt/v2/users',
            headers={
                "Content-Type": "application/json",
                "admin-key":"VVNFUk5BTUU6UEFTU1dPUkQ=",
            },
        )
        users = [v['username'] for v in resp.json()]
        assert 'as' in users

        resp = requests.put(
            url=f'http://{host}:30281/api/proton-mongodb-mgmt/v2/users/at_db/as/roles',
            headers={
                "Content-Type": "application/json",
                "admin-key":"VVNFUk5BTUU6UEFTU1dPUkQ=",
            },
            data=json.dumps(
                [
                    {
                        "db_name":"at_db",
                        "role": "readWrite",
                    }
                ]
            )
        )
        assert resp.status_code == 204

        resp = requests.patch(
            url=f'http://{host}:30281/api/proton-mongodb-mgmt/v2/users/at_db/as/roles',
            headers={
                "Content-Type": "application/json",
                "admin-key":"VVNFUk5BTUU6UEFTU1dPUkQ=",
            },
            data=json.dumps(
                [
                    {
                        "db_name":"at_db",
                        "role": "read",
                    }
                ]
            )
        )
        assert resp.status_code == 204

    @allure.title("备份管理")
    def test_8(self, host):
        resp = requests.post(
                url=f'http://{host}:30281/api/proton-mongodb-mgmt/v2/backups',
                headers={
                    "Content-Type": "application/json",
                    "admin-key":"VVNFUk5BTUU6UEFTU1dPUkQ=",
                },
                data=json.dumps(
                    {
                    "backup_dirxxx":"/data" # cSpell:ignore dirxxx
                    }
                )
            )
        assert resp.status_code == 400

        resp = requests.post(
            url=f'http://{host}:30281/api/proton-mongodb-mgmt/v2/backups',
            headers={
                "Content-Type": "application/json",
                "admin-key":"VVNFUk5BTUU6UEFTU1dPUkQ=",
            },
            data=json.dumps(
                {
                "backup_dir":"/data"
                }
            )
        )
        assert resp.status_code == 200
        backup_id = resp.json()["id"]
        resp = requests.get(
            url=f'http://{host}:30281/api/proton-mongodb-mgmt/v2/backups',
            headers={
                "Content-Type": "application/json",
                "admin-key":"VVNFUk5BTUU6UEFTU1dPUkQ=",
            }
        )
        assert resp.status_code == 200
        assert resp.json()[0]["id"] == backup_id

        resp = requests.get(
            url=f'http://{host}:30281/api/proton-mongodb-mgmt/v2/backup_size',
            headers={
                "Content-Type": "application/json",
                "admin-key":"VVNFUk5BTUU6UEFTU1dPUkQ=",
            }
        )
        assert resp.status_code == 200
        time.sleep(10)
        stdout, stderr = SSH_CLIENT.exec_cmd(
            cmd_str=f'du -k /data/mongodboperator/rs0/{backup_id}.tar | awk {{\'print $1\'}}',
            timeout=60
        )
        assert resp.json() > int(stdout.replace("\n",""))

        resp = requests.delete(
            url=f'http://{host}:30281/api/proton-mongodb-mgmt/v2/backups/{backup_id}',
            headers={
                "Content-Type": "application/json",
                "admin-key":"VVNFUk5BTUU6UEFTU1dPUkQ=",
            }
        )
        assert resp.status_code == 204
