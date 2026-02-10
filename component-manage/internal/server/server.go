package server

// @title			Component Manage API
// @version		0.0.1
// @description	This is a server to manage resource component
// @contact.name	Kain.Jiang
// @contact.email	Kain.Jiang@aishu.cn
// @BasePath		/
func Main() error {
	svc := &Serve{}
	go svc.RunHttpServe()
	return svc.WaitShutDownHttpServe()
}
