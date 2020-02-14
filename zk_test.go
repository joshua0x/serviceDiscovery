package serviceDiscovery
import (
	"testing"
	"github.com/prometheus/common/log"
)
//MakeItWorks done
var config = ZkConfig{Addr:[]string{"127.0.0.1:2181"},BasePath:"/alex"}


type Args struct {
	A, B int
}


func TestZk(t *testing.T) {
	//zoo := newZoo(&ZkConfig{Addr:[]string{"127.0.0.1:2181"},BasePath:"/abc",SrvPath:"test"})
	//zoo.Register("ip1")
	//zoo.watch()
	//exports
	//client := New(&config)
	regisrv()
	//go func(){
	//	for {
	//	time.Sleep(time.Second * 2)
	//	arg := Args{1, 23}
	//	var reply int
	//	err := client.Call("Arith.Multiply", arg, &reply, Random, "")
	//	log.Error(err, reply)
	//}}()
	//go func(){
	//	for {
	//		time.Sleep(time.Second * 2)
	//		arg := Args{1, 23}
	//		var reply int
	//		err := client.Call("Arith.Multiply", arg, &reply, Designate, ":123")
	//		log.Error(err, reply)
	//	}}()
	//select {
	//
	//	}

}

func regisrv(){
	zoo := NewZooRegister(&config)
	err := zoo.Register(":9999")
	log.Error("register ",err)

}





