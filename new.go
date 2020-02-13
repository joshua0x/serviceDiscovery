package serviceDiscovery

import (
	"net/rpc"
	//"github.com/silenceper/pool"
	"github.com/joshua0x/pool"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	//"github.com/samuel/go-zookeeper/zk"
	"math/rand"
	"sync"
)

//map[srv_name]
//rpc clients  net/rpc  pool

const (
	Random int = iota
	Designate
)

var pconf *pool.Config

type PoolOptions struct {
	MaxCap, InitialCap,MaxIdle int
}
type SdConfig struct {
	Zkc   ZkConfig
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


// init Client ,zoo  sel_ip
//转发 selector
func (client *Client) Call(serviceMethod string, args interface{}, reply interface{}, selector int, selPara string) error {
	//rr  choose_srv  concurs
	//1.nodeaddr Pool conn Call return
	var (
		cli      *rpc.Client
		nodeAddr string
	)

	client.RLock()
	//todo error_retry
	if selector == Designate {
		nodeAddr = selPara
	} else {
		if len(client.zoo.nodeList) == 0 {
			return errors.New("srv down")
		}
		index := rand.Intn(len(client.zoo.nodeList))
		nodeAddr = client.zoo.nodeList[index]
	}
	//load cache or
	log.Debug("nodeAddr", nodeAddr)
	if p, exist := client.srvMap[nodeAddr]; exist {
		client.RUnlock()
		iv, err := p.Get()
		if err != nil {
			return err
		} else {
			cli = iv.(*rpc.Client)
			//p.Close()
			err := cli.Call(serviceMethod, args, reply)
			if err != nil {
				p.Close(iv)
				return err
			}
			p.Put(iv)
			return err
		}
	} else {
		client.RUnlock()
		client.Lock()
		log.Debugf("newChanPool %v %v\n", nodeAddr,pconf)
		//newPool paras configs  func defines  ___ Dial  func called  pool
		conf := &pool.Config{InitialCap: pconf.InitialCap, MaxCap: pconf.MaxCap,
			Factory: connFac(nodeAddr), Close: rpcClose,MaxIdle:pconf.MaxIdle}
		p, e := pool.NewChannelPool(conf)
		if e != nil {
			log.Error(e)
			client.Unlock()
			return e
		}
		client.srvMap[nodeAddr] = p
		client.Unlock()
		iv, err := p.Get()
		if err != nil {
			return e
		} else {
			cli = iv.(*rpc.Client)
			err := cli.Call(serviceMethod, args, reply)
			//todo conn err close conn
			if err != nil {
				p.Close(iv)
				return err
			}
			p.Put(iv)
			return err
		}
	}
	//return errors.New("srv down")
}

func New(config *SdConfig) *Client {
	//zks watch mecha synced
	pconf = &pool.Config{InitialCap: config.PoolC.InitialCap, MaxCap: config.PoolC.MaxCap,MaxIdle:config.PoolC.MaxIdle}
	zoo := newZoo(&config.Zkc)
	client := Client{zoo: zoo, srvMap: make(map[string]pool.Pool)}
	go client.watch()
	return &client
}

func (client *Client) watch() {
	//var data *[]string
	//ranges done sel chans
	for {
		select {
		case data := <-client.zoo.ch:
			client.Lock()
			client.zoo.nodeList = *data
			//del srvCaches []str
			for k, p := range client.srvMap {
				found := false
				for _, node := range *data {
					if node == k {
						found = true
						break
					}
				}
				if !found {
					delete(client.srvMap, k)
					//todo ReleaseConn
					p.Release()
				}
			}
			log.Infof("client.node %v\n", client.zoo.nodeList)
			client.Unlock()
		}
	}
}

//func (client *Client) DumpNodes() []string{
//	return client.zoo.nodeList
//}
