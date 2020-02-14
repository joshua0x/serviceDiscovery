package serviceDiscovery
import (
	"net/rpc"
)
//paras

func connFac(ip string) func()(interface{},error) {
	return func()(interface{},error){
		return rpc.Dial("tcp",ip)
	}
}


func rpcClose(v interface{})error{
	client := v.(*rpc.Client)
	return client.Close()
}

