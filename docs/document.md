# 开发文档

## 技术选型

### 编程语言：[Golang With Go Module](https://golang.org/)

Golang with go module 需要 go 的版本在 1.11 / 1.12 

### 基本框架：[Iris](https://iris-go.com/)

使用当前的最新版本

- [中文文档](https://learnku.com/docs/iris-go/10)

- [官方教程](https://docs.iris-go.com/)

### 数据库：

待定 （Mysql / Mongodb)

### 缓存 [Redis](https://github.com/go-redis/redis)

主要用于缓存 session

### 架构设计

MVC (Router -> Controller -> Services -> Model)

// 具体流程待定，大致框架可以查看 iris 文档中的说明

- Controller负责验证和转发从Router中传递过来的参数，并对请求做出应答。Controller实际与Request和Response接触。

- Model负责数据操作，封装与数据库进行操作的逻辑。

参考目录结构如下




### 测试工具
// 简易测试，并非测试框架

- [postman](https://www.getpostman.com/) : 测试后端 api


## 参考链接

- [Go 教程](https://go.wuhaolin.cn/)

- [Go 标准库](https://books.studygolang.com/The-Golang-Standard-Library-by-Example/)