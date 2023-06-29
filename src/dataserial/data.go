package dataserial

import (
	"bytes"
	"encoding/gob"
)
// gob 在进行序列化和反序列化时，它能够保留数据的类型信息；
// 这意味着可以序列化和反序列化包括结构体、切片、映射、接口等复杂数据类型。

// RPCdata represents the serializing format of structured data
type RPCdata struct {
	Name string        // name of the function
	// 空接口切片[]interface{} 可以容纳多种不同类型的值，例如data := []interface{}{42, "hello", true}
	Args []interface{} // request's or response's body expect error.
	Err  string        // Error any executing remote server
}

// Encode The RPCdata in binary format which can
// be sent over the network.
func Encode(data RPCdata) ([]byte, error) {
	var buf bytes.Buffer
	// 将RPCdata格式的数据进行序列化
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decode the binary data into the Go RPC struct
func Decode(b []byte) (RPCdata, error) {
	buf := bytes.NewBuffer(b)
	// 从字节信息中恢复原始RPCdata数据（反序列化）
	decoder := gob.NewDecoder(buf)
	var data RPCdata
	if err := decoder.Decode(&data); err != nil {
		return RPCdata{}, err
	}
	return data, nil
}