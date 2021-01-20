# sapicli

[sapicd/sapic](https://github.com/sapicd/sapic)图床的命令行上传工具，
使用golang编写，支持Windows/macOS/Linux

> [English](README.md) | 中文

## 安装

### Windows

在[github release](https://github.com/sapic/cli/releases)中选择发行版下的附件：
sapicli.{VERSION}-windows-amd64.zip，解压后的sapicli.exe即程序。

### Linux

Linux操作系统如CentOS、Ubuntu等，除参考上述Windows方法外（下载
sapicli.{VERSION}-linux-amd64.tar.gz），还可以直接命令行下载：

```bash
version=0.4.1
wget -c https://static.saintic.com/download/picbed-cli/sapicli.${version}-linux-amd64.tar.gz
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

## 用法

Doc to https://picbed.rtfd.vip/cli.html
