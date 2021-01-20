# sapicli

The command line upload tool of [sapicd/sapic](https://github.com/sapicd/sapic),
written in go language, supports Windows/macOS/Linux.

> English | [中文](README-cn.md)

## Install

### Windows

In [github release](https://github.com/sapicd/cli/releases) page,
select the attachment under the distribution:
sapicli.{VERSION}-windows-amd64.zip, get sapicli.exe after decompression.

### Linux

As above, download the sapicli.{VERSION}-linux-amd64.tar.gz format package
and unzip it to get sapicli.

Or, you can also download from the command line:

```bash
version=0.5.0
wget -c https://github.com/sapicd/cli/releases/download/${version}/sapicli.${version}-linux-amd64.tar.gz
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

Doc to https://picbed.rtfd.vip/cli.html
