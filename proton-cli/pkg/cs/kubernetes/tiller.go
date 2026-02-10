package kubernetes

import (
	"bytes"
	"text/template"
)

type Tiller struct {
	ImageRepository string
	IPFamilies      []string
	IPFamilyPolicy  string
	TemplateYAML    []byte
}

const TillerYamlTemplate = `
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: helm
    name: tiller
  name: tiller-deploy
  namespace: kube-system
spec:
  ipFamilies:
  {{- range .IPFamilies }}
   - {{ . }}
  {{- end }}
  ipFamilyPolicy: {{ .IPFamilyPolicy }}
  ports:
  - name: tiller
    port: 44134
    targetPort: tiller
  selector:
    app: helm
    name: tiller
  type: ClusterIP
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    app: helm
    name: tiller
  name: tiller-deploy
  namespace: kube-system
spec:
  selector:
    matchLabels:
      app: helm
      name: tiller
  template:
    metadata:
      labels:
        app: helm
        name: tiller
    spec:
      automountServiceAccountToken: true
      nodeSelector:
        kubernetes.io/os: linux
        node-role.kubernetes.io/master: ""
      containers:
      - env:
        - name: TILLER_NAMESPACE
          value: kube-system
        - name: TILLER_HISTORY_MAX
          value: "0"
        - name: KUBERNETES_SERVICE_HOST
          value: kubernetes.default
        - name: KUBERNETES_SERVICE_PORT
          value: "443"
        image: {{ .ImageRepository }}/kubernetes-helm/tiller:v2.16.9
        imagePullPolicy: IfNotPresent
        livenessProbe:
          httpGet:
            path: /liveness
            port: 44135
          initialDelaySeconds: 1
          timeoutSeconds: 1
        name: tiller
        command:
          - /tiller
          - -storage=secret
        ports:
        - containerPort: 44134
          name: tiller
        - containerPort: 44135
          name: http
        readinessProbe:
          httpGet:
            path: /readiness
            port: 44135
          initialDelaySeconds: 1
          timeoutSeconds: 1
      serviceAccountName: tiller
  updateStrategy:
    rollingUpdate:
      maxUnavailable: 1
    type: RollingUpdate

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: tiller
  namespace: kube-system

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: tiller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
  - kind: ServiceAccount
    name: tiller
    namespace: kube-system
`

func (c *Tiller) renderTemplate(t string) error {
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
