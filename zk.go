package serviceDiscovery

import (
	"github.com/samuel/go-zookeeper/zk"
	"time"
	log "github.com/sirupsen/logrus"
	"errors"
	"strings"
)

var zkConn *zk.Conn
//patched srvs  srvlist
// basePath
type ZkConfig struct {
	Addr []string
	BasePath string // for watch /im/test/123 base = /im srv=test
}



//replaced
type ZooKeeper struct {
	zkConn *zk.Conn
	basepath string
	nodeList []string
	ch chan *[]string
}

func NewZooRegister(config *ZkConfig) *ZooKeeper{
	var err error
	zkConn ,_,err = zk.Connect(config.Addr,time.Second*10)
	if err != nil {
		log.Panic(err)
	}
	createPath(config.BasePath)
	zoo := &ZooKeeper{zkConn:zkConn,basepath:config.BasePath}
	return zoo
	}

func newZoo(config *ZkConfig) *ZooKeeper{
	//init conn path_create
	var err error
	zkConn ,_,err = zk.Connect(config.Addr,time.Second*10)
	if err != nil {
		log.Panic(err)
	}
	createPath(config.BasePath)

	zoo := &ZooKeeper{zkConn:zkConn,ch:make(chan *[]string,1),basepath:config.BasePath}
	go zoo.watch()
	return zoo
	//return &zooKeeper{}  upds
}


//watched  getChilds path:value libkv  diffs  scan_deleted

func watchChildren(nodePath string) ([]string, <-chan zk.Event, error) {
	servers, _, ch, err := zkConn.ChildrenW(nodePath)
	if err == zk.ErrNoNode {
		log.Println("znode not exist", nodePath, err)
		panic(err)
	}
	return servers, ch, err
}


func getChildrenData(nodePath string, servers []string) ([]string, error) {

	data_arr := make([]string, 0, len(servers))
	for i := 0; i < len(servers); i++ {
		data, _, err := zkConn.Get(nodePath + "/" + servers[i])
		if err != nil {
			log.Println("get znode data error",)
			continue
		}
		data_arr = append(data_arr, string(data))
	}
	return data_arr, nil
}

//chans  chan [][]string consumed replaced
//bufs
//watch 写 Chans
//read     更新  srvList

func (zoo *ZooKeeper) watch(){
	for {
		servers, ch, err := watchChildren(zoo.basepath)
		if err != nil {
			log.Println(err)
			continue
		}
		//diffs zoo
		d, _ := getChildrenData(zoo.basepath, servers)
		log.Debug("getCdata")
		select {
		case zoo.ch <- &d:

		}
		log.Printf("got_data %v \n",d)
		<-ch
		//time.Sleep(time.Second*2)
	}
}

//update srvLists  tests register  zkConns
func (zoo *ZooKeeper)Register(data string) error {
	//createPath(nodePath)
	//serverName := nodePath + "/server"
	regiPath := zoo.basepath
	if zoo.basepath[len(zoo.basepath)-1:] != "/" {
		regiPath += "/"
	}
	ac_path, err := zkConn.Create(regiPath, []byte(data),zk.FlagSequence|zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	if err != nil && err != zk.ErrNodeExists {
		log.Println("create server znode err, path=", zoo.basepath, err)
		return errors.New("create server znode err")
	}
	log.Info("zkRegister",ac_path)
	return nil
}


func createPath(path string) error {
	currPath := ""
	for _, p := range strings.Split(path[1:], "/") {
		currPath += "/" + p
		_, err := zkConn.Create(currPath, []byte{1}, 0, zk.WorldACL(zk.PermAll))
		if err != nil && err != zk.ErrNodeExists {
			log.Println("create znode err", err, "path=", currPath)
			return errors.New("create znode error")
		}
	}
	return nil
}





