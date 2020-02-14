package serviceDiscovery

import (
	"net/rpc"
	"github.com/silenceper/pool"
	log "github.com/sirupsen/logrus"
	//"github.com/fatih/pool"
	"github.com/pkg/errors"
	//"github.com/samuel/go-zookeeper/zk"

	"sync"
	"math/rand"
)
//map[srv_name]
//rpc clients  net/rpc  pool

const (
	Random int = iota
	Designate
)

type Client struct {
	//rpc Call Do concurs pools  interfaces
	//RpcClient *rpc.Client
	srvMap map[string]pool.Pool
	//srvName string ipList
	zoo *zooKeeper
	sync.RWMutex
}
//locks ip    call 拿RLOCK  . Watch LOck

// init Client ,zoo  sel_ip
//转发 selector
func (client *Client) Call(serviceMethod string, args interface{}, reply interface{},selector int,selPara string) error {
	//rr  choose_srv  concurs
	//1.nodeaddr Pool conn Call return
	var (
		cli *rpc.Client
		nodeAddr string
	)
	if selector == Designate {
		nodeAddr = selPara
	}else{
		index := rand.Intn(len(client.zoo.nodeList))
		nodeAddr = client.zoo.nodeList[index]
	}
	//load cache or new
	client.RLock()
	defer client.RUnlock()
	if p,exist := client.srvMap[nodeAddr] ; exist {
		iv,err := p.Get()
		if err != nil {
			return errors.New("srv down")
		}else{
			cli = iv.(*rpc.Client)
			//p.Close()
			err :=  cli.Call(serviceMethod,args,reply)
			p.Put(iv)
			return err
		}
	}else{
		//newPool paras configs  func defines  ___ Dial  func called  pool
		conf := &pool.Config{InitialCap:10,MaxCap:20,Factory:connFac(nodeAddr),Close:rpcClose}
		p,e := pool.NewChannelPool(conf)
		if e != nil {
			log.Error(e)
			return errors.New("srv down")
		}
		client.srvMap[nodeAddr] = p
		iv,err := p.Get()
		if err != nil {
			return errors.New("srv down")
		}else{
			cli = iv.(*rpc.Client)
			err := cli.Call(serviceMethod,args,reply)
			p.Put(iv)
			return err
		}
	}
	//return errors.New("srv down")
}


func New(config *ZkConfig) *Client{
	//zks watch mecha synced
	log.SetLevel(log.DebugLevel)
	zoo := newZoo(config)
	client := Client{zoo:zoo,srvMap:make(map[string]pool.Pool)}
	go client.watch()
	return &client
}

func (client *Client) watch(){
	//var data *[]string
	//ranges done sel chans
	for {
		select {
		case data := <- client.zoo.ch:
			client.Lock()
			client.zoo.nodeList = *data
			log.Infof("client.node %v\n",client.zoo.nodeList)
			client.Unlock()
		}
	}
}

//func (client *Client) DumpNodes() []string{
//	return client.zoo.nodeList
//}



