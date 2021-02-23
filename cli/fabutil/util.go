package fabutil

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"github.com/golang/protobuf/proto"
	pb_msp "github.com/hyperledger/fabric-protos-go/msp"
	"github.com/pkg/errors"
	"io/ioutil"
	"math/big"
)

type ecdsaSignature struct {
	R, S *big.Int
}

var (
	// curveHalfOrders contains the precomputed curve group orders halved.
	// It is used to ensure that signature' S value is lower or equal to the
	// curve group order halved. We accept only low-S signatures.
	// They are precomputed for efficiency reasons.
	curveHalfOrders = map[elliptic.Curve]*big.Int{
		elliptic.P224(): new(big.Int).Rsh(elliptic.P224().Params().N, 1),
		elliptic.P256(): new(big.Int).Rsh(elliptic.P256().Params().N, 1),
		elliptic.P384(): new(big.Int).Rsh(elliptic.P384().Params().N, 1),
		elliptic.P521(): new(big.Int).Rsh(elliptic.P521().Params().N, 1),
	}
)

func Sign(object []byte, key *ecdsa.PrivateKey) ([]byte, error) {
	if len(object) == 0 {
		return nil, errors.New("object (to sign) required")
	}

	if key == nil {
		return nil, errors.New("key (for signing) required")
	}
	hash := sha256.New()
	hash.Write(object)
	digest := hash.Sum(nil)
	r, s, err := ecdsa.Sign(rand.Reader, key, digest)
	if err != nil {
		panic(err)
	}
	s, err = toLowS(&key.PublicKey, s)
	if err != nil {
		panic(err)
	}

	signature, err := asn1.Marshal(ecdsaSignature{r, s})
	if err != nil {
		return nil, err
	}
	return signature, nil

}

func GetPrivateKey(privateKeyPath string) (*ecdsa.PrivateKey, error) {
	raw, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return nil, errors.Errorf("Failed loading private key [%s]: [%v].", privateKeyPath, err)
	}
	privateKey, err := pemToPrivateKey(raw, nil)
	if err != nil {
		return nil, errors.Errorf("Failed parsing private key [%s]: [%s].", privateKeyPath, err.Error())
	}

	pKey, ok := privateKey.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.Errorf("Failed loading private key from [%s]", privateKeyPath)
	}

	return pKey, nil
}

func GetCreator(mspID, signCertPath string) ([]byte, error) {
	idBytes, err := ioutil.ReadFile(signCertPath)
	if err != nil {
		return nil, errors.Errorf("Failed loading identity info [%s]: [%v].", signCertPath, err)
	}
	serializedIdentity := &pb_msp.SerializedIdentity{
		Mspid:   mspID,
		IdBytes: idBytes,
	}
	identity, err := proto.Marshal(serializedIdentity)
	if err != nil {
		return nil, errors.Errorf("Failed loading identity info [%s]: [%v].", signCertPath, err)
	}
	return identity, nil
}

func pemToPrivateKey(raw []byte, pwd []byte) (interface{}, error) {
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, errors.Errorf("failed decoding PEM. Block must be different from nil [% x]", raw)
	}

	// TODO: derive from header the type of the key

	if x509.IsEncryptedPEMBlock(block) {
		if len(pwd) == 0 {
			return nil, errors.New("encrypted Key. Need a password")
		}

		decrypted, err := x509.DecryptPEMBlock(block, pwd)
		if err != nil {
			return nil, errors.Errorf("failed PEM decryption: [%s]", err)
		}

		key, err := derToPrivateKey(decrypted)
		if err != nil {
			return nil, err
		}
		return key, err
	}

	cert, err := derToPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return cert, err
}

func derToPrivateKey(der []byte) (key interface{}, err error) {

	if key, err = x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}

	if key, err = x509.ParsePKCS8PrivateKey(der); err == nil {
		switch key.(type) {
		case *ecdsa.PrivateKey:
			return
		default:
			return nil, errors.New("found unknown private key type in PKCS#8 wrapping")
		}
	}

	if key, err = x509.ParseECPrivateKey(der); err == nil {
		return
	}

	return nil, errors.New("invalid key type. The DER must contain an ecdsa.PrivateKey")
}

// IsLow checks that s is a low-S
func isLowS(k *ecdsa.PublicKey, s *big.Int) (bool, error) {
	halfOrder, ok := curveHalfOrders[k.Curve]
	if !ok {
		return false, errors.Errorf("curve not recognized [%s]", k.Curve)
	}

	return s.Cmp(halfOrder) != 1, nil

}

func toLowS(k *ecdsa.PublicKey, s *big.Int) (*big.Int, error) {
	lowS, err := isLowS(k, s)
	if err != nil {
		return nil, err
	}

	if !lowS {
		// Set s to N - s that will be then in the lower part of signature space
		// less or equal to half order
		s.Sub(k.Params().N, s)

		return s, nil
	}

	return s, nil
}
