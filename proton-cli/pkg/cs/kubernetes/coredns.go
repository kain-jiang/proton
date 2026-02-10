package kubernetes

import (
	"bytes"
	"text/template"
)

type coreDNS struct {
	ImageRepository string
	ClusterIP       string
	ClusterIPs      []string
	IPFamilies      []string
	IPFamilyPolicy  string
	TemplateYAML    []byte
}

const coreDNSv186YamlTemplate = `
---
apiVersion: v1
kind: Service
metadata:
  labels:
    k8s-app: kube-dns
    kubernetes.io/cluster-service: "true"
    kubernetes.io/name: "KubeDNS"
  name: kube-dns
  namespace: kube-system
  annotations:
    prometheus.io/port: "9153"
    prometheus.io/scrape: "true"
  # Without this resourceVersion value, an update of the Service between versions will yield:
  #   Service "kube-dns" is invalid: metadata.resourceVersion: Invalid value: "": must be specified for an update
  resourceVersion: "0"
spec:
  clusterIP: {{ .ClusterIP }}
  clusterIPs:
  {{- range .ClusterIPs }}
   - {{ . }}
  {{- end }}
  ipFamilies:
  {{- range .IPFamilies }}
   - {{ . }}
  {{- end }}
  ipFamilyPolicy: {{ .IPFamilyPolicy }}
  ports:
  - name: dns
    port: 53
    protocol: UDP
    targetPort: 53
  - name: dns-tcp
    port: 53
    protocol: TCP
    targetPort: 53
  - name: metrics
    port: 9153
    protocol: TCP
    targetPort: 9153
  selector:
    k8s-app: kube-dns
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: coredns
  namespace: kube-system
  labels:
    k8s-app: kube-dns
spec:
  selector:
    matchLabels:
      k8s-app: kube-dns
  template:
    metadata:
      labels:
        k8s-app: kube-dns
    spec:
      priorityClassName: system-cluster-critical
      serviceAccountName: coredns
      tolerations:
      - key: CriticalAddonsOnly
        operator: Exists
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
      nodeSelector:
        kubernetes.io/os: linux
        node-role.kubernetes.io/master: ""
      containers:
      - name: coredns
        image: {{ .ImageRepository }}/coredns:v1.8.6
        imagePullPolicy: IfNotPresent
        resources:
          limits:
            memory: 170Mi
          requests:
            cpu: 100m
            memory: 70Mi
        args: [ "-conf", "/etc/coredns/Corefile" ]
        volumeMounts:
        - name: config-volume
          mountPath: /etc/coredns
          readOnly: true
        ports:
        - containerPort: 53
          name: dns
          protocol: UDP
        - containerPort: 53
          name: dns-tcp
          protocol: TCP
        - containerPort: 9153
          name: metrics
          protocol: TCP
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
            scheme: HTTP
          initialDelaySeconds: 60
          timeoutSeconds: 5
          successThreshold: 1
          failureThreshold: 5
        readinessProbe:
          httpGet:
            path: /ready
            port: 8181
            scheme: HTTP
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            add:
            - NET_BIND_SERVICE
            drop:
            - all
          readOnlyRootFilesystem: true
      dnsPolicy: Default
      volumes:
        - name: config-volume
          configMap:
            name: coredns
            items:
            - key: Corefile
              path: Corefile
  updateStrategy:
    rollingUpdate:
      maxUnavailable: 1
    type: RollingUpdate

---
# CoreDNSConfigMap is the CoreDNS ConfigMap manifest
apiVersion: v1
kind: ConfigMap
metadata:
  name: coredns
  namespace: kube-system
data:
  Corefile: |
    .:53 {
        errors
        health {
           lameduck 5s
        }
        ready
        kubernetes cluster.local in-addr.arpa ip6.arpa {
           pods insecure
           fallthrough in-addr.arpa ip6.arpa
           ttl 30
        }
        prometheus :9153
        forward . /etc/resolv.conf
        cache 30
        loop
        reload
        loadbalance
    }

---
# CoreDNSClusterRole is the CoreDNS ClusterRole manifest
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:coredns
rules:
- apiGroups:
  - ""
  resources:
  - endpoints
  - services
  - pods
  - namespaces
  verbs:
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - get
- apiGroups:
  - discovery.k8s.io
  resources:
  - endpointslices
  verbs:
  - list
  - watch

---
# CoreDNSClusterRoleBinding is the CoreDNS Clusterrolebinding manifest
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:coredns
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:coredns
subjects:
- kind: ServiceAccount
  name: coredns
  namespace: kube-system

---
# CoreDNSServiceAccount is the CoreDNS ServiceAccount manifest
apiVersion: v1
kind: ServiceAccount
metadata:
  name: coredns
  namespace: kube-system
`

func (c *coreDNS) renderTemplate(t string) error {
	tmpl, err := template.New("").Parse(t)
	if err != nil {
		return err
	}

	var renderTemplate bytes.Buffer
	if err := tmpl.Execute(&renderTemplate, c); err != nil {
		return err
	}
	c.TemplateYAML = renderTemplate.Bytes()
	return nil
}
