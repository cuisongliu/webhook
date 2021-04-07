# k8s webhook手脚架

1. 自动签发证书
   
    - 进入 `example/cert/testdata` 执行 `kubectl apply -f webhook_init.yaml` 进行初始化webhook
    - 进入 `example/cert` 调整参数执行cert.go
    - 进入 `example/cert/testdata` 执行 `kubectl get -f webhook_init.yaml -o yaml ` 验证是否替换证书和service成功
    - `example/cert/rbac.yaml`是需要的rbac，用管理员权限可忽略
   
2. 普通webhook （借鉴kubebuilder实现）
   
   - main.go加入方法
   
     ```go
      hookServer := mgr.GetWebhookServer()
      webhook.SetupWebhook(hookServer, mgr)
     ```
     
      
