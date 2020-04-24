package main

/*
# 自签名证书 & client使用server证书校验server

# 查看证书的内容
openssl x509 -text -noout -in server.crt

# 1.生成私钥文件server.key
openssl genrsa -out server.key 2048

# 指定加密算法，通过 'openssl genrsa -help' 查看支持的算法
openssl genrsa -aes128 -out server.key 2048

# 2.通过私钥生成证书文件server.crt
#    注意CN不要加端口号;
#    注意证书常以 .crt .pem 结尾
openssl req -new -x509 -days 3650 -key server.key -out server.crt \
    -subj "/C=CN/L=beijing/O=otaku/OU=male/CN=localhost/"

# e.g. 百度的证书中的身份信息
C  = CN                                               # Country Name
S  = beijing                                          # Sate or Province Name
L  = beijing                                          # Locality Name
O  = Beijing Baidu Netcom Science Technology Co., Ltd # Organization Name
OU = service operation department                     # Organization Unit Name
CN = baidu.com                                        # Common Name（证书所请求的域
名）

# 这种方式需提前将服务器的证书告知客户端
# 这样客户端在连接服务器时才能进行对服务器证书认证。
client ----(tcp three way handshake)-----> server
client ----(Client Hello)-----> server
client <----(Server Hello，public key)----- server
client (通过事先保存本地的 server.crt 和服务器发过来的 server.crt 进行比较)
client -(客户端生成对称密钥，通过server.crt提取public key进行加密发送)-> server
client ----(密钥交换后，按照对称密钥进行加密通讯)-----> server

# 3.访问
curl -k https://localhost:8080/                  # 不校验server
curl --cacert server.crt https://localhost:8080/ # 校验server
*/

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

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
	certBytes, err := ioutil.ReadFile("server.crt")
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
