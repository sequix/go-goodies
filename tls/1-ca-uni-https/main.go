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
# CA签署证书 & client使用ca证书校验server

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

# 单向鉴权(只有client端校验server端的合法性)

client ----(tcp three way handshake)-----> server
client ----(Client Hello)-----> server
client <----(Server Hello，public key)----- server
client (通过 CA 根证书对服务器发过来的 server.crt 进行合法性验证)
client ----(客户端生成对称密钥，通过 public key 进行加密发送)-----> server
client ----(密钥交换后，按照对称密钥进行加密通讯)-----> server

# 6.访问
curl --cacert ca.cert https://localhost:8080 # 校验server
curl -k https://localhost:8080               # 不校验server
*/

func main() {
	// server
	http.HandleFunc("/", handler)
	go http.ListenAndServeTLS(":8080", "server.crt", "server.key", nil)

	req, err := http.NewRequest("GET", "https://localhost:8080/", nil)
	if err != nil {
		log.Fatal(err)
	}

	// client without server verification
	get(
		&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
		req,
	)

	// client with server verification
	certBytes, err := ioutil.ReadFile("ca.crt")
	if err != nil {
		log.Fatal(err)
	}

	clientCertPool := x509.NewCertPool()
	ok := clientCertPool.AppendCertsFromPEM(certBytes)
	if !ok {
		panic("failed to parse root certificate")
	}

	get(
		&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: clientCertPool,
				},
			},
		},
		req,
	)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi, This is an example of https service in golang!")
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
	fmt.Println(string(rsp1Body))
}
