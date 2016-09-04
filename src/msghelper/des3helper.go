package msghelper

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"encoding/base64"
	"fmt"
)

const (
	ConstStrIV = "12345678" //加密使用的IV，直接写死
)

type Des3Helper struct {
	m_strKey string //DES3的key
	m_strIV  string //DES3的IV
}

//编码
func (this *Des3Helper) Encode(str string) string {
	return des3encode(str, this.m_strKey, this.m_strIV)
}

//解码
func (this *Des3Helper) Decode(str string) string {
	return des3decode(str, this.m_strKey, this.m_strIV)
}

//设置des3 key
func (this *Des3Helper) SetKey(strKey string) {
	this.m_strKey = strKey
}

//设置des3 IV
func (this *Des3Helper) SetIV(strIV string) {
	this.m_strIV = strIV
}

//DES3编码
func des3encode(str, strKey, strIV string) string {
	//DES3加密
	block, err := des.NewTripleDESCipher([]byte(strKey))
	if err != nil {
		return ""
	}

	originData := pKCS5Padding([]byte(str), block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, []byte(strIV))
	crypted := make([]byte, len(originData))
	blockMode.CryptBlocks(crypted, originData)

	//Base64编码
	return base64.StdEncoding.EncodeToString(crypted)
}

//DES3解码
func des3decode(str, strKey, strIV string) string {
	//BASE64 解码
	crypted, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		fmt.Println("Failed to base64 decode string:", str)
		return ""
	}

	//DES3 解码
	block, err := des.NewTripleDESCipher([]byte(strKey))
	if err != nil {
		fmt.Println("Failed to new triple cihper!")
		return ""
	}

	blockMode := cipher.NewCBCDecrypter(block, []byte(strIV))
	originData := make([]byte, len(crypted))
	blockMode.CryptBlocks(originData, crypted)
	originData = pKCS5UnPadding(originData)

	return string(originData)
}

func pKCS5UnPadding(origData []byte) []byte {
	length := len(origData)

	// 去掉最后一个字节 unpadding 次
	unpadding := int(origData[length-1])

	return origData[:(length - unpadding)]
}

func pKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)

	return append(ciphertext, padtext...)
}
