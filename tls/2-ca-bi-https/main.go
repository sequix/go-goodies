package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

/*
# CA签署证书 & 双向校验

# 查看证书的内容
openssl x509 -text -noout -in server.crt

# 1.生成ca私钥
openssl genrsa -out ca.key 2048

# 2.根据私钥生成ca证书(公钥)，注意ca证书中的CN域名不能和该ca签署的证书中的CN域名相同
openssl req -x509 -new -nodes -key ca.key -subj "/CN=ca.com" -days 5000 -out ca.crt

# 3.生成server端私钥
openssl genrsa -out server.key 2048

# 4.生成server端的csr (证书签署请求Certificate Sign Request，指的是向ca请求签署证书)
openssl req -new -key server.key -subj "/CN=localhost" -out server.csr

# 5.把server端的csr传给ca，ca用私钥签署（参数中也要传递ca证书）
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 5000

# client端生成ca签署证书的流程同server端
openssl genrsa -out client.key 2048
openssl req -new -key client.key -subj "/CN=localhost" -out client.csr
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt -days 5000

# 双向鉴权
client ----(tcp three way handshake)-----> server
client ----(Client Hello)-----> server
client <----(Server Hello，public key)----- server
client (通过 CA 根证书对服务器发过来的 server.crt 进行合法性验证)
client ----(客户端生成对称密钥，通过 public key 进行加密发送，并发送客户端 client.crt)-----> server
server (服务端使用 CA 根证书对 client.crt 进行做法性校验)
client ----(密钥交换后，按照对称密钥进行加密通讯)-----> server

# 6.访问
# server校验client，client不校验server
curl --key client.key -E client.crt -k https://localhost:8080/
# client校验server，server不校验client
client无法决定server是否要校验自己，所以这种场景不成立。
# 双向校验
curl --cacert ca.crt --key client.key -E client.crt https://localhost:8080/
# 双向均不校验
那是http。
*/
func main() {
	caCrt, err := ioutil.ReadFile("ca.crt")
	if err != nil {
		fmt.Println("ReadFile err:", err)
		return
	}

	req, err := http.NewRequest("GET", "https://localhost:8080/", nil)
	if err != nil {
		log.Fatal(err)
	}

	// server
	serverPool := x509.NewCertPool()
	serverPool.AppendCertsFromPEM(caCrt)

	s := &http.Server{
		Addr:    ":8080",
		Handler: &myhandler{},
		TLSConfig: &tls.Config{
			ClientCAs:  serverPool,
			ClientAuth: tls.RequireAndVerifyClientCert,
		},
	}
	s.ListenAndServeTLS("server.crt", "server.key")

	// client without server verification
	clientPool := x509.NewCertPool()
	clientPool.AppendCertsFromPEM(caCrt)
	clientCrt, err := tls.LoadX509KeyPair("client.crt", "client.key")
	if err != nil {
		log.Fatal(err)
	}

	get(
		&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
					Certificates:       []tls.Certificate{clientCrt},
				},
			},
		},
		req,
	)

	// client with server verification
	get(
		&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs:      clientPool,
					Certificates: []tls.Certificate{clientCrt},
				},
			},
		},
		req,
	)
}

type myhandler struct {
}

func (h *myhandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi, This is an example of http service in golang!")
}

func get(client *http.Client, req *http.Request) {
	rsp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer rsp.Body.Close()
	rsp1Body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(string(rsp1Body))
}
