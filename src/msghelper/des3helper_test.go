package msghelper

import (
	"testing"
)

func Test_des3Encode(t *testing.T) {
	//测试编码
	myDes3Helper := &Des3Helper{m_strKey: "123456789012345678901234", m_strIV: ConstStrIV}

	//测试加密
	encodestring := myDes3Helper.Encode("{\"token\":\"123456789012345678901234\",\"cardId\":\"11111112\"}")
	t.Log("encode string:", encodestring)
}

func Test_des3Decode(t *testing.T) {
	//测试解码
	myDes3Helper := &Des3Helper{m_strKey: "123456789012345678901234", m_strIV: ConstStrIV}

	//测试解密
	decodeString := myDes3Helper.Decode("KmFtkK9xdZTBrpLedtKgWU141SRSbU2PKNmufmW5IpnqKVn2e+QqwdofybrghxtvMdepdeFFQhYC68J2XRhOUg==")
	t.Log("decode string:", decodeString)

}
