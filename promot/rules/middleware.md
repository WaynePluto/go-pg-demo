# 扩展中间件规范
参考 internal/middlewares/logger.go中间件的创建步骤：
1. 在 internal/middlewares目录下创建新的中间件文件；
2. 对于全局应用的中间件，在internal/middlewares/provider.go的NewUseMiddlewares内直接注册；