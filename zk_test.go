package serviceDiscovery
import (
	"testing"
)
//MakeItWorks done
var config = ZkConfig{Addr:[]string{"127.0.0.1:2181"},BasePath:"/abc",SrvPath:"test"}

func TestZk(t *testing.T) {
	//zoo := newZoo(&ZkConfig{Addr:[]string{"127.0.0.1:2181"},BasePath:"/abc",SrvPath:"test"})
	//zoo.Register("ip1")
	//zoo.watch()
	//exports
	New(&config)
	//call srvs
	select {

	}



}





