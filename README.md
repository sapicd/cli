# picbed-cli

The command line upload tool of [staugur/picbed](https://github.com/staugur/picbed),
written in go language, supports Windows/macOS/Linux.

> English | [中文](README-cn.md)

## Install

### Windows

In [github release](https://github.com/staugur/picbed-cli/releases) page,
select the attachment under the distribution:
picbed-cli.{VERSION}-windows-amd64.zip, get picbed-cli.exe after decompression.

### Linux

As above, download the picbed-cli.{VERSION}-linux-amd64.tar.gz format package
and unzip it to get picbed-cli.

Or, you can also download from the command line:

```bash
version=0.4.1
wget -c https://github.com/staugur/picbed-cli/releases/download/${version}/picbed-cli.${version}-linux-amd64.tar.gz
tar zxf picbed-cli.${version}-linux-amd64.tar.gz 
mv picbed-cli ~/bin/
picbed-cli -v
```

### macOS

As above, download the picbed-cli.{VERSION}-darwin-amd64.tar.gz format package
and unzip it to get picbed-cli, or with command line(Refer to Linux).

In addition to these two methods, you can use homebrew to install directly.

I've made [Tap](https://github.com/staugur/homebrew-tap),
On macOS or homebrew supported operating systems,
you can install the following:

```bash
brew tap staugur/tap
brew install picbed-cli
```

#### update

```bash
brew update
brew upgrade picbed-cli
```

#### uninstall

```bash
brew uninstall picbed-cli
brew untap staugur/tap
```

------

## Usage

Doc to https://picbed.rtfd.vip/cli.html
