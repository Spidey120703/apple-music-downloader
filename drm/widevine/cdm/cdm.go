package cdm

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"math/rand"
	"time"

	"github.com/aead/cmac"
	"google.golang.org/protobuf/proto"
)

var random *rand.Rand

func init() {
	random = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func RandomSessionID() (s [32]byte) {
	h := []byte("0123456789ABCDEF")
	var i int
	for i = 0; i < 16; i++ {
		s[i] = h[random.Intn(len(h))]
	}
	s[16] = '0'
	s[17] = '1'
	for i += 2; i < 32; i++ {
		s[i] = '0'
	}
	return
}

type ContentDecryptionModule struct {
	PrivateKey *rsa.PrivateKey
	ClientID   []byte
	SessionID  [32]byte
	WidevineCencHeader
}

func (c *ContentDecryptionModule) Init(privateKeyRaw string, clientID []byte, initData []byte) (err error) {
	block, _ := pem.Decode([]byte(privateKeyRaw))
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return errors.New("private key is invalid")
	}
	c.PrivateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return
	}

	c.ClientID = clientID
	c.SessionID = RandomSessionID()

	if len(initData) < 32 {
		return errors.New("initData is too short")
	}
	if err = proto.Unmarshal(initData[32:], &c.WidevineCencHeader); err != nil {
		return
	}
	return
}

func assign[T any](dest **T, src T) {
	*dest = new(T)
	**dest = src
}

func (c *ContentDecryptionModule) Challenge() (challenge []byte, err error) {
	var signedLicenseRequest SignedLicenseRequest
	assign(&signedLicenseRequest.Type, SignedLicenseRequest_LICENSE_REQUEST)
	signedLicenseRequest.Msg = new(LicenseRequest)
	{
		signedLicenseRequest.Msg.ClientId = new(ClientIdentification)
		{
			if err := proto.Unmarshal(c.ClientID, signedLicenseRequest.Msg.ClientId); err != nil {
				return nil, err
			}
		}

		signedLicenseRequest.Msg.ContentId = new(LicenseRequest_ContentIdentification)
		{
			signedLicenseRequest.Msg.ContentId.CencId = new(LicenseRequest_ContentIdentification_CENC)
			{
				signedLicenseRequest.Msg.ContentId.CencId.Pssh = &c.WidevineCencHeader
				assign(&signedLicenseRequest.Msg.ContentId.CencId.LicenseType, LicenseType_STREAMING)
				signedLicenseRequest.Msg.ContentId.CencId.RequestId = c.SessionID[:]
			}
		}

		assign(&signedLicenseRequest.Msg.Type, LicenseRequest_NEW)
		assign(&signedLicenseRequest.Msg.RequestTime, time.Now().Unix())
		assign(&signedLicenseRequest.Msg.ProtocolVersion, ProtocolVersion_VERSION_2_1)
		assign(&signedLicenseRequest.Msg.KeyControlNonce, random.Uint32())
	}
	{
		data, err := proto.Marshal(signedLicenseRequest.Msg)
		if err != nil {
			return nil, err
		}
		hash := sha1.Sum(data)
		signedLicenseRequest.Signature, err = rsa.SignPSS(
			crand.Reader,
			c.PrivateKey,
			crypto.SHA1,
			hash[:],
			&rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash},
		)
		if err != nil {
			return nil, err
		}
	}
	return proto.Marshal(&signedLicenseRequest)
}

// PKCS7Unpadding removes AES/CBC/PKCS7Padding from buf.
//
// Reference: Public-Key Cryptography Standards #7 (PKCS #7)
func PKCS7Unpadding(buf []byte) []byte {
	if buf == nil || len(buf) == 0 {
		return buf
	}
	padCount := buf[len(buf)-1]
	return buf[:len(buf)-int(padCount)]
}

func (c *ContentDecryptionModule) GetLicenseKeys(challenge []byte, license []byte) (licenseKeys []LicenseKey, err error) {
	var signedLicenseRequest SignedLicenseRequest
	if err = proto.Unmarshal(challenge, &signedLicenseRequest); err != nil {
		return
	}
	licenseRequest, err := proto.Marshal(signedLicenseRequest.Msg)
	if err != nil {
		return
	}

	var signedLicense SignedLicense
	if err = proto.Unmarshal(license, &signedLicense); err != nil {
		return
	}

	var cipherBlock cipher.Block
	{
		sessionKey, err := rsa.DecryptOAEP(sha1.New(), crand.Reader, c.PrivateKey, signedLicense.SessionKey, nil)
		if err != nil {
			return nil, err
		}
		sessionKeyBlock, err := aes.NewCipher(sessionKey)
		if err != nil {
			return nil, err
		}

		encryptionKey := []byte("\x01ENCRYPTION\x00")
		encryptionKey = append(encryptionKey, licenseRequest...)
		encryptionKey = append(encryptionKey, 0, 0, 0, 0x80)

		encryptionKeyCmac, err := cmac.Sum(encryptionKey, sessionKeyBlock, sessionKeyBlock.BlockSize())
		if err != nil {
			return nil, err
		}
		cipherBlock, err = aes.NewCipher(encryptionKeyCmac)
		if err != nil {
			return nil, err
		}
	}

	for _, key := range signedLicense.Msg.Key {
		decrypter := cipher.NewCBCDecrypter(cipherBlock, key.Iv)
		decryptedKey := make([]byte, len(key.Key))
		decrypter.CryptBlocks(decryptedKey, key.Key)
		licenseKeys = append(licenseKeys, LicenseKey{
			ID:   key.Id,
			Type: *key.Type,
			Key:  PKCS7Unpadding(decryptedKey),
		})
	}

	return
}

type LicenseKey struct {
	ID   []byte
	Type License_KeyContainer_KeyType
	Key  []byte
}

func New(initData []byte) (*ContentDecryptionModule, error) {
	cdm := &ContentDecryptionModule{}
	err := cdm.Init(DefaultPrivateKey, DefaultClientID, initData)
	return cdm, err
}
