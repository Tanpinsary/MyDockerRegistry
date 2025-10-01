---
title: " Docker Registry HTTP API V2 规范"
date: 2025-09-21 21:31:20
categories: 开发
tags:
  - docker
  - 后端
---
整理一下 BYR2025 后端考核题 Docker It Yourself 需要的 Docker Registry API 规范文档内容，供项目编写使用。内容参考网站为 [Docker Registry HTTP API V2 规范](https://docs.docker.com/reference/api/registry/latest/)

---

## 须知：

1. 此 API 中的所有端点均以**版本号**和**存储库名称**作为前缀。

   ```
   /v2/<name>
   ```

   例如，我想和 `library/ubuntu` 库进行交互，则使用：

   ```
   /v2/library/ubuntu
   ```

2. Autunentication 不做要求

## Docker Registry 构成

Registry 一共要维护三类实体的信息：blob, manifest 和 manifest list.

Manifest 是描述镜像的 JSON 文档，包含其配置 blob、各每层 blob 的 digest，以及平台类型和注释等元数据。

Blob 是从 Manifest 文件中引用的二进制对象，包含 json config 和若干 tar 包。

Manifest list 是一个 Tag 所对应所有 Manifest 的元数据列表。获取 manifest list 后，通过对应的 digest 值获取 manifest, 再根据 manifest 的内容下载所有 Blob。

## Pulling Images

拉取镜像，也就是检索 Manifest 并下载对应 Image 中的各层 Blob。以下为步骤：

1. 获取仓库 Tokens 鉴权（本次项目不需要实现）
2. 获取 Image Manifest List
3. 若前两步获取的是一个多架构的 Manifest List，则需要：
   1. 解析 Manifests[] 并根据架构定位具体 digest
   2. 根据 digest 获取 Image Manifest
4. 下载前确认 Blob 是否存在。客户端应该向每一层发送 **HEAD** 请求
5. 根据前几步获取的 digest 从 manifest 中获取每一层的 Blob。客户端应发送 **GET** 请求

下面是官网给的 Pulling Images 的脚本示例。示例拉取 linux/amd64 平台上 ubuntu:latest 镜像：

```bash
#!/bin/bash

# Step 1: Get a bearer token
TOKEN=$(curl -s "https://auth.docker.io/token?service=registry.docker.io&scope=repository:library/ubuntu:pull" | jq -r .token)

# Step 2: Get the image manifest. In this example, an image manifest list is returned.
curl -s -H "Authorization: Bearer $TOKEN" \
     -H "Accept: application/vnd.docker.distribution.manifest.list.v2+json" \
     https://registry-1.docker.io/v2/library/ubuntu/manifests/latest \
     -o manifest-list.json

# Step 3a: Parse the `manifests[]` array to locate the digest for your target platform (e.g., `linux/amd64`).
IMAGE_MANIFEST_DIGEST=$(jq -r '.manifests[] | select(.platform.architecture == "amd64" and .platform.os == "linux") | .digest' manifest-list.json)

# Step 3b: Get the platform-specific image manifest
curl -s -H "Authorization: Bearer $TOKEN" \
     -H "Accept: application/vnd.docker.distribution.manifest.v2+json" \
     https://registry-1.docker.io/v2/library/ubuntu/manifests/$IMAGE_MANIFEST_DIGEST \
     -o manifest.json

# Step 4: Send a HEAD request to check if the layer blob exists
DIGEST=$(jq -r '.layers[0].digest' manifest.json)
curl -I -H "Authorization: Bearer $TOKEN" \
     https://registry-1.docker.io/v2/library/ubuntu/blobs/$DIGEST

# Step 5: Download the layer blob
curl -L -H "Authorization: Bearer $TOKEN" \
     https://registry-1.docker.io/v2/library/ubuntu/blobs/$DIGEST
```

 

## Pushing Images

推送镜像，与拉取相对应的提交镜像的各层 blob（例如 config 和 layers），然后上传引用这些 blob 的 manifest。以下为步骤：

1. 获取仓库 Tokens 鉴权（本次项目不需要实现）

2. 使用 HEAD 请求确保对于每一个 blob digest，对应的 blob 都存在

3. 若不存在，则使用单体 PUT 请求上传 blob

   1. 使用 POST 初始化上传
   2. 使用 PUT 完成上传

   > [!NOTE]
   >
   > 也可以使用分块上传的方式上传一个大型对象或者恢复终端上传。具体操作是使用 PATCH 请求发送每一个数据块，最后用 PUT 请求完成上传。

4. 使用 PUT 请求上传 Image Manifest 以关联 config 和 layers

下面是官网给的 Pushing Images 的脚本示例。示例推送了一个空的 blob 和 manifest 到 Docker Hub 上。

```bash
#!/bin/bash

USERNAME=yourusername
PASSWORD=dckr_pat
REPO=yourusername/helloworld
TAG=latest
CONFIG=config.json
MIME_TYPE=application/vnd.docker.container.image.v1+json

# Step 1: Get a bearer token
TOKEN=$(curl -s -u "$USERNAME:$PASSWORD" \
"https://auth.docker.io/token?service=registry.docker.io&scope=repository:$REPO:push,pull" \
| jq -r .token)

# Create a dummy config blob and compute its digest
echo '{"architecture":"amd64","os":"linux","config":{},"rootfs":{"type":"layers","diff_ids":[]}}' > $CONFIG
DIGEST="sha256:$(sha256sum $CONFIG | awk '{print $1}')"

# Step 2: Check if the blob exists
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -I \
  -H "Authorization: Bearer $TOKEN" \
  https://registry-1.docker.io/v2/$REPO/blobs/$DIGEST)

if [ "$STATUS" != "200" ]; then
  # Step 3: Upload blob using monolithic upload
  LOCATION=$(curl -sI -X POST \
    -H "Authorization: Bearer $TOKEN" \
    https://registry-1.docker.io/v2/$REPO/blobs/uploads/ \
    | grep -i Location | tr -d '\r' | awk '{print $2}')

  curl -s -X PUT "$LOCATION&digest=$DIGEST" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/octet-stream" \
    --data-binary @$CONFIG
fi

# Step 4: Upload the manifest that references the config blob
MANIFEST=$(cat <<EOF
{
  "schemaVersion": 2,
  "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
  "config": {
    "mediaType": "$MIME_TYPE",
    "size": $(stat -c%s $CONFIG),
    "digest": "$DIGEST"
  },
  "layers": []
}
EOF
)

curl -s -X PUT \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/vnd.docker.distribution.manifest.v2+json" \
  -d "$MANIFEST" \
  https://registry-1.docker.io/v2/$REPO/manifests/$TAG

echo "Pushed image to $REPO:$TAG"
```

该样例推送了一个不包含任何 layers 的镜像。如果要推送一个完整的有内容的镜像，需要对每个 layer 重复 2-3 次的推送并且把每个 layer 的 digest 保存在 manifest 的 layers[] 内。

## Deleting Images

删除镜像，包括通过 digest 删除 manifest。想要删除首先得获取 manifest digest，然后再使用 DELETE 请求删除 digest。

并非所有的 manifest 都可以删除。只有没有 tag 标记的 manifest 或者没有被其他 tag 或者 images 引用的 manifest 才能被删除（我猜是避免空指针问题）。如果删除的 manifest 是被引用的，返回 `403 Forbidden`。以下为步骤：

1. 获取仓库 Tokens 鉴权（本项目不需要实现）
2. 使用 Images Tag 获取 manifest
3. 从清单响应中获取 Docker-Content-Digest 标头。该摘要可唯一标识该清单
4. 使用 DELETE 请求，根据 digest 删除 manifest

下面是官网给的 Deleting Images 的脚本示例。示例删除了 Docker Hub 上 yourusername/helloworld 的 latest，不过似乎没有检查是否该 image 被其他 tag 引用。

```bash
#!/bin/bash

USERNAME=yourusername
PASSWORD=dckr_pat
REPO=yourusername/helloworld
TAG=latest

# Step 1: Get a bearer token
TOKEN=$(curl -s -u "$USERNAME:$PASSWORD" \
  "https://auth.docker.io/token?service=registry.docker.io&scope=repository:$REPO:pull,push,delete" \
  | jq -r .token)

# Step 2 and 3: Get the manifest and extract the digest from response headers
DIGEST=$(curl -sI -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/vnd.docker.distribution.manifest.v2+json" \
  https://registry-1.docker.io/v2/$REPO/manifests/$TAG \
  | grep -i Docker-Content-Digest | tr -d '\r' | awk '{print $2}')

echo "Deleting manifest with digest: $DIGEST"

# Step 4: Delete the manifest by digest
curl -s -X DELETE \
  -H "Authorization: Bearer $TOKEN" \
  https://registry-1.docker.io/v2/$REPO/manifests/$DIGEST

echo "Deleted image: $REPO@$DIGEST"
```

## API

### API 列表和响应码（省流版）

#### Manifest

##### Get image manifest

- 200 Manifest fetched successfully.
- 400 Invalid name or reference.
- 404 Respository or manifest not found.

##### Put image manifest

- 201 Manifest created successfully.
- 400 Invalid name, reference, or manifest.
- 404 Respository not found.

##### Check if manifest exists

- 200 Manifest exists.
- 404 Manifest not found.

##### Delete image manifest

- 202 Manifest deleted successfully. No content returned.
- 404 Manifest or respository not found.

#### Blob

##### Initiate blob upload or attempt cross-repository blob mount

- 201 Blob successfully mounted from another repository.
- 202 Upload initiated successfully (fallback if mount fails).
- 404 Repository not found.

##### Check existence of blob

- 200 Blob exists.
- 404 Blob not found.

##### Retrieve blob

- 200 Blob content returned directly.
- 404 Blob not found.

##### Get blob upload status

- 204 Upload in progress. No body returned.
- 404 Upload session not found.

##### Complete blob upload

- 201 Upliad completed successfully.
- 400 Invalid digest or missing parameters
- 404 Upload session not found.

##### Upload blob chunk

- 202 Chunk accepted and stored.
- 400 Malformed content or range.
- 404 Upload session not found.

##### Cancel blob upload

- 204 Upload session cancelled successfully. No body is returned.
- 404 Upload session not found.

### API 详细参数相应格式

#### Manifest

- GET image manifest
  由 name 和 reference 获取 manifest，其中 reference 可以是 tag 也可以是 digest
  获取一个标准的 Manifest 格式
  
  ```api
  GET /V2/{name}/manifests/{reference}
  ```

  - Request:![](https://pic.arctanp.top/PicGo/10d1d408-3d35-43e7-a855-e65376b84928.png)
  
  - Responses:
    ![](https://pic.arctanp.top/PicGo/20251001194824.png)
  
- PUT image manifest
  上传指定 tag 或者 digest 的 manifest 到 registry。
  Manifest 的 media type 为 `application/vnd.docker.distribution.manifest.v2+json`

  ```api
  PUT /v2/{name}/manifests/{reference}
  ```

  - Request:![](https://pic.arctanp.top/PicGo/20250923153715.png)

  - Response:
    ![](https://pic.arctanp.top/PicGo/20250923153803.png)

- Check if manifest exists
  通过 tag 或 digest 验证 manifest 是否存在。

  仅返回头信息（无正文）。适用于：

  - 检查特定镜像版本是否存在
  - 在下载或删除前确定清单的摘要或大小
  此接口需要具有拉取权限的身份验证。

  ```api
  HEAD /v2/{name}/manifests/{reference}
  ```

  - Requests:
    ![](https://pic.arctanp.top/PicGo/20250923154038.png)

  - Response:
    ![](https://pic.arctanp.top/PicGo/20251001194750.png)

- Delete image manifest
  通过摘要从存储库中删除映像清单，仅可删除未标记或未被引用的清单。若清单仍被标签或其他映像引用，注册表将返回403禁止访问状态码。
  
  ```api
  DELETE /v2/{name}/manifests/{reference}
  ```
  
  - Request:
    ![](https://pic.arctanp.top/PicGo/20250923154332.png)
  
  - Response:![](https://pic.arctanp.top/PicGo/20250923154411.png)

#### Blob

- Initiate blob upload or attempt cross-repository blob mount

  为存储库中的 blob（layer 或 config）启动上传会话。这是上传 blob 的第一步。该操作返回一个位置URL，用户可通过 PATCH（分块上传）或 PUT（整体上传）方式将块对象上传至该位置。

  客户端也可尝试挂载其他存储库中的 blob（需具备读取权限），通过添加 mount 和 from 查询参数实现。

  若挂载成功，注册表将返回 201 Created 状态码，blob 将被复用而无需重新上传。

  若挂载失败，上传将按常规流程进行并返回 202 Accepted 状态码。

  必须使用具有目标存储库推送权限的凭据进行身份验证。

  ```api
  POST /v2/{name}/blobs/uploads/
  ```

  - Request:
    ![](https://pic.arctanp.top/PicGo/20250923155044.png)

  - Response:
    ![](https://pic.arctanp.top/PicGo/20250923155121.png)

- Check existence of blob

  ```api
  HEAD /v2/{name}/blobs/{digest}
  ```

  检查 blob 是否已经存在

  - Request:

    ![](https://pic.arctanp.top/PicGo/20250923155456.png)

  - Response:
    ![](https://pic.arctanp.top/PicGo/20251001194850.png)

- Retrieve blob

  ```api
  GET /v2/{name}/blobs/{digest}
  ```

  - Request:
    ![](https://pic.arctanp.top/PicGo/20250923160743.png)
  - Responses:
    ![](https://pic.arctanp.top/PicGo/20250923160728.png)

- Get blob upload status

  ```api
  GET /v2/{name}/blobs/uploads/{uuid}
  ```

  检索正在进行的 blob 上传的当前状态。

  此功能适用于：恢复中断的上传；确定当前已接收的字节数；在分块上传中从正确偏移量重试；响应包含 Range 标头（指示当前接收的字节范围）和用于识别会话的 Docker-Upload-UUID。

  - Request:

    ![](https://pic.arctanp.top/PicGo/20250923161104.png)

  - Responses:
    ![](https://pic.arctanp.top/PicGo/20250923162739.png)

- Complete blob upload

  ```api
  PUT /v2/{name}/blobs/uploads/{uuid}
  ```

  此请求必须包含摘要查询参数，并可选地包含最后一个数据块。当存储注册表收到此请求时，它会验证摘要并存储 blob。

  此端点支持：整体上传（在此请求中上传整个 blob）或完成分块上传（最后一个数据块加摘要）

  - Request:
    ![](https://pic.arctanp.top/PicGo/20250923162651.png)
  - Responses:
    ![](https://pic.arctanp.top/PicGo/20250923162721.png)

- Upload blob chunk

  ```api
  PATCH /v2/{name}/blobs/uploads/{uuid}
  ```

  将块状数据分段上传至活跃的上传会话。

  此方法适用于分段上传，尤其适用于大型块状数据或恢复中断上传的情境。

  客户端通过PATCH请求发送二进制数据，可选添加Content-Range头部。

  每当分块被接受后，注册表将返回202 Accepted响应，包含：

  Range：当前存储的字节范围
  Docker-Upload-UUID：上传会话标识符
  Location：继续上传或通过PUT完成上传的URL

  - Request:
    ![](https://pic.arctanp.top/PicGo/20250923165640.png)
  - Response:
    ![](https://pic.arctanp.top/PicGo/20250923165722.png)

- Cancel blob upload

  ```api
  DELETE /v2/{name}/blobs/uploads/{uuid}
  ```

  取消正在进行的 blob 上传会话。

  此操作将丢弃已上传的所有数据并使上传会话失效。

  适用场景：

  上传失败或中途中止时
  客户端需要清理未使用的上传会话时
  取消后，UUID 将失效，必须重新发出 POST 请求才能重启上传。

  - Request:
    ![](https://pic.arctanp.top/PicGo/20250923170532.png)
  - Response:
    ![](https://pic.arctanp.top/PicGo/20250923170551.png)