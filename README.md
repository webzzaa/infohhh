# 小程序敏感信息扫描工具

这是一个用Go语言编写的工具，用于扫描反编译后的小程序代码，根据配置文件中的规则，查找其中的敏感信息（如API路径、URL、账号密码、密钥等），并将结果输出到Excel文件中。

## 功能特性

- 支持多种敏感信息类型的检测：
  - API路径
  - URL链接
  - 账号密码
  - 密钥（API Key、Secret Key等）
  - 手机号
  - 身份证号
  - JDBC数据库配置
  - 静态资源路径
- 可自定义检测规则（通过config.yaml文件）
- 支持扫描指定目录
- 结果输出为格式美观的Excel文件

## 安装依赖

确保您的系统已安装Go 1.16或更高版本，然后运行：

```bash
go mod tidy
```

## 使用方法

### 基本用法

```bash
go run main.go -dir 要扫描的目录路径 -output 输出的Excel文件路径
```

### 参数说明

- `-config`：配置文件路径，默认为`config.yaml`
- `-dir`：要扫描的目录路径，默认为当前目录
- `-output`：输出的Excel文件路径，默认为`sensitive_info.xlsx`

### 示例

```bash
# 扫描当前目录，输出结果到默认文件
go run main.go

# 扫描指定目录，输出结果到指定文件
go run main.go -dir ./decompiled_miniprogram -output 小程序敏感信息.xlsx

# 使用自定义配置文件
go run main.go -config ./custom_config.yaml -dir ./scan_dir -output result.xlsx
```

## 配置文件说明

配置文件使用YAML格式，示例如下：

```yaml
rules:
- id: 1
  name: API
  enable: true
  regexes:
    - (?i)(?:^|[^\\])["'](/+[^'"]+/*(?:\?[^'"]*)?)["']
    # 更多正则表达式...
```

每条规则包含以下字段：

- `id`：规则ID，唯一标识
- `name`：规则名称，用于在结果中显示
- `enable`：是否启用该规则
- `regexes`：用于匹配敏感信息的正则表达式列表

## 输出结果

输出的Excel文件包含以下列：

- 文件名：包含敏感信息的文件名
- 行号：敏感信息在文件中的行号
- 行内容：包含敏感信息的完整行内容
- 规则ID：匹配的规则ID
- 规则名称：匹配的规则名称
- 匹配文本：具体匹配到的敏感信息

## 注意事项

1. 请确保您有合法权限扫描相关代码
2. 扫描结果可能包含误报，建议人工审核
3. 对于大型项目，扫描可能需要较长时间
4. 请及时更新配置文件中的规则，以适应新的敏感信息类型



