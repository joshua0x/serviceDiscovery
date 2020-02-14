package serviceDiscovery
import (
	"net/rpc"
	log "github.com/sirupsen/logrus"
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

