忘老师见谅，我平常工作太忙，没办法开发完成所有功能,后续会不断在工作中优化这个项目的

项目工程目录说明
api     各服务协议定义

app     应用服务
    cmd 启动主程序
    admin   后台管理服务
        service
    user    用户服务
        service 
            biz     业务接口
            conf    配置对象
            data    数据处理
            model   业务model
route   路由
conf
    config.yaml    
common 公共处理方法
    dateTime.go 日期时间处理

doc     说明文档以及参考资料

pkg     公共工具包
    es
    gorm
    redis
    kafka
    ...
web     前端
    admin
    user
