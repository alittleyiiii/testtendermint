package carstore

import (
	"encoding/json"
	"fmt"
	types2 "github.com/tendermint/tendermint/abci/types"
	"io/ioutil"
	"log"
	"time"
	"unsafe"

	crypto_rand "crypto/rand"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	cmn "github.com/tendermint/tendermint/libs/os"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	"golang.org/x/crypto/nacl/box"
)

//为北谷建立客户端发起交易,始终觉着这么写有问题，先这样吧。client里面没有广播交易的函数BroadcastTxSync，所以就用了我之前修改的一份代码
//同时，用客户端的话就始终需要建立socket连接，但是我们发起交易的方式是在系统内部，不需要用的socket
//localclient提供了本地接口，没读懂，为什么要用到mutex，同时localclient里也没有交易同步代码
//app := kvstore.NewApplication()
//cc := proxy.NewLocalClientCreator(app)
//var bgclient abcicli.Client
var (
	bgcli rpchttp.HTTP
	//client abcicli.Client
	logger log.Logger
)
var (
	// global
	flagAddress  string
	flagAbci     string
	flagVerbose  bool // for the println output
	flagLogLevel string
)

//启动北谷客户端
func init() {
	bgcli, err := rpchttp.New(flagAddress, "/websocket")
	if err != nil {
		log.Fatalf("file open error : %v", err)
	}
	if err := bgcli.Start(); err != nil {
		log.Fatalf("clent not start : %v", err)
	}
}

//--------------------------------------------------------------------
//签名代码
// KEYFILENAME 私钥文件名
const KEYFILENAME string = ".userkey"

type cryptoPair struct {
	PrivKey *[32]byte
	PubKey  *[32]byte
}

type user struct {
	SignKey    crypto.PrivKey `json:"sign_key"` // 节点私钥，用户签名
	CryptoPair cryptoPair     // 密钥协商使用
	bottleIDs  []string       // 投放的所有漂流瓶id集合
}

func loadOrGenUserKey() (*user, error) {
	if cmn.FileExists(KEYFILENAME) {
		uk, err := loadUserKey()
		if err != nil {
			return nil, err
		}
		return uk, nil
	}
	//fmt.Println("userkey file not exists")
	uk := new(user)
	uk.SignKey = ed25519.GenPrivKey()
	pubKey, priKey, err := box.GenerateKey(crypto_rand.Reader)
	if err != nil {
		return nil, err
	}
	uk.CryptoPair = cryptoPair{PrivKey: priKey, PubKey: pubKey}
	jsonBytes, err := json.Marshal(uk)
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(KEYFILENAME, jsonBytes, 0644)
	if err != nil {
		return nil, err
	}
	return uk, nil
}
func loadUserKey() (*user, error) {
	//copy(privKey[:], bz)
	jsonBytes, err := ioutil.ReadFile(KEYFILENAME)
	if err != nil {
		return nil, err
	}
	uk := new(user)
	err = json.Unmarshal(jsonBytes, uk)
	if err != nil {
		return nil, fmt.Errorf("Error reading UserKey from %v: %v", KEYFILENAME, err)
	}
	return uk, nil
}

//--------------------------------------------------------------------------------
//签名结束，生成交易，打包交易，同步交易
//user==北谷
func str2bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func bytes2str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func (bg *user) throwTx(content string) error {
	//交易填充
	now := time.Now()
	tx := new(types2.Transx)
	tx.SendTime = &now
	var merkleRoot []byte
	i := 0
	for i <= 10 {
		merkleRoot = GenMerkleRoot()
		//merkle是[]byte型，在交易中用string存的
		tx.Merkletoot[i] = bytes2str(merkleRoot)
	}
	tx.SendTime = &now
	//交易签名
	tx.Sign(bg.SignKey)
	tx.SignPubKey = bg.SignKey.PubKey()
	bz, err := json.Marshal(&tx)
	if err != nil {
		return err
	}
	//异步广播交易
	ret, err := bgcli.BroadcastTxSync(bz)
	if err != nil {
		return err
	}
	fmt.Printf("BroadcastTx=> %+v\n", ret)
	return nil
}
