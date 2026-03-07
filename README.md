# ipinfo

查询 IP 地址的地理位置与网络信息的命令行工具。

支持多个公共 API 提供商、代理查询、带熔断保护的分级随机轮询，以及内存 TTL+LRU 缓存。

## 功能特性

- 查询任意 IP 或自动获取本机出口 IP
- **分级随机轮询**：同一优先级内随机打散请求，均衡各提供商用量
- **熔断保护**：连续失败达阈值后自动跳过故障提供商，超时后探测恢复
- **TTL + LRU 缓存**：避免重复查询，二次查询几乎即时返回
- **代理支持**：`http://`、`https://`、`socks5h://` 均可
- **YAML 配置**：提供商优先级、Token、超时均可自定义
- 输出格式：美化 JSON

## 安装

### 一键安装（推荐）

```bash
curl -fsSL https://raw.githubusercontent.com/lupguo/ip_info/main/install.sh | bash
```

自动检测系统和架构，从 GitHub Release 下载二进制并校验 SHA256，同时将示例配置写入 `~/.ipinfo/config.yaml`（已有则不覆盖）。

### 从源码编译

```bash
git clone https://github.com/lupguo/ip_info.git
cd ip_info
make build      # 输出 ./ipinfo
# 或
make install    # 安装到 $GOPATH/bin
```

## 使用

```bash
# 查询本机出口 IP
ipinfo

# 查询指定 IP
ipinfo 8.8.8.8

# 详细模式（含 ASN、ISP、类型、VPN 标记等）
ipinfo -d 8.8.8.8

# 通过 HTTP 代理查询
ipinfo -x http://127.0.0.1:8118

# 通过 SOCKS5 代理查询
ipinfo -x socks5h://127.0.0.1:2222

# 使用自定义配置文件
ipinfo -c ./config.example.yaml 8.8.8.8
```

### 输出示例

**默认模式**

```json
{
    "ip": "8.8.8.8",
    "country": "US",
    "region": "California",
    "city": "Mountain View",
    "org": "AS15169 Google LLC",
    "lat": 37.4056,
    "lon": -122.0775,
    "zip": "94043",
    "timezone": "America/Los_Angeles",
    "source": "ipinfo.io"
}
```

**详细模式（`-d`）**

```json
{
    "ip": "8.8.8.8",
    "country": "US",
    "region": "California",
    "city": "Mountain View",
    "org": "AS15169 Google LLC",
    "lat": 37.4056,
    "lon": -122.0775,
    "zip": "94043",
    "timezone": "America/Los_Angeles",
    "asn": "AS15169",
    "isp": "Google LLC",
    "type": "datacenter",
    "source": "ipwho.is"
}
```

## 提供商与优先级

| 优先级 | 提供商                   | 数据类型       | 是否需要 Token |
|-----|-----------------------|------------|------------|
| 1   | ipinfo.io             | 完整地理 + 组织  | 可选（提升配额）   |
| 2   | ip-api.com            | 完整地理 + ASN | 否          |
| 2   | ipapi.co              | 完整地理 + ASN | 否          |
| 3   | ipgeolocation.io      | 完整地理 + 安全  | 是          |
| 3   | ipwho.is              | 完整地理 + 类型  | 否          |
| 3   | ipstack.com           | 完整地理       | 是（默认禁用）    |
| 4   | api.ipify.org         | 仅 IP       | 否          |
| 4   | icanhazip.com         | 仅 IP       | 否          |
| 4   | checkip.amazonaws.com | 仅 IP       | 否          |

同一优先级内随机打散，低优先级仅在高优先级全部失败后使用。

## 配置

首次运行时自动生成 `~/.ipinfo/config.yaml`，也可手动编辑：

```yaml
cache:
  ttl: 5m
  max_size: 1000

circuit_breaker:
  max_failures: 5
  reset_timeout: 30s

providers:
  - name: ipinfo.io
    base_url: https://ipinfo.io
    priority: 1
    token: ""       # 填入 Token 可提升速率限制
    enabled: true
    timeout: 5s
  # ... 更多提供商见 config.example.yaml
```

## 跨平台发布

推送 `v*` 标签后，GitHub Actions 自动编译 5 个平台的二进制并上传到 Release：

| 平台      | 架构                    |
|---------|-----------------------|
| Linux   | amd64 / arm64         |
| macOS   | amd64 / arm64 (M1/M2) |
| Windows | amd64                 |

```bash
git tag v1.0.0 && git push origin v1.0.0
```

## Makefile

```bash
make build    # 编译当前平台
make cross    # 编译全平台（输出到 dist/）
make install  # 安装到 $GOPATH/bin
make test     # 运行测试
make lint     # go vet 检查
make clean    # 清理构建产物
```

## 依赖

- [cobra](https://github.com/spf13/cobra) — CLI 框架
- [yaml.v3](https://gopkg.in/yaml.v3) — 配置解析
- [golang.org/x/net](https://pkg.go.dev/golang.org/x/net) — SOCKS5 代理支持

缓存与熔断均使用标准库 `sync` 自行实现，无额外依赖。

## License

MIT
