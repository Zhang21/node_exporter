# node_exporter执行自定义脚本

在`node_exporter`自定义一个`custom_script.go` collector，来执行自定义的脚本。当然，也可以使用脚本往pushgateway推送数据，然后prometheus从pushgateway取数据。

<br>
<br>


## 一些配置

```
# 编译
go test
go build

# 查看帮助
node_exporter -h

# custom_script默认禁用，启用
node_exporter --collector.custom_script
# 默认自定以脚本路径 /opt/prometheus/customScript
node_exporter --collector.custom_script --collector.customscript.scriptPath="/dir/path"
```

<br>

示例脚本:

```shell
#!/bin/bash

########
# Description: A custom shell script which exec by node_exporter when prometheus pull metrics.
########

key="process_nums"
value=$(ps -ef | wc -l)

echo "$key=$value"
```

不符合此输出格式的脚本，将进行异常处理，打出error错误日志并跳过。
