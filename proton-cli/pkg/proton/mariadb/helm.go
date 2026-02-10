package mariadb

const (
	OldHelmReleaseMariaDBName    = "proton-mariadb"
	OldHelmReleaseMariaDBVersion = "1.10.3"

	OldHelmReleaseHAProxyName    = "proton-rds-haproxy"
	OldHelmReleaseHAProxyVersion = "1.1.0"
)

const DeleteOldHelmReleaseMariaDBStderr = "Error: deletion completed with 2 error(s): release \"proton-mariadb\": object \"\" not found, skipping delete; release \"proton-mariadb\": object \"\" not found, skipping delete\n"
