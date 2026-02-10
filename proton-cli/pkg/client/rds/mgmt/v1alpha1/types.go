package v1alpha1

type Database struct {
	DBName    string    `json:"db_name,omitempty"`
	Charset   Charset   `json:"charset,omitempty"`
	Collation Collation `json:"collate,omitempty"`
}

type User struct {
	Username   string      `json:"username,omitempty"`
	Privileges []Privilege `json:"privileges,omitempty"`
	SSLType    SSLType     `json:"ssl_type,omitempty"`
}

type Privilege struct {
	DBName        string        `json:"db_name,omitempty"`
	PrivilegeType PrivilegeType `json:"privilege_type,omitempty"`
}

type PrivilegeType string

const (
	PrivilegeReadWrite = "ReadWrite"
	PrivilegeReadOnly  = "ReadOnly"
	PrivilegeDDLOnly   = "DDLOnly"
	PrivilegeDMLOnly   = "DMLOnly"
	PrivilegeNone      = "None"
	PrivilegeAll       = "All"
)

type SSLType string

const (
	SSLAny  = "Any"
	SSLNone = "None"
)

type Charset string

const (
	CharsetUTF8MB4 = "utf8mb4"
)

type Collation string

const (
	CollationUTF8MB4GeneralCI = "utf8mb4_general_ci"
)
