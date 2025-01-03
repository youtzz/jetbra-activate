package main

import (
	"crypto/rand"
	"crypto/rsa"
	"io"

	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"

	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"time"
	"encoding/asn1"

)

func main() {
	var err error

	certPath := "ca.crt"
	keyPath := "ca.key"
	err = genCert(certPath, keyPath)
	if err != nil {
		fmt.Println("Error generating cert:", err)
	}

	err = genPowerPluginConfig("ca.crt", "./")
	if err != nil {
		fmt.Println("Error generate config")
	}
}

func genCert(certPath, keyPath string) error {
	// Generate RSA key pair
	privKey, pubKey, err := generateRSAKeyPair(4096)
	if err != nil {
		return err
	}

	// Certificate details
	serialNumber := big.NewInt(time.Now().Unix())
	notBefore := time.Now().Add(-time.Hour * 24 * 365)
	notAfter := time.Now().Add(time.Hour * 24 * 365 * 10)

	// Create X509 certificate template
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   "llj",
		},
		Issuer: pkix.Name{
			CommonName: "JetProfile CA",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true, // Root certificate
	}

	// Create the certificate
	certDER, err := x509.CreateCertificate(
		nil,       // parent (nil because it's a self-signed certificate)
		&template, // certificate template
		&template, // parent (self-signed, so same as template)
		pubKey,    // public key
		privKey,   // private key
	)
	if err != nil {
		return err
	}

	// Save the certificate and private key to PEM files
	err = saveToPemFile(certPath, "CERTIFICATE", certDER)
	if err != nil {
		return err
	}

	pk, _ := x509.MarshalPKCS8PrivateKey(privKey)
	err = saveToPemFile(keyPath, "PRIVATE KEY", pk)
	if err != nil {
		return err
	}

	return nil
}

// generateRSAKeyPair generates a public/private key pair
func generateRSAKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	return privKey, &privKey.PublicKey, nil
}

// saveToPemFile saves the data to a PEM file
func saveToPemFile(filename, blockType string, data []byte) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	block := &pem.Block{
		Type:  blockType,
		Bytes: data,
	}
	err = pem.Encode(file, block)
	if err != nil {
		return err
	}
	return nil
}

const jetPrivateKey = "860106576952879101192782278876319243486072481962999610484027161162448933268423045647258145695082284265933019120714643752088997312766689988016808929265129401027490891810902278465065056686129972085119605237470899952751915070244375173428976413406363879128531449407795115913715863867259163957682164040613505040314747660800424242248055421184038777878268502955477482203711835548014501087778959157112423823275878824729132393281517778742463067583320091009916141454657614089600126948087954465055321987012989937065785013284988096504657892738536613208311013047138019418152103262155848541574327484510025594166239784429845180875774012229784878903603491426732347994359380330103328705981064044872334790365894924494923595382470094461546336020961505275530597716457288511366082299255537762891238136381924520749228412559219346777184174219999640906007205260040707839706131662149325151230558316068068139406816080119906833578907759960298749494098180107991752250725928647349597506532778539709852254478061194098069801549845163358315116260915270480057699929968468068015735162890213859113563672040630687357054902747438421559817252127187138838514773245413540030800888215961904267348727206110582505606182944023582459006406137831940959195566364811905585377246353"

func genPowerPluginConfig(certFile, baseDir string) error {
	// 读取证书内容
	certContent, err := ioutil.ReadFile(certFile)
	if err != nil {
		return err
	}

	// 解析 X.509 证书
	block, _ := pem.Decode(certContent)
	if block == nil {
		return fmt.Errorf("failed to parse PEM block containing the certificate")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return err
	}

	// 提取公钥
	publicKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("not an RSA public key")
	}

	// 获取公钥的指数
	exponent := publicKey.E

	// 计算证书的签名
	bytes, err := encodeSignature(cert.RawTBSCertificate, publicKey.N.BitLen())
	if err != nil {
		return err
	}

	// 创建 power.conf 配置内容
	powerConfig := fmt.Sprintf("[Result]\nEQUAL,%s,%d,%s->%s",
		new(big.Int).SetBytes(cert.Signature).String(),
		exponent,
		jetPrivateKey,
		new(big.Int).SetBytes(bytes).String(),
	)

	// 写入配置文件
	err = ioutil.WriteFile(baseDir+"/power.conf", []byte(powerConfig), 0644)
	if err != nil {
		return err
	}

	return nil
}

// Define the ASN.1 structures
type pkcs1DigestInfo struct {
	Algorithm asn1.ObjectIdentifier
	Digest    []byte
}

var oidSHA256 = asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 1}

func encodeSignature(values []byte, keySize int) ([]byte, error) {
	if keySize%8 != 0 {
			return nil, fmt.Errorf("key size must be a multiple of 8")
	}

	hashed := sha256.Sum256(values)

	digestInfo := pkcs1DigestInfo{
			Algorithm: oidSHA256,
			Digest:    hashed[:],
	}

	asn1Bytes, err := asn1.Marshal(digestInfo)
	if err != nil {
			return nil, fmt.Errorf("asn1 marshal error: %w", err)
	}

// PKCS#1 v1.5 padding (equivalent to RSAPadding.PAD_BLOCKTYPE_1)
padded := make([]byte, (keySize+7)/8)
padded[1] = 0x01
for i := 2; i < len(padded)-len(asn1Bytes)-1; i++ {
	padded[i] = 0xff
}
padded[len(padded)-len(asn1Bytes)-1] = 0x00
copy(padded[len(padded)-len(asn1Bytes):], asn1Bytes)

	return padded, nil
}

// encodeSignature 使用 SHA-256 和 RSA 填充来编码签名
func encodeSignature2(pubKey *rsa.PublicKey, values []byte) ([]byte, error) {
	// 计算 SHA-256 哈希
	hash := sha256.New()
	hash.Write(values)
	digest := hash.Sum(nil)

	// 使用 RSA 填充生成签名
	// Go 直接使用 pkcs1v15 填充（通常用于签名），并且它已经实现了 RSAPadding

	signature, err := PKCS1Padding(pubKey, digest)
	if err != nil {
		return nil, err
	}
	return signature, nil
}

func PKCS1Padding(pub *rsa.PublicKey, msg []byte) ([]byte, error) {
	k := pub.Size()
	if len(msg) > k-11 {
		return nil, nil
	}

	em := make([]byte, k)
	em[1] = 2
	ps, mm := em[2:len(em)-len(msg)-1], em[len(em)-len(msg):]
	err := nonZeroRandomBytes(ps, rand.Reader)
	if err != nil {
		return nil, err
	}

	em[len(em)-len(msg)-1] = 0
	copy(mm, msg)
	return em, nil
}

func nonZeroRandomBytes(s []byte, rand io.Reader) error {
	_, err := io.ReadFull(rand, s)
	if err != nil {
		return err
	}

	for i := 0; i < len(s); i++ {
		for s[i] == 0 {
			_, err = io.ReadFull(rand, s[i:i+1])
			if err != nil {
				return err
			}
			s[i] ^= 0x42
		}
	}
	return nil
}
