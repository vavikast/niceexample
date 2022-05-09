package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

func main()  {
	privKey := loadPrivateKeyFromPemFile1("")

	pubkey := privKey.PublicKey

	cecertTemplate := setupCeritificateTemplate(true)
	certificate, err := x509.CreateCertificate(rand.Reader, &cecertTemplate, &cecertTemplate, &pubkey, privKey)
	if err != nil {
		log.Fatal("Failed to create to certificate.",err)
	}
	writeCertToPemFile("",certificate)
}

func setupCeritificateTemplate(isCa bool) x509.Certificate  {
	notBefore := time.Now()

	notAfter := notBefore.Add(time.Hour*24*365)

	//Generate secure random serial number
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	randomNumber,err := rand.Int(rand.Reader,serialNumberLimit)
	if err != nil {
		log.Fatal("Error generating random serial number. ",err)
	}
	nameInfo := pkix.Name{
		Organization:       []string{"My organization"},
		CommonName:         "localhost",
		OrganizationalUnit: []string{"my Business unit"},
		Country: []string{"china"},
		Province: []string{"gd"},
		Locality: []string{"sz"},
	}
	certTemplate := x509.Certificate{
		SerialNumber: randomNumber,
		Subject: nameInfo,
		EmailAddresses: []string{"test@localhost"},
		NotBefore: notBefore,
		NotAfter: notAfter,
		KeyUsage: x509.KeyUsageKeyEncipherment|x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
		BasicConstraintsValid: true,
		IsCA: false,
	}

	//to create ceritifacate authority
	if isCa{
		certTemplate.IsCA=true
		certTemplate.KeyUsage = certTemplate.KeyUsage|x509.KeyUsageCertSign
	}
	certTemplate.IPAddresses = []net.IP{net.ParseIP("127.0.0.1")}
	certTemplate.DNSNames = []string{"localhost","localhost.local"}
	return certTemplate
}
func loadPrivateKeyFromPemFile1(privateKeyFilename string)*rsa.PrivateKey  {
	fileData, err := ioutil.ReadFile(privateKeyFilename)
	if err != nil {
		log.Fatal("Error loading private key file.",err)
	}
	//Get the block data form the pem encoded file
	block, _ := pem.Decode(fileData)
	if block == nil || block.Type != "RSA PRIVATE KEY"{
		log.Fatal("unable to load a valid private key")
	}
	privatekey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		log.Fatal("Error loading private key.",err)
	}
	return privatekey
}
func writeCertToPemFile(outputFilename string,derBytes []byte)  {
	//Create  a Pem from the certficate
	certPem := &pem.Block{Type: "CERRIFICATE",Bytes: derBytes}
	certOutfile, err := os.Create(outputFilename)
	if err != nil {
		log.Fatal("unable to open certificate output file ",err)
	}
	pem.Encode(certOutfile,certPem)
	certOutfile.Close()
}


