# Jot

一个极简的网页笔记应用，支持 Markdown 编辑、实时预览和自动备份。

## 特性

- ✨ **极简设计** - 清爽的界面，专注于内容
- 📝 **Markdown 支持** - 实时预览 Markdown 内容
- 🔄 **实时同步** - WebSocket 支持多设备实时同步
- 💾 **自动备份** - 7 天未修改的笔记自动移动到备份文件夹
- 🔐 **管理后台** - Token 认证的管理界面，查看所有笔记和备份
- 🚀 **快速启动** - 3 位随机笔记名称，快速创建和分享
- 📦 **轻量级** - Go 语言实现，单文件部署

## 快速开始

### 安装

```bash
# 克隆项目
git clone <repository-url>
cd jot

# 安装依赖
go mod download

# 编译
go build -o jot .
```

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

- `-token`: 管理后台访问令牌（必需）
- `-admin-path`: 管理后台路径（默认: `/admin`）
- `ADMIN_TOKEN`: 环境变量方式设置令牌
- `ADMIN_PATH`: 环境变量方式设置管理路径

### 使用 .env 文件

创建 `.env` 文件：

```env
ADMIN_TOKEN=your-secret-token
ADMIN_PATH=/admin
```

## 使用方式

### Web 界面

1. 访问 `http://localhost:8080` 会自动创建随机笔记
2. 访问 `http://localhost:8080/笔记名称` 编辑指定笔记
3. 访问 `http://localhost:8080/admin?token=your-token` 进入管理后台

### 命令行

```bash
# 创建/更新笔记
curl http://localhost:8080/abc -d "笔记内容"

# 获取笔记内容
curl http://localhost:8080/abc

# 获取原始内容（纯文本）
curl http://localhost:8080/abc?raw

# 获取 JSON 格式的笔记列表（需要 token）
curl http://localhost:8080/admin?token=your-token -H "Accept: application/json"
```

## 功能说明

### 笔记管理

- **自动创建**: 访问不存在的笔记会自动创建
- **实时保存**: 输入内容自动保存（延迟 500ms）
- **实时预览**: Markdown 内容实时渲染
- **文件信息**: 顶部显示文件大小、创建时间和修改时间

### 备份功能

- 超过 7 天未修改的笔记会自动移动到 `bak/YYYYMMDD/` 目录
- 备份按日期组织，便于管理
- 管理后台可以查看所有备份笔记

### 管理后台

- Token 认证保护
- 查看所有活跃笔记和备份笔记
- 显示笔记统计信息（数量、大小）
- 支持标签切换查看活跃/备份笔记

## 项目结构

```
jot/
├── main.go          # 主程序文件
├── go.mod           # Go 模块定义
├── _tmp/            # 笔记存储目录
├── bak/             # 备份目录（按日期组织）
│   └── YYYYMMDD/    # 日期目录
│       └── note_name # 备份笔记
└── .env             # 配置文件（可选）
```

## 技术栈

- **Go 1.21+** - 后端语言
- **Gorilla Mux** - HTTP 路由
- **Gorilla WebSocket** - WebSocket 支持
- **Blackfriday** - Markdown 渲染

## 开发

```bash
# 运行开发服务器
go run main.go -token dev-token

# 编译
go build -o jot .

# 测试
go test ./...
```



