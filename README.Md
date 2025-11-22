# Jot

一个极简的网页笔记应用，支持 Markdown 编辑、实时预览和自动备份。

## 特性

- ✨ **极简设计** - 清爽的界面，专注于内容
- 📝 **Markdown 支持** - 实时预览 Markdown 内容
- 🔄 **实时同步** - WebSocket 支持多设备实时同步
- 💾 **自动备份** - 7 天未修改的笔记自动移动到备份文件夹
- 🔐 **管理后台** - Token 认证的管理界面，查看所有笔记和备份
- 🚀 **快速启动** - 可配置长度的随机笔记名称，快速创建和分享
- ⚙️ **高度可配置** - 支持命令行参数、环境变量、.env 文件和 config.json 配置
- 📦 **轻量级** - Go 语言实现，单文件部署
- 📤 **文件上传** - 支持拖拽上传和点击上传，图片自动显示，其他文件显示为下载链接
- 👁️ **只读模式** - 通过 `/read/{noteName}` 访问只读模式，适合分享和查看
- 💾 **配置持久化** - 配置修改后自动保存到 `config.json`，重启后自动加载

## 快速开始

### 方式一：Docker（推荐）

```bash
# 从 GitHub Container Registry 拉取镜像
docker pull ghcr.io/hello--world/jot:latest

# 运行容器
docker run -d \
  --name jot \
  -p 8080:8080 \
  -e ADMIN_TOKEN=your-secret-token \
  ghcr.io/hello--world/jot:latest
```

### 方式二：从源码编译

```bash
# 克隆项目
git clone <repository-url>
cd jot

# 安装依赖
go mod download

# 编译
go build -o jot .
```

### 方式三：下载预编译二进制

从 [GitHub Releases](https://github.com/hello--world/jot/releases) 下载对应平台的二进制文件。

### 运行

```bash
# 使用命令行参数
./jot -token your-secret-token

# 或使用环境变量
export ADMIN_TOKEN=your-secret-token
./jot

# 或使用 .env 文件
echo "ADMIN_TOKEN=your-secret-token" > .env
./jot
```

### 配置选项

所有配置项都支持多种方式设置（优先级从高到低）：
1. 命令行参数
2. 环境变量
3. `.env` 文件
4. `config.json` 文件（如果存在，会优先使用配置文件，忽略环境变量和命令行参数，但端口和 token 除外）

**注意**：如果 `config.json` 文件存在，程序会优先使用配置文件中的值，环境变量和命令行参数（除端口和 token 外）将被忽略。配置文件可以通过管理后台动态修改。

#### 必需配置

- `-token` / `ADMIN_TOKEN`: 管理后台访问令牌（必需）

#### 可选配置

- `-port` / `PORT`: 服务器端口（默认: `:8080`）
  - 可以只指定端口号（如 `8080`），会自动添加 `:` 前缀
  - 也可以完整指定（如 `:8080` 或 `:3000`）

- `-access-token` / `ACCESS_TOKEN`: 访问令牌（可选）
  - 如果设置，所有笔记访问都需要提供此令牌（`/read` 路径除外）
  - 可以通过 URL 参数 `?token=xxx`、Cookie `access_token` 或 `Authorization: Bearer xxx` header 提供
  - 留空表示无需授权即可访问笔记

- `-admin-path` / `ADMIN_PATH`: 管理后台路径（默认: `/admin`）

- `-note-name-len` / `NOTE_NAME_LEN`: 笔记名称最小长度（默认: `3`）
  - 这是最小值，当生成的名称已存在时，会自动增加到 4 位或更长
  - 如果使用多了，超过了可用数量，会自动使用 4 位或更多位

- `-backup-days` / `BACKUP_DAYS`: 备份天数（默认: `7`）
  - 超过指定天数未修改的笔记会自动移动到备份文件夹

- `-note-chars` / `NOTE_CHARS`: 随机字符串字符集（默认: `0123456789abcdefghijklmnopqrstuvwxyz`）
  - 用于生成笔记名称的字符集合
  - 可以自定义字符集，例如只使用数字：`0123456789`
  - 或使用大小写字母和数字：`0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ`

- `-max-file-size` / `MAX_FILE_SIZE`: 最大文件大小（默认: `10MB`）
  - 限制单个笔记文件的最大大小
  - 支持格式：`10M`, `100MB`, `1G`, `500KB` 等
  - 单位：B (字节), K/KB (千字节), M/MB (兆字节), G/GB (千兆字节), T/TB (太字节)

- `-max-path-length` / `MAX_PATH_LENGTH`: 最大路径/笔记名称长度（默认: `20`）
  - 限制笔记名称的最大字符数
  - 防止过长的路径名称

- `-max-total-size` / `MAX_TOTAL_SIZE`: 最大总文件大小（默认: `500MB`）
  - 限制所有笔记和上传文件的总大小（不包括备份目录）
  - 支持格式：`500M`, `1GB`, `100MB` 等
  - 可在管理后台动态修改

- `-max-note-count` / `MAX_NOTE_COUNT`: 最大笔记数量（默认: `500`）
  - 限制可以创建的笔记总数
  - 可在管理后台动态修改

### 使用 .env 文件

创建 `.env` 文件：

```env
# 必需配置
ADMIN_TOKEN=your-secret-token

# 可选配置
PORT=8080
ACCESS_TOKEN=your-access-token  # 可选，设置后所有笔记访问都需要此令牌（/read 路径除外）
ADMIN_PATH=/admin
NOTE_NAME_LEN=3
BACKUP_DAYS=7
NOTE_CHARS=0123456789abcdefghijklmnopqrstuvwxyz
MAX_FILE_SIZE=10MB
MAX_PATH_LENGTH=20
MAX_TOTAL_SIZE=500MB
MAX_NOTE_COUNT=500
```

### 配置文件 (config.json)

程序首次运行后会自动创建 `config.json` 文件，保存所有配置项。如果配置文件存在，程序会优先使用配置文件中的值。

配置文件格式：

```json
{
  "adminToken": "your-secret-token",
  "accessToken": "your-access-token",
  "adminPath": "/admin",
  "noteNameLen": 3,
  "backupDays": 7,
  "noteChars": "0123456789abcdefghijklmnopqrstuvwxyz",
  "maxFileSize": 10485760,
  "maxPathLength": 20,
  "maxTotalSize": 524288000,
  "maxNoteCount": 500
}
```

**注意**：
- 配置文件中的大小值以字节为单位
- 可以通过管理后台修改配置，修改后会自动保存到 `config.json`
- 如果配置文件存在，环境变量和命令行参数（除端口和 token 外）将被忽略

### 配置示例

```bash
# 使用命令行参数
./jot -token my-token -port 3000 -note-name-len 4 -backup-days 14 -max-file-size 20MB -max-path-length 30

# 使用环境变量
export ADMIN_TOKEN=my-token
export PORT=3000
export NOTE_NAME_LEN=4
export BACKUP_DAYS=14
export NOTE_CHARS=0123456789abcdefghijklmnopqrstuvwxyz
export MAX_FILE_SIZE=20MB
export MAX_PATH_LENGTH=30
./jot

# 混合使用（命令行参数优先级最高）
export ADMIN_TOKEN=my-token
./jot -port 3000 -note-name-len 5 -max-file-size 50MB
```

## 使用方式

### Web 界面

1. 访问 `http://localhost:8080` 会自动创建随机笔记
   - 如果设置了访问令牌（access_token），需要先输入令牌才能访问
2. 访问 `http://localhost:8080/笔记名称` 编辑指定笔记
3. 访问 `http://localhost:8080/read/笔记名称` 以只读模式查看笔记（不支持编辑）
   - `/read` 路径不需要访问令牌，但需要笔记的锁令牌（如果笔记有锁）
   - 使用 `?raw=1` 参数可以下载原始文件内容（纯文本）
4. 访问 `http://localhost:8080/admin` 进入管理后台
   - 首次访问需要输入管理员令牌（admin token）
   - 登录后使用 session token（HttpOnly cookie），不会暴露在 URL 中

### 文件上传

- 点击左下角的"📤 上传"按钮打开上传窗口
- 支持拖拽文件到上传窗口
- 支持点击"选择文件"按钮选择文件
- 上传的图片会自动以 Markdown 图片格式显示：`![文件名](url)`
- 上传的其他文件会显示为下载链接：`[下载 文件名](url)`
- 上传的文件会自动插入到当前光标位置

### 命令行

```bash
# 创建/更新笔记（如果设置了访问令牌，需要提供 token）
curl http://localhost:8080/abc -d "笔记内容" -H "Authorization: Bearer your-access-token"
# 或使用 URL 参数（不推荐，会暴露在日志中）
curl "http://localhost:8080/abc?token=your-access-token" -d "笔记内容"

# 获取笔记内容
curl http://localhost:8080/abc -H "Authorization: Bearer your-access-token"

# 获取原始内容（纯文本）
curl "http://localhost:8080/abc?raw=1" -H "Authorization: Bearer your-access-token"

# 访问只读模式（不需要访问令牌）
curl http://localhost:8080/read/abc

# 下载原始文件内容（纯文本，不需要访问令牌）
curl "http://localhost:8080/read/abc?raw=1"

# 访问有锁的笔记（需要提供 lock_token）
curl "http://localhost:8080/read/abc?raw=1&lock_token=your-lock-token"

# 获取 JSON 格式的笔记列表（需要管理员 token）
curl http://localhost:8080/admin -H "Authorization: Bearer your-admin-token" -H "Accept: application/json"

# 上传文件（如果设置了访问令牌，需要提供 token）
curl -F "file=@image.png" http://localhost:8080/api/upload -H "Authorization: Bearer your-access-token"
```

## 功能说明

### 笔记管理

- **自动创建**: 访问不存在的笔记会自动创建
- **实时保存**: 输入内容自动保存（延迟 500ms）
- **实时预览**: Markdown 内容实时渲染
- **文件信息**: 顶部显示文件大小、创建时间和修改时间
- **只读模式**: 通过 `/read/{noteName}` 访问只读模式，适合分享和查看，不支持编辑
  - `/read` 路径不需要访问令牌（access_token），但需要笔记的锁令牌（如果笔记有锁）
  - 使用 `?raw=1` 参数可以下载原始文件内容（纯文本，不带锁标记）
  - 使用 `?raw=1&lock_token=xxx` 可以下载有锁笔记的原始内容（需要正确的锁令牌）
- **笔记锁**: 可以为笔记设置锁令牌，只有提供正确的锁令牌才能查看或编辑
  - 在编辑页面点击"设置锁"按钮可以设置锁令牌
  - 有锁的笔记在访问时需要提供 `lock_token` 参数或通过 Cookie/Authorization header
  - 下载原始内容时，如果锁令牌正确，会自动去掉锁标记 `<!-- LOCK:token -->`
- **文件上传**: 支持上传图片和其他文件，图片自动显示，其他文件显示为下载链接

### 备份功能

- 超过指定天数（默认 7 天，可通过 `BACKUP_DAYS` 配置）未修改的笔记会自动移动到 `bak/YYYYMMDD/` 目录
- 备份按日期组织，便于管理
- 管理后台可以查看所有备份笔记

### 笔记名称生成

- 笔记名称使用随机字符串生成，默认最小长度为 3 位
- 字符集默认为小写字母和数字（`0123456789abcdefghijklmnopqrstuvwxyz`）
- **自动扩展机制**：当生成的名称已存在时，会自动增加长度
  - 如果最小长度为 3，当名称冲突时会自动使用 4 位
  - 如果 4 位也冲突，会继续增加到 5 位、6 位等
  - 确保即使笔记数量很多，也能生成唯一的名称

### 管理后台

- **Session 认证保护**：使用 HttpOnly cookie 存储 session token，不会暴露在 URL 中
  - 首次访问需要输入管理员令牌（admin token）
  - 登录后创建 session token，有效期 2 小时
  - 管理员令牌不会存储到浏览器 localStorage
- 查看所有活跃笔记和备份笔记
- 显示笔记统计信息（数量、大小）
- 支持标签切换查看活跃/备份笔记
- **动态配置管理**：可以在管理后台修改以下配置项，修改后自动保存到 `config.json`：
  - 访问令牌（access_token）- 用于控制笔记访问权限
  - 管理后台路径
  - 笔记名称最小长度
  - 备份天数
  - 随机字符串字符集
  - 最大文件大小
  - 最大路径长度
  - 最大总文件大小限制
  - 最大笔记数量限制

## 项目结构

```
jot/
├── main.go          # 主程序文件
├── go.mod           # Go 模块定义
├── _tmp/            # 笔记存储目录
├── bak/             # 备份目录（按日期组织）
│   └── YYYYMMDD/    # 日期目录
│       └── note_name # 备份笔记
├── uploads/         # 上传文件存储目录
│   └── filename     # 上传的文件
├── config.json      # 配置文件（自动生成，保存所有配置项）
└── .env             # 环境变量配置文件（可选）
```

## 技术栈

- **Go 1.21+** - 后端语言
- **Gorilla Mux** - HTTP 路由
- **Gorilla WebSocket** - WebSocket 支持
- **Blackfriday** - Markdown 渲染

## Docker 部署

### 使用 Docker 运行

```bash
# 构建镜像
docker build -t jot .

# 运行容器
docker run -d \
  --name jot \
  -p 8080:8080 \
  -e ADMIN_TOKEN=your-secret-token \
  -v $(pwd)/_tmp:/root/_tmp \
  -v $(pwd)/bak:/root/bak \
  -v $(pwd)/uploads:/root/uploads \
  -v $(pwd)/config.json:/root/config.json \
  jot

# 或使用 docker-compose
```

### Docker Compose 示例

创建 `docker-compose.yml`：

```yaml
version: '3.8'

services:
  jot:
    build: .
    container_name: jot
    ports:
      - "8080:8080"
    environment:
      - ADMIN_TOKEN=your-secret-token
      - PORT=8080
    volumes:
      - ./_tmp:/root/_tmp
      - ./bak:/root/bak
      - ./uploads:/root/uploads
      - ./config.json:/root/config.json
    restart: unless-stopped
```

运行：

```bash
docker-compose up -d
```

### 从 GitHub Container Registry 拉取

```bash
# 拉取最新版本
docker pull ghcr.io/hello--world/jot:latest

# 运行
docker run -d \
  --name jot \
  -p 8080:8080 \
  -e ADMIN_TOKEN=your-secret-token \
  ghcr.io/hello--world/jot:latest
```

**注意**：如果镜像是私有的，需要先登录：
```bash
echo $GITHUB_TOKEN | docker login ghcr.io -u hello--world --password-stdin
```

## 开发

```bash
# 运行开发服务器
go run main.go -token dev-token

# 编译
go build -o jot .

# 测试
go test ./...
```

## CI/CD

项目使用 GitHub Actions 进行自动构建和发布：

- **自动构建**：推送 tag（格式：`v*`）时自动构建多平台二进制文件
- **自动发布**：构建完成后自动创建 GitHub Release
- **Docker 镜像**：自动构建并推送 Docker 镜像到 Docker Hub

### 发布新版本

```bash
# 创建并推送 tag
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0
```

GitHub Actions 会自动：
1. 构建 Linux、Windows、macOS 的二进制文件（amd64、arm64）
2. 创建 GitHub Release 并上传所有构建产物
3. 构建并推送 Docker 镜像到 GitHub Container Registry (ghcr.io)

**注意**：使用 GitHub Container Registry 不需要额外的 Docker Hub 账号，使用 `GITHUB_TOKEN` 即可自动认证。



