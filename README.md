# sapicli

The command line upload tool of [sapicd/sapic](https://github.com/sapicd/sapic),
written in go language, supports Windows/macOS/Linux.

> English | [中文](README-cn.md)

## Install

### Docker

#### Download the image from Docker Hub

```bash
docker pull staugur/sapicli
```

#### Build from source code

```bash
git clone https://github.com/sapicd/cli sapicli && cd sapicli
docker build -t staugur/sapicli . # or: make docker
```

#### Usage

```bash
docker run --rm -ti staugur/sapicli
docker run --rm -ti staugur/sapicli -v
docker run --rm -ti staugur/sapicli -h
```

### Windows

In [github release](https://github.com/sapicd/cli/releases) page,
select the attachment under the distribution:
sapicli.{VERSION}-windows-amd64.zip, get sapicli.exe after decompression.

### Linux

As above, download the sapicli.{VERSION}-linux-amd64.tar.gz format package
and unzip it to get sapicli.

Or, you can also download from the command line:

```bash
version=0.6.0
wget -c https://github.com/sapicd/cli/releases/download/v${version}/sapicli.${version}-linux-amd64.tar.gz
tar zxf sapicli.${version}-linux-amd64.tar.gz
mv sapicli ~/bin/
sapicli -v
```

### macOS

As above, download the sapicli.{VERSION}-darwin-amd64.tar.gz format package
and unzip it to get sapicli, or with command line(Refer to Linux).

In addition to these two methods, you can use homebrew to install directly.

I've made [Tap](https://github.com/staugur/homebrew-tap),
On macOS or homebrew supported operating systems,
you can install the following:

```bash
brew tap staugur/tap
brew install sapicli
```

#### update

```bash
brew update
brew upgrade sapicli
```

#### uninstall

```bash
brew uninstall sapicli
brew untap staugur/tap
```

------

## Usage

```bash
# upload and print the json result
sapicli -u https://your.site -t <Token> screenshot.png

# upload and copy the markdown link to the clipboard
sapicli -u https://your.site -t <Token> -c md screenshot.png
```

```text
$ sapicli -h
usage: sapicli [-h] [-v] [-i] [-u SAPIC_URL] [-t SAPIC_TOKEN] [-a ALBUM]
                  [-d DESC] [-e EXPIRE] [-s STYLE] [-c {url,md,rst}]
                  [-A {auto,v1,v2}]
                  file [file ...]
```

| Option | Description |
| --- | --- |
| `-u, --sapic-url` | Site url, `http://` is optional, `/api/upload` (v1) or `/api/v2/upload` (v2) is appended automatically. Env: `sapicli_apiurl` |
| `-t, --sapic-token` | Token: a LinkToken for v1, a PAT (starts with `sp_`, needs the `upload` scope) for v2. Env: `sapicli_apitoken` |
| `-a, --album` | Album, its name for v1, its name or id for v2 |
| `-d, --desc` | Image title (description) |
| `-e, --expire` | Image expire in seconds, for temporary images |
| `-s, --style` | Output style: `default` / `typora` / `line` / `empty` / `<python module>` |
| `-c, --copy` | Copy the uploaded `url` / `md` / `rst` to the clipboard |
| `-A, --api-version` | Api version: `auto` (default) / `v1` / `v2` |

### Both v1 and v2 apis are supported

| | v1 | v2 |
| --- | --- | --- |
| Upload endpoint | `POST /api/upload` | `POST /api/v2/upload` |
| Auth header | `Authorization: LinkToken <LinkToken>` | `Authorization: Bearer <PAT>` |
| Body | form data, the image is base64 encoded in the `picbed` field | `multipart/form-data`, the image is in the `file` field |
| Album | `album` (name) | `album_id` (a name is looked up and resolved automatically) |

`auto` detection order: explicit `-A` > an `/api/v2` or `/api/upload` path already in `-u` > site probe (`GET /api/v2/site`) > the `sp_` token prefix > v1 by default.

So v1 sites keep the old `-u site -t LinkToken` usage; for v2 sites just replace `-t` with a PAT created in "Settings → Access tokens", the other options and the output styles stay the same. The v2 `src` (`data.image.src`) is normalized into the v1-shaped output, so `-c` copying and custom style modules work unchanged.

------

## Documentation

Doc to [sapic/cli](https://sapic.readthedocs.io/zh-cn/latest/cli.html)
