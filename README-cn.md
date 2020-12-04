# picbed-cli

[staugur/picbed](https://github.com/staugur/picbed)图床的命令行上传工具，
使用golang编写，支持Windows/macOS/Linux

> [English](README.md) | 中文

## 安装

### Windows

在[github release](https://github.com/staugur/picbed-cli/releases)中选择发行版下的附件：
picbed-cli.{VERSION}-windows-amd64.zip，解压后的picbed-cli.exe即程序。

### Linux

Linux操作系统如CentOS、Ubuntu等，除参考上述Windows方法外（下载
picbed-cli.{VERSION}-linux-amd64.tar.gz），还可以直接命令行下载：

```bash
version=0.4.1
wget -c https://static.saintic.com/download/picbed-cli/picbed-cli.${version}-linux-amd64.tar.gz
tar zxf picbed-cli.${version}-linux-amd64.tar.gz 
mv picbed-cli ~/bin/  # 移动到PATH目录下
picbed-cli -v
```

### macOS

参考Windows安装方法，下载解压picbed-cli.{VERSION}-darwin-amd64.tar.gz，或者参考Linux
命令行下载，除这两种方法外，还可以使用homebrew直接安装。

我已制作了[Tap](https://github.com/staugur/homebrew-tap)，
在macOS或支持homebrew的操作系统中可以这么安装：

```bash
brew tap staugur/tap
brew install picbed-cli
```

这么更新：

```bash
brew update
brew upgrade picbed-cli
```

这么卸载：

```bash
brew uninstall picbed-cli
brew untap staugur/tap
```

------

## 用法

Doc to https://picbed.rtfd.vip/cli.html
