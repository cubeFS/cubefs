# 保活脚本
保活脚本将在3.5.1版本提供

## 运行指南
进入/tool/keepalive目录，执行安装脚本
```
./cbfs_auto_install.sh arg1 arg2 arg3 arg4 arg5 arg6 arg7 arg8 arg9
```

* 执行命令前需要先创建用户和卷
* 如果在非root权限下执行，执行命令前需要在 `/etc/fuse.conf` 中添加 `user_allow_other`

## 参数说明
arg1：脚本和二进制客户端放置目录, 如 `/home/service/app/cfs`
arg2：mountPoint 需要挂载的本地目录（如果需要启动多个实例，需要区分）
arg3：volName 申请的卷名, 如 ltptest
arg4：owner 申请的账号，如 prod_xxx
arg5：accessKey
arg6：secretKey
arg7：master地址
arg8：客户端日志目录（如果需要启动多个实例，需要区分）
arg9：只读挂载，可选参数，true为只读挂载。

## 运行说明
`cbfs_auto_install.sh`会复制`cbfs_restart.sh`和`cbfs_clean.sh`到指定路径`arg1`，在`arg1/volname`路径下创建`client.conf`并将cfs-client复制到该路径下，最后会将重启规则添加到`crontab`中。cbfs_restart脚本默认每分钟执行一次。

cbfs_restart.sh：用于重启崩溃的客户端，定时自动挂载
cbfs_clean.sh：用于清理挂载点和 crontab 规则，需要手动执行
