# fabric_project_simple
go 语言简版的fabric项目，基于fabric1.4，包括的简单的chaincode，单节点的peer、order。 下载后直接放到goPath路径下即可。

v1.0 是一个最简单的版本，代码比较粗糙，但也可以正常的运行；

v2.0 是在第一个版本的基础上添加了一个core.yml 文件，专门用来做配置项的，并把v1.0 版本中的很多配置改成了读取配置文件的内容； 在main文件的，先读取配置文件； 在util文件中通过很多get函数获取各个配置项的内容；

v3.0 其是在2.0版本的基础上增加了Restful服务功能，可以通过接口访问数据

v4.0 把fabric中的默认的stateDB从levelDB改成couchDB，并使用nosql语句进行查询
