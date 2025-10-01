# Docker Registry HTTP API V2 Implementation

BYRTeam 2025 考核题 Docker It Yourself 实现

一个用 Go 语言实现的轻量级 Docker Registry HTTP API V2 服务器，支持完整的 manifest 和 blob 管理功能。

## 使用时间

https://wakatime.com/share/badges/projects?q=my_docker_registry

前期花了几个小时读文档写文档，大概理清思路之后开始设计接口，把文档喂给 ai 接口写的和文档要求差别好大，于是接口大部分是自己设计的，剩下的存储层和处理层几乎都是 vibe coding（），wakatime 记录的主要是项目开始到写现在 README 用的时间。
由于是第一次接触这种偏向于对着文档造轮子的项目，还是挺痛苦的（）。前期瞎几把整理的文档贴在 [这里](/tanp's_docs.md) 了， yysy 感觉还不如直接对着官方文档看（）。

## 功能

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

## 测试 API

### 基本功能测试

**1. 检查 API 版本**
```bash
curl -i http://localhost:5000/v2/
```

**2. 初始化 blob 上传**
```bash
curl -i -X POST http://localhost:5000/v2/test/blobs/uploads/
```

### 错误响应测试

**测试 404 - Blob 不存在**
```bash
curl -i http://localhost:5000/v2/test/blobs/sha256:0000000000000000000000000000000000000000000000000000000000000000
```

**测试 404 - 上传会话不存在**
```bash  
curl -i http://localhost:5000/v2/test/blobs/uploads/invalid-uuid
```

**测试 400 - 缺少 Content-Range**
```bash
curl -i -X PATCH http://localhost:5000/v2/test/blobs/uploads/{uuid} -d "data"
```

**测试 400 - 缺少 digest 参数**
```bash
curl -i -X PUT http://localhost:5000/v2/test/blobs/uploads/{uuid}
```

所有错误响应都将返回符合 Docker Registry API 标准的 JSON 格式错误信息。

### 错误响应格式

所有错误都遵循 Docker Registry HTTP API V2 标准格式：

```json
{
  "errors": [
    {
      "code": "BLOB_UNKNOWN",
      "message": "blob unknown to registry",
      "detail": {
        "digest": "sha256:abc123..."
      }
    }
  ]
}
```

#### 常见错误响应示例

**404 - Blob 不存在**
```json
{
  "errors": [
    {
      "code": "BLOB_UNKNOWN",
      "message": "blob unknown to registry",
      "detail": {"digest": "sha256:0000..."}
    }
  ]
}
```

**404 - 上传会话不存在**
```json
{
  "errors": [
    {
      "code": "BLOB_UPLOAD_UNKNOWN", 
      "message": "blob upload unknown to registry",
      "detail": {"uuid": "invalid-uuid"}
    }
  ]
}
```

**400 - 缺少必需参数**
```json
{
  "errors": [
    {
      "code": "DIGEST_INVALID",
      "message": "provided digest did not match uploaded content",
      "detail": {"digest": "digest parameter required"}
    }
  ]
}
```

**400 - 无效的 Content-Range**
```json
{
  "errors": [
    {
      "code": "RANGE_INVALID",
      "message": "invalid content range", 
      "detail": {"range": "Content-Range header required"}
    }
  ]
}
```

## 📚 API 响应码规范

### Manifest API 响应码
| API | 成功响应 | 错误响应 |
|-----|----------|----------|
| **GET Manifest** | 200 获取成功 | 400 无效名称或引用<br>404 Repository 或 manifest 不存在 |
| **PUT Manifest** | 201 创建成功 | 400 无效名称、引用或 manifest<br>404 Repository 不存在 |
| **HEAD Manifest** | 200 Manifest 存在 | 404 Manifest 不存在 |
| **DELETE Manifest** | 202 删除成功 | 404 Manifest 或 repository 不存在 |

### Blob API 响应码
| API | 成功响应 | 错误响应 |
|-----|----------|----------|
| **HEAD Blob** | 200 Blob 存在 | 404 Blob 不存在 |
| **GET Blob** | 200 内容返回 | 404 Blob 不存在 |
| **POST Upload Init** | 201 挂载成功<br>202 上传初始化成功 | 404 Repository 不存在 |
| **GET Upload Status** | 204 上传进行中 | 404 上传会话不存在 |
| **PATCH Upload Chunk** | 202 Chunk 接受 | 400 格式错误或范围无效<br>404 上传会话不存在 |
| **PUT Complete Upload** | 201 上传完成 | 400 无效摘要或缺少参数<br>404 上传会话不存在 |
| **DELETE Cancel Upload** | 204 会话取消成功 | 404 上传会话不存在 |

## 🚫 标准错误码

| 错误码 | HTTP 状态 | 消息 | 使用场景 |
|--------|-----------|------|----------|
| `BLOB_UNKNOWN` | 404 | blob unknown to registry | Blob 不存在 |
| `MANIFEST_UNKNOWN` | 404 | manifest unknown | Manifest 不存在 |
| `NAME_UNKNOWN` | 404 | repository name not known to registry | Repository 不存在 |
| `BLOB_UPLOAD_UNKNOWN` | 404 | blob upload unknown to registry | 上传会话不存在 |
| `NAME_INVALID` | 400 | invalid repository name | 无效的仓库名称 |
| `MANIFEST_INVALID` | 400 | manifest invalid | 无效的 manifest 格式 |
| `DIGEST_INVALID` | 400 | provided digest did not match uploaded content | 摘要不匹配 |
| `RANGE_INVALID` | 400 | invalid content range | Content-Range 头无效 |
| `BLOB_UPLOAD_INVALID` | 400 | blob upload invalid | 无效的上传参数 |

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