package v1

type NginxHTTP struct {
	Name string `json:"name,omitempty"`

	Conf NginxHTTPConf `json:"conf,omitempty"`
}

type NginxHTTPConf struct {
	Server NginxServer `json:"server,omitempty"`

	Upstream map[string]NginxUpstream `json:"upstream,omitempty"`
}

type NginxServer struct {
	Listen string `json:"listen,omitempty"`

	ClientMaxBodySize string `json:"client_max_body_size,omitempty"`

	Locations []map[string]map[string]interface{} `json:"location,omitempty"`

	ProxyRequestBuffering string `json:"proxy_request_buffering,omitempty"`

	AccessLog string `json:"access_log,omitempty"`

	SSLCertificate string `json:"ssl_certificate,omitempty"`

	SSLCertificateKey string `json:"ssl_certificate_key,omitempty"`

	AddHeaders []string `json:"add_header,omitempty"`

	IF interface{} `json:"if,omitempty"`
}

type NginxUpstream struct {
	CheckHTTPSend string `json:"check_http_send,omitempty"`

	Check string `json:"check,omitempty"`

	Servers interface{} `json:"server,omitempty"`

	Keepalive string `json:"keepalive,omitempty"`
}
