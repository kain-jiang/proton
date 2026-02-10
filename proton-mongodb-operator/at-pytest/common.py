import paramiko
import pymongo

class SSHClient:
    def __init__(self, host, port, user, password):
        self.client = paramiko.SSHClient()
        self.client.load_system_host_keys()
        self.client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
        self.client.connect(hostname=host, port=port, username=user, password=password, timeout=10)
    def exec_cmd(self, cmd_str, timeout=30):
        stdin, stdout, stderr = self.client.exec_command(cmd_str, timeout=timeout)
        out, err = stdout.read().decode(), stderr.read().decode()
        # print(f"Exec cmd {cmd_str}, out: {out}, err: {err}")
        return out, err
    def scp(self, src, dst):
        c = self.client.open_sftp()
        c.put(localpath=src, remotepath=dst)
        c.close()
    def __del__(self):
        self.client.close()

class DBClient:
    def __init__(self, host, port, user, password):
        self.dbc = pymongo.MongoClient(
            host=host,
            port=port,
            username=user,
            password=password,
            authSource='admin',
        )
    def __del__(self):
        self.dbc.close()

