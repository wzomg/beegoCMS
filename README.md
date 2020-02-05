# beegoCMS
一个适合拿来练手的GoWeb框架：beego小项目——文章发布系统

### 项目简介

技术点：golang，redis，mysql，html5，css3，jquery

Go-Web框架：beego

### 快速上手

1、首先要安装好golang开发环境，[win10下安装Go环境（点我直达）](https://www.jianshu.com/p/a86b4b3cb9ba)、[linux下安装Go环境（点我直达）](https://www.jianshu.com/p/09480e44b87a)；git工具；前2个工具配置好用，安装bee工具：`go get -u github.com/beego/bee`

2、将当前项目克隆放置于GOPATH/src目录下，需要修改2个配置文件中的redis和mysql登录信息：.../controllers/utils/redisPool.go、.../models/model.go

3、需要安装的依赖包
```
go get github.com/gomodule/redigo/redis
go get github.com/astaxie/beego/orm
go get -u github.com/go-sql-driver/mysql
```
