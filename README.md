# Docker Registry HTTP API V2 Implementation

BYRTeam 2025 考核题 Docker It Yourself 实现

一个用 Go 语言实现的轻量级 Docker Registry HTTP API V2 服务器，支持完整的 manifest 和 blob 管理功能。

## 🚀 功能特性

### 核心 API 支持
- **Manifest 管理** - 支持获取、上传、检查和删除
- **Blob 管理** - 支持分片上传、下载和跨仓库挂载
- **错误处理** - 符合 Docker Registry API 规范的标准错误响应

### 架构特点
- **分层架构** - Handler → Storage → FileSystem
- **接口驱动** - 可扩展的存储驱动设计
- **参数对象模式** - 类型安全的接口设计
- **完整的错误处理** - 标准化的错误码和响应格式

## 支持的 API 端点

### Manifest API
| 方法 | 路径 | 描述 |
|------|------|------|
| `GET` | `/v2/{name}/manifests/{reference}` | 获取 manifest |
| `PUT` | `/v2/{name}/manifests/{reference}` | 上传 manifest |
| `HEAD` | `/v2/{name}/manifests/{reference}` | 检查 manifest 是否存在 |
| `DELETE` | `/v2/{name}/manifests/{reference}` | 删除 manifest |

### Blob API
| 方法 | 路径 | 描述 |
|------|------|------|
| `HEAD` | `/v2/{name}/blobs/{digest}` | 检查 blob 是否存在 |
| `GET` | `/v2/{name}/blobs/{digest}` | 获取 blob 内容 |
| `POST` | `/v2/{name}/blobs/uploads/` | 初始化 blob 上传 |
| `GET` | `/v2/{name}/blobs/uploads/{uuid}` | 获取上传状态 |
| `PATCH` | `/v2/{name}/blobs/uploads/{uuid}` | 上传 blob 数据块 |
| `PUT` | `/v2/{name}/blobs/uploads/{uuid}?digest={digest}` | 完成 blob 上传 |
| `DELETE` | `/v2/{name}/blobs/uploads/{uuid}` | 取消 blob 上传 |

## 项目结构

```
my_docker_registry/
├── cmd/registry/
│   └── main.go                 # 应用程序入口
├── internal/
│   ├── handler/
│   │   └── handler.go          # HTTP 请求处理层
│   ├── storage/
│   │   ├── interface.go        # 存储驱动接口定义
│   │   └── filesystem.go       # 文件系统存储实现
│   └── types/
│       ├── errors.go           # 标准错误定义
│       ├── blob.go             # Blob 相关数据结构
│       └── manifest.go         # Manifest 相关数据结构
├── go.mod
├── go.sum
└── registry_data/              # 数据存储目录（运行时创建）
    ├── blobs/                  # Blob 存储（按摘要组织）
    │   └── sha256/
    │       └── {xx}/           # sha256 前两位作为索引
    │           └── {hash}
    └── repositories/           # 仓库数据
        └── {name}/
            ├── _manifests/
            │   ├── revisions/sha256/{hash}
            │   └── tags/{tag}/current/link
            └── _uploads/{uuid}/
```

## 快速开始

### 环境要求
- Go 1.21
- 网络端口 5000（可配置）

### 安装和运行

1. **克隆项目**
```bash
git clone https://github.com/Tanpinsary/my_docker_registry.git
cd my_docker_registry
```

2. **安装依赖**
```bash
go mod download
```

3. **编译项目**
```bash
go build ./cmd/registry
```

4. **运行服务器**
```bash
./registry
```

服务器将在 `http://localhost:5000` 启动，数据将存储在 `./registry_data` 目录中。

### 错误响应格式

所有错误都遵循 Docker Registry API 标准格式：

```json
{
  "errors": [
    {
      "code": "BLOB_UNKNOWN",
      "message": "blob unknown",
      "detail": {
        "digest": "sha256:abc123"
      }
    }
  ]
}
```

## 📚 支持的错误码

| 错误码 | HTTP 状态 | 描述 |
|--------|-----------|------|
| `BLOB_UNKNOWN` | 404 | Blob 不存在 |
| `MANIFEST_UNKNOWN` | 404 | Manifest 不存在 |
| `BLOB_UPLOAD_UNKNOWN` | 404 | 上传会话不存在 |
| `DIGEST_INVALID` | 400 | 摘要格式无效 |
| `MANIFEST_INVALID` | 400 | Manifest 格式无效 |
| `RANGE_INVALID` | 416 | Content-Range 无效 |
| `UNSUPPORTED` | 400 | 操作不支持 |

## 🔍 存储结构说明

### Blob 存储
- 路径：`blobs/sha256/{前两位}/{完整摘要}`
- 示例：`blobs/sha256/ab/abcdef123...`
- 特点：全局去重，跨仓库共享

### Manifest 存储
- 内容文件：`repositories/{name}/_manifests/revisions/sha256/{hash}`
- 标签链接：`repositories/{name}/_manifests/tags/{tag}/current/link`
- 临时上传：`repositories/{name}/_uploads/{uuid}/`

## 许可证

MIT License