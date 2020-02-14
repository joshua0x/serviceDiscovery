package serviceDiscovery

import (
	"net/rpc"
	//"github.com/silenceper/pool"
	"github.com/joshua0x/pool"
	log "github.com/sirupsen/logrus"
	//"github.com/fatih/pool"
	"github.com/pkg/errors"
	"github.com/keepeye/logrus-filename"
	//"github.com/samuel/go-zookeeper/zk"
	"time"
	"sync"
	"math/rand"
)
//map[srv_name]
//rpc clients  net/rpc  pool

const (
	Random int = iota
	Designate
)
var pconf *pool.Config

type PoolOptions struct {
	MaxCap ,InitialCap int
	WaitTimeOut time.Duration
}
type SdConfig struct {
	Zkc ZkConfig
	PoolC PoolOptions
}

type Client struct {
	//rpc Call Do concurs pools  interfaces
	//RpcClient *rpc.Client
	srvMap map[string]pool.Pool
	//srvName string ipList
	zoo *ZooKeeper
	sync.RWMutex
}

func init(){
	filenameHook := filename.NewHook()
	filenameHook.Field = "line"
	log.AddHook(filenameHook)
}

// init Client ,zoo  sel_ip
//转发 selector
func (client *Client) Call(serviceMethod string, args interface{}, reply interface{},selector int,selPara string) error {
	//rr  choose_srv  concurs
	//1.nodeaddr Pool conn Call return
	var (
		cli *rpc.Client
		nodeAddr string
	)

	client.RLock()
	defer client.RUnlock()
	//todo error_retry
	if selector == Designate {
		nodeAddr = selPara
	}else{
		if len(client.zoo.nodeList) == 0  {
			return errors.New("srv down")
		}
		index := rand.Intn(len(client.zoo.nodeList))
		nodeAddr = client.zoo.nodeList[index]
	}
	//load cache or
	log.Info("nodeAddr",nodeAddr)
	if p,exist := client.srvMap[nodeAddr] ; exist {
		iv,err := p.Get()
		if err != nil {
			return err
		}else{
			cli = iv.(*rpc.Client)
			//p.Close()
			err :=  cli.Call(serviceMethod,args,reply)
			p.Put(iv)
			return err
		}
	}else{
		//newPool paras configs  func defines  ___ Dial  func called  pool
		conf := &pool.Config{InitialCap:pconf.InitialCap,MaxCap:pconf.MaxCap,WaitTimeOut:pconf.WaitTimeOut,
				Factory:connFac(nodeAddr),Close:rpcClose}
		p,e := pool.NewChannelPool(conf)
		if e != nil {
			log.Error(e)
			return e
		}
		client.srvMap[nodeAddr] = p
		iv,err := p.Get()
		if err != nil {
			return e
		}else{
			cli = iv.(*rpc.Client)
			err := cli.Call(serviceMethod,args,reply)
			//todo conn err close conn
			if err != nil {
				cli.Close()
				return err
			}
			p.Put(iv)
			return err
		}
	}
	//return errors.New("srv down")
}


func New(config *SdConfig) *Client{
	//zks watch mecha synced
	pconf = &pool.Config{InitialCap:config.PoolC.InitialCap,MaxCap:config.PoolC.MaxCap,WaitTimeOut:config.PoolC.WaitTimeOut}
	zoo := newZoo(&config.Zkc)
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
			//del srvCaches []str
			for k,p := range client.srvMap{
				found := false
				for _,node := range *data{
					if node == k {
						found = true
						break
					}
				}
				if !found{
					delete(client.srvMap,k)
					//todo ReleaseConn
					p.Release()
				}
			}
			log.Infof("client.node %v\n",client.zoo.nodeList)
			client.Unlock()
		}
	}
}

//func (client *Client) DumpNodes() []string{
//	return client.zoo.nodeList
//}



