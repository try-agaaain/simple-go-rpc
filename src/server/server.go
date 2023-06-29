package server

import (
	"fmt"
	"github.com/ankur-anand/simple-go-rpc/src/dataserial"
	"github.com/ankur-anand/simple-go-rpc/src/transport"
	"io"
	"log"
	"net"
	"reflect"
)

// RPCServer ...
type RPCServer struct {
	addr  string
	funcs map[string]reflect.Value
}

// NewServer creates a new server
func NewServer(addr string) *RPCServer {
	return &RPCServer{addr: addr, funcs: make(map[string]reflect.Value)}
}

// Register the name of the function and its entries
func (s *RPCServer) Register(fnName string, fFunc interface{}) {
	// 如果 fnName 已被注册，则 s.funcs[fnName] 返回true，退出函数
	if _, ok := s.funcs[fnName]; ok {
		return
	}
	// 如果fnName未被注册，则将其添加到 s.funcs
	s.funcs[fnName] = reflect.ValueOf(fFunc)
}

// Execute the given function if present
func (s *RPCServer) Execute(req dataserial.RPCdata) dataserial.RPCdata {
	// get method by name
	f, ok := s.funcs[req.Name]
	if !ok {
		// since method is not present
		e := fmt.Sprintf("func %s not Registered", req.Name)
		log.Println(e)
		return dataserial.RPCdata{Name: req.Name, Args: nil, Err: e}
	}

	log.Printf("func %s is called\n", req.Name)
	// unpack request arguments 提取调用函数req中传入的各个参数到inArgs
	inArgs := make([]reflect.Value, len(req.Args))
	for i := range req.Args {
		inArgs[i] = reflect.ValueOf(req.Args[i])
	}

	// invoke requested method  执行函数得到结果out
	out := f.Call(inArgs)
	// now since we have followed the function signature style where last argument will be an error
	// so we will pack the response arguments expect error.	最后一个参数是error，所以长度为len(out)-1
	resArgs := make([]interface{}, len(out)-1)
	for i := 0; i < len(out)-1; i++ {
		// Interface returns the constant value stored in v as an interface{}.
		// out[i] 是一个 reflect.Value 类型的值，通过 out[i].Interface() 将其转换为 interface{}
		resArgs[i] = out[i].Interface()
	}

	// pack error argument
	var er string
	// 通过 (error) 类型断言将其转换为 error 类型；
	// 如果转换成功，ok 将被设置为 true，表示接口值属于 error 类型，即f.Call(inArgs)执行出错
	if _, ok := out[len(out)-1].Interface().(error); ok {
		// convert the error into error string value
		er = out[len(out)-1].Interface().(error).Error()
	}
	return dataserial.RPCdata{Name: req.Name, Args: resArgs, Err: er}
}

// Run server
func (s *RPCServer) Run() {
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Printf("listen on %s err: %v\n", s.addr, err)
		return
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("accept err: %v\n", err)
			continue
		}
		go func() {
			connTransport := transport.NewTransport(conn)
			for {
				// read request
				req, err := connTransport.Read()
				if err != nil {
					if err != io.EOF {
						log.Printf("read err: %v\n", err)
						return
					}
				}

				// decode the data and pass it to execute
				// 在服务器端反序列化请求信息，得到函数和传入的参数，保存到PRCdata结构中
				decReq, err := dataserial.Decode(req)
				if err != nil {
					log.Printf("Error Decoding the Payload err: %v\n", err)
					return
				}
				// get the executed result.
				resP := s.Execute(decReq)
				// encode the data back
				b, err := dataserial.Encode(resP)
				if err != nil {
					log.Printf("Error Encoding the Payload for response err: %v\n", err)
					return
				}
				// send response to client
				err = connTransport.Send(b)
				if err != nil {
					log.Printf("transport write err: %v\n", err)
				}
			}
		}()
	}
}
