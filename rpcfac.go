package serviceDiscovery
import (
	"net/rpc"
	"github.com/prometheus/common/log"
)
//paras

func connFac(ip string) func()(interface{},error) {
	return func()(interface{},error){
		log.Info("fac ",ip)
		return rpc.DialHTTP("tcp",ip)
	}
}


func rpcClose(v interface{})error{
	client := v.(*rpc.Client)
	return client.Close()
}

