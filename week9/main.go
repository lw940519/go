package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
)

// FixLength 固定长度
const FixLength int = 16

const DelimiterSingle byte = 'P'

//const DelimiterMulti string = "PH"

// FieldLength 包长度字段所占字节数
const FieldLength int = 4

// goim 包相关字段
const (
	MaxBodySize    = int32(1 << 12)
	_packSize      = 4
	_headerSize    = 2
	_verSize       = 2
	_opSize        = 4
	_seqSize       = 4
	_rawHeaderSize = _packSize + _headerSize + _verSize + _opSize + _seqSize
	_maxPackSize   = MaxBodySize + int32(_rawHeaderSize)
	// offset
	_packOffset   = 0
	_headerOffset = _packOffset + _packSize
	_verOffset    = _headerOffset + _headerSize
	_opOffset     = _verOffset + _verSize
	_seqOffset    = _opOffset + _opSize
)

func main() {
	var handle int
	flag.IntVar(&handle, "handle", 0, "拆包类型")
	flag.Parse()
	fmt.Println(handle)

	handle = 0

	listen, err := net.Listen("tcp", "127.0.0.1:12003")
	if err != nil {
		log.Fatalf("listen error:%v\n", err)
	}
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Printf("accept error.%v\n", err)
			continue
		}

		// 连接自治
		switch handle {
		case 0:
			go fixLengthHandleConn(conn) //固定长度粘包
		case 1:
			go fixDelimiterHandleConn(conn) //固定分隔符粘包
		case 2:
			go lengthFieldHandleConn(conn) // 包长度字段
		case 3:
			go goimHandleConn(conn) // goim协议解码器
		}

	}
}

func fixLengthHandleConn(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, FixLength) // 新增一个缓冲区
	// 读写缓冲区
	rd := bufio.NewReader(conn)
	wr := bufio.NewWriter(conn)
	for {
		_, err := io.ReadFull(rd, buf)
		if err != nil {
			log.Printf("fixLengthHandleConn io readFull err:%v\n", err)
			continue
		}
		b := make([]byte, len(buf)) //拷贝也业务层去做处理
		copy(b, buf)
		go func() {
			// 模拟应答-原包返回
			wr.Write(b)
			wr.Flush() // 直接syscall
		}()
	}
}

func fixDelimiterHandleConn(conn net.Conn) {
	defer conn.Close()
	// 读写缓冲区
	rd := bufio.NewReader(conn)
	wr := bufio.NewWriter(conn)
	for {
		p, err := rd.ReadSlice(DelimiterSingle) // 依赖的是'P' 0x50
		if err != nil {
			log.Printf("fixDelimiterHandleConn rd.ReadLine err:%v\n", err)
			break
		}
		go func() {
			// 模拟应答-原包返回
			wr.Write(p)
			wr.Flush() // 直接syscall
		}()
	}
}

func lengthFieldHandleConn(conn net.Conn) {
	defer conn.Close()

	l := make([]byte, FieldLength) // 包长字段
	// 读写缓冲区
	rd := bufio.NewReader(conn)
	wr := bufio.NewWriter(conn)
	for {
		_, err := io.ReadFull(rd, l)
		if err != nil {
			log.Printf("lengthFieldHandleConn io readFull err:%v\n", err)
			continue
		}
		bodyLength, err := BytesToInt32(l, binary.BigEndian)
		if err != nil {
			log.Printf("lengthFieldHandleConn BytesToInt err:%v\n", err)
			continue
		}
		body := make([]byte, int(bodyLength)-FieldLength)
		_, err = io.ReadFull(rd, body)
		if err != nil {
			log.Printf("lengthFieldHandleConn io readFull err:%v\n", err)
			continue
		}
		go func() {
			wr.Write(body)
			wr.Flush() // 直接syscall
		}()
	}
}

func goimHandleConn(conn net.Conn) {
	defer conn.Close()
	p := make([]byte, _rawHeaderSize) // 头
	// 读写缓冲区
	rd := bufio.NewReader(conn)
	wr := bufio.NewWriter(conn)
	for {
		_, err := io.ReadFull(rd, p)
		if err != nil {
			log.Printf("goimHandleConn io readFull err:%v\n", err)
			break
		}
		pack := goimPackStruct{}
		packSize, err := BytesToInt32(p[_packOffset:_headerOffset], binary.BigEndian)
		if err != nil {
			log.Printf("BytesToInt32 packSize err:%v\n", err)
			continue
		}
		headerSize, err := BytesToInt16(p[_headerOffset:_verOffset], binary.BigEndian)
		if err != nil {
			log.Printf("BytesToInt16 headerSize err:%v\n", err)
			continue
		}
		pack.verSize, err = BytesToInt16(p[_verOffset:_opOffset], binary.BigEndian)
		if err != nil {
			log.Printf("BytesToInt16 verSize err:%v\n", err)
			continue
		}
		pack.opSize, err = BytesToInt32(p[_opOffset:_seqOffset], binary.BigEndian)
		if err != nil {
			log.Printf("BytesToInt32 opSize err:%v\n", err)
			continue
		}
		pack.seqSize, err = BytesToInt32(p[_seqOffset:], binary.BigEndian)
		if err != nil {
			log.Printf("BytesToInt32 seqSize err:%v\n", err)
			continue
		}

		if packSize > _maxPackSize {
			break
		}
		if headerSize != _rawHeaderSize {
			break
		}
		if bodyLen := int(packSize - int32(headerSize)); bodyLen > 0 {
			pack.body = make([]byte, bodyLen)
			_, err := io.ReadFull(rd, pack.body)
			if err != nil {
				log.Printf("goimHandleConn body io readFull err:%v\n", err)
				continue
			}
			// Todo 拿p去做相关业务做相关业务，考虑goroutine
			go func() {
				// 模拟应答-返回包体
				wr.Write((&pack).body)
				wr.Flush() // 直接syscall
			}()
		} else {
			pack.body = nil
		}
	}
}

type goimPackStruct struct {
	verSize int16
	opSize  int32
	seqSize int32
	body    []byte
}

func BytesToInt32(b []byte, order binary.ByteOrder) (int32, error) {
	bytesBuffer := bytes.NewBuffer(b)
	var x int32
	err := binary.Read(bytesBuffer, order, &x)
	if err != nil {
		return 0, err
	}
	return x, nil
}

func BytesToInt16(b []byte, order binary.ByteOrder) (int16, error) {
	bytesBuffer := bytes.NewBuffer(b)
	var x int16
	err := binary.Read(bytesBuffer, order, &x)
	if err != nil {
		return 0, err
	}
	return x, nil
}
