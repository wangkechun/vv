# vv 编辑器
![vv logo](https://raw.githubusercontent.com/wangkechun/vv/master/doc/icon.png)

[![Build Status](https://travis-ci.org/wangkechun/vv.svg?branch=master)](https://travis-ci.org/wangkechun/vv)
[![Go Report Card](https://goreportcard.com/badge/github.com/wangkechun/vv)](https://goreportcard.com/report/github.com/wangkechun/vv)

## 项目简介：

该项目旨在为经常在服务器、虚拟机、docker容器中修改代码或者配置文件的小伙伴提供本地coding新体验。

你是否一直为修改服务器代码而困扰？无代码高亮，少之又少的快捷方式，一个不注意又写了一个莫名其妙的bug。

我不是资深的vim使用者，我只是一个追求完美和效率的偷懒者。

vv编辑器的想法就是我在偷懒的时候萌生的。

vv+ ‘filename’，你盯着的那个文件已经在你最喜欢的编辑器中打开

'command + s', 服务器已如你期望自动更新完毕

这种开发方式，相信你再也不会拒绝。

## 主要用到的开源库

- google.golang.org/grpc grpc、protobuf
- github.com/kr/binarydist  bsdiff 的 go 语言实现
- github.com/spf13/cobra CLI 界面

## p2p 数据传输实现
测试过 https://github.com/vzex/dog-tunnel 这种 udp 打洞的方式，不过实测下来稳定性不够，某些网络不能成功连接。

于是改为尝试 frp 这种中间服务器中转的模式。

A -> 中转服务器

B -> 中转服务器

中转服务器 pipe 配对的 tcp， A -> B 成功建立连接。
这里存在一个问题是 A、B 和中转服务器的的时间间隔可能很长，等到成功配对的时候可能有 tcp 已经断掉了，建立的连接是坏的。
并且中转服务器不能私自往 tcp 里面写数据，所以没法提前检测连接是否断掉了。

最后改成 A -> 中转服务器 建立长连接， 然后中转服务器如果收到 B 的连接，就告诉 A 需要新起一个连接， 然后把 A 新起的连接和 B 的连接配置。
由于两个连接都是刚刚建立的，不会有超时问题。

这种方式的话问题是会慢点，因为要通知 A 去建立一个新连接， 解决方案很简单，A 需要新起连接的时候多发起一个，之后的连接就可以复用了。 如果多余的连接长时间等不到配对会主动断开。

A 和 B 配对成功了就可以随意通信了， 然后直接使用 grpc， 可以很方便的双向通信。

虽然 grpc 是需要一些hack的，正常是一个 server 负责 listen，一个 client 负责 dial。 这里实际上两边都是 dial， 需要自定义一个 listen， 如果成功建立一条到 B 的连接，则 Accept 一个连接给 grpc。 

