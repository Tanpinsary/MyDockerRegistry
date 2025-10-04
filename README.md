# Docker Registry HTTP API V2 Implementation

BYRTeam 2025 考核题 Docker It Yourself 报告

一个用 Go 语言实现的轻量级 Docker Registry HTTP API V2 服务器，支持完整的 manifest 和 blob 管理功能。

项目采用分层架构设计：

- Handler层：HTTP请求处理
- Storage层：存储驱动接口和文件系统实现
- Types层：数据结构和错误处理

## 使用时间

[![wakatime](https://wakatime.com/badge/user/a0a1a76d-0ee4-4a97-8f98-32faeaba5485/project/fa2900c8-1179-4984-84e6-f87a8b623dba.svg)](https://wakatime.com/badge/user/a0a1a76d-0ee4-4a97-8f98-32faeaba5485/project/fa2900c8-1179-4984-84e6-f87a8b623dba)

## 简要报告

由于是第一次接触这种偏向于对着文档造轮子的项目，还是挺痛苦的。前期瞎几把整理的文档贴在 [这里](/tanp's_docs.md) 了， yysy 写的非常非常乱，感觉还不如直接对着官方文档看。等到初步完工了才想起来可以用 APIFOX 做管理和测试，所以交了问卷又返工，这玩意真是比自己截图方便 998244353 倍啊。

[APIFOX 文档](https://m3kc0i3ss7.apifox.cn/)

前期花了几个小时读文档写文档，大概理清思路之后设计项目结构，然后花费大量时间设计接口，剩下的存储层和处理层的代码大部分为 Vibe Coding 主导 + 人工 Review。AI 生成代码使用 vsc 集成 copilot，模型为 Claude Sonnet 4 以及 Gemini 2.5 pro。

### 心路历程

这道题开始拿到的时候感觉“除了 Docker 这个词之外没几个认识的”，加上自己不能做 BYR-Archive 那题只能做这个其实挺恐慌的。不过在群里 xdd 老师提示读懂文档就完成大部分的时候找回了信心，花了几个小时用 DeepL + Gemini 把文档通读一遍大致理解 Registry 到底是干什么之后，其实本身已经没什么特别可怕的地方了。

不过本人其实没啥后端开发经验，好在 Vibe Coding 已经把写代码的难度从 *“精通才能写好”* 降到了 *“理解需求 + 会说话能写个差不多”* ，加之时间给的也很充裕（导致我做的很松弛），边学边做也并非不可能。

## 实现思路

/internal/types & /internal/storage/interface.go -> /internal/storage/filesystem.go -> /internal/handler/handler.go

### step1

花费时间最大的部分大概是第一步（types 设计和接口设计）。众所周知 Golang 是一门没有类和继承概念的静态编译型语言，意味者 struct 设计和品味的重要性。

对于这个项目多达 11 个的 api，为了提高代码复用，本人设计时经常左脑搏击右脑，几个不同 api 可以实现复用的结构捋不清楚，加上我把文档喂给 AI 写的还和文档要求大相径庭，于是乎自己推翻重来对着官方文档和自己写的文档改了好几次。哎还是太菜了。

### step2 & step3

贯彻 Vibe Coding 主导 + 人工 Review （）。其实接口写好、参数写好，代码本身的实现问题并非很大，基本都交给了 AI 来写然后我再 Review。

### 缺点

首先完全没考虑并发问题。希望这点不计入考量（）；其次最大的缺点是本人太菜 Vibe 有点多...

## 生产部署

图省事部署到了 [render](https://mydockerregistry.onrender.com) 上。

```bash
# API 访问
curl https://mydockerregistry.onrender.com/v2/
```

## 实现功能 （以下部分主要由 AI 总结生成）

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

## API 响应码规范

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

## 标准错误码

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

## 存储结构说明

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
