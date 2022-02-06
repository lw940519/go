#### 一、总结几种 socket 粘包的解包方式：fix length/delimiter based/length field based frame decoder

1、包长度固定不变。
+ 处理思路：
    + 1 确定包固定长度为a
    + 2 循环获取固定长度的byte,可以交给goroutine进行处理。
2、固定分隔符。
+ 特别注意：处理固定分隔符外，包内不要出现该字符，如若冲突可以做其它字符规则替换
+ 换行符可以使用bufio.ReadLine
+ 其它字符可以使用bufio.ReadSlicebu或者fio.Scanner
+ 处理思路：
    + 1 确定固定分隔符（包尾）
    + 2 引入变量var a []byte 存储包的前半部分（也就是不包含分割符）
    + 3 取出字节并拼接到a后面,利用bufio.Scanner 进行拆包。
    + 4 循环处理包。
        + 完整包直接进行业务处理,不完整包放入a。
3、包头添加表示包长度的字段。
+ 处理思路：
    + 1 确定包长度字段的长度为n字节
    + 2 获取包字段长度y的数据
    + 3 得到包长度x
    + 4 获取包体长度：x-y
    + 5 交给goroutine处理。


### 二、实现一个从 socket connection 中解码出 goim 协议的解码器。

1 goim协议定义地址：https://github.com/Terry-Mao/goim/blob/e742c99ad76e626d5f6df8b33bc47ca005501980/api/protocol/protocol.go
+ 结构如下：
1. 包头
+ 包头结构

|字段| 含义  | 大小（字节） |
|---|-----|----|
|_packSize| 包长度 | 4|
|_headerSize| 头长度 | 2  |
|_verSize| 版本号 | 2  |
|_opSize| 包类别 | 4  |
|_seqSize| 序列码 | 4  |
+ 包头长度：_packSize + _headerSize + _verSize + _opSize + _seqSize = 4+2+2+4+4 = 16
2. 包体
+ 特殊包体：心跳 4字节
3. 其它规定
+ 最大包体长度：MaxBodySize = int32(1 << 12)

2、 解析思路
+ 先根据包头长，获取包头
+ 然后根据包长-包头长得到包体长
+ 获取包体
+ 根据包类别进行路由解析
