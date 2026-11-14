# sapicli

[sapicd/sapic](https://github.com/sapicd/sapic) 图床的命令行上传工具，
使用golang编写，支持Windows/macOS/Linux

> [English](README.md) | 中文

## 安装

### Docker

#### 从 Docker Hub 下载镜像

```bash
docker pull staugur/sapicli
```

#### 从 源码 构建

```bash
git clone https://github.com/sapicd/cli sapicli && cd sapicli
docker build -t staugur/sapicli . # or: make docker
```

#### 使用

```bash
docker run --rm -ti staugur/sapicli
docker run --rm -ti staugur/sapicli -v
docker run --rm -ti staugur/sapicli -h
```

### Windows

在[github release](https://github.com/sapic/cli/releases)中选择发行版下的附件：
sapicli.{VERSION}-windows-amd64.zip，解压后的sapicli.exe即程序。

### Linux

Linux操作系统如CentOS、Ubuntu等，除参考上述Windows方法外（下载
sapicli.{VERSION}-linux-amd64.tar.gz），还可以直接命令行下载：

```bash
version=0.6.0
wget -c https://github.com/sapicd/cli/releases/download/v${version}/sapicli.${version}-linux-amd64.tar.gz
tar zxf sapicli.${version}-linux-amd64.tar.gz
mv sapicli ~/bin/  # 移动到PATH目录下
sapicli -v
```

### macOS

参考Windows安装方法，下载解压sapicli.{VERSION}-darwin-amd64.tar.gz，或者参考Linux
命令行下载，除这两种方法外，还可以使用homebrew直接安装。

我已制作了[Tap](https://github.com/staugur/homebrew-tap)，
在macOS或支持homebrew的操作系统中可以这么安装：

```bash
brew tap staugur/tap
brew install sapicli
```

这么更新：

```bash
brew update
brew upgrade sapicli
```

这么卸载：

```bash
brew uninstall sapicli
brew untap staugur/tap
```

------

## 使用

```bash
# 上传并输出 JSON 结果
sapicli -u https://your.site -t <Token> screenshot.png

# 上传后把 Markdown 链接复制到剪贴板
sapicli -u https://your.site -t <Token> -c md screenshot.png
```

```text
$ sapicli -h
usage: sapicli [-h] [-v] [-i] [-u SAPIC_URL] [-t SAPIC_TOKEN] [-a ALBUM]
                  [-d DESC] [-e EXPIRE] [-s STYLE] [-c {url,md,rst}]
                  [-A {auto,v1,v2}]
                  file [file ...]
```

| 选项 | 说明 |
| --- | --- |
| `-u, --sapic-url` | 站点地址，可省略 `http://`，会自动补齐 `/api/upload`（v1）或 `/api/v2/upload`（v2）；也可用环境变量 `sapicli_apiurl` |
| `-t, --sapic-token` | 访问令牌：v1 用 LinkToken，v2 用 PAT（`sp_` 开头、需含 `upload` 权限）；也可用环境变量 `sapicli_apitoken` |
| `-a, --album` | 相册，v1 传相册名，v2 可传相册名或相册 ID |
| `-d, --desc` | 图片标题（描述） |
| `-e, --expire` | 有效期（秒），用于临时图片 |
| `-s, --style` | 输出风格：`default` / `typora` / `line` / `empty` / `<python 模块名>` |
| `-c, --copy` | 上传后把 `url` / `md` / `rst` 复制到剪贴板 |
| `-A, --api-version` | 接口版本：`auto`（默认）/ `v1` / `v2` |

### 同时支持 v1 与 v2 接口

| | v1 | v2 |
| --- | --- | --- |
| 上传接口 | `POST /api/upload` | `POST /api/v2/upload` |
| 认证请求头 | `Authorization: LinkToken <LinkToken>` | `Authorization: Bearer <PAT>` |
| 请求体 | 表单，图片以 base64 放在 `picbed` 字段 | `multipart/form-data`，图片放在 `file` 字段 |
| 相册参数 | `album`（相册名） | `album_id`（传相册名时自动查询成 ID） |

`auto` 的判定顺序：`-A` 显式指定 > `-u` 中已带的 `/api/v2` 或 `/api/upload` > 站点探测（`GET /api/v2/site`）> 令牌前缀 `sp_` > 默认 v1。

所以 v1 站点沿用原来的 `-u 站点 -t LinkToken`；v2 站点把 `-t` 换成网页「设置 → 访问令牌」创建的 PAT 即可，其它参数与输出风格不变。V2 上传成功返回的 `src`（`data.image.src`）会被规整成与 v1 一致的输出结构，`-c` 复制与自定义风格模块都无需修改。

------

## 文档

Doc to [sapic/cli](https://sapic.rtfd.vip/cli.html)
