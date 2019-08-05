package wallet

import (
	"os"
	"fmt"
	"errors"
	"strings"
	"io/ioutil"

	. "github.com/elastos/Elastos.ELA.Client.SideChain/wallet"

	. "github.com/elastos/Elastos.ELA/common"
	"github.com/elastos/Elastos.ELA/crypto"
	"github.com/urfave/cli"
)

func addAccount(context *cli.Context, wallet Wallet, content string) error {
	// Get public key string
	if _, err := os.Stat(content); err == nil { // if content is a file

		file, err := os.OpenFile(content, os.O_RDONLY, 0666)
		if err != nil {
			return errors.New("open public key(s) file failed")
		}
		rawData, err := ioutil.ReadAll(file)
		if err != nil {
			return errors.New("read public key(s) file failed")
		}
		content = string(rawData)
	}

	var err error
	var programHash *Uint168
	if !strings.Contains(content, ",") { // single public key
		publicKey, err := getPublicKey(content)
		if err != nil {
			return err
		}

		programHash, err = wallet.AddStandardAccount(publicKey)
		if err != nil {
			return err
		}
	} else { // multi public keys
		publicKeys, err := getPublicKeys(content)
		if err != nil {
			return err
		}

		if len(publicKeys) < MinMultiSignKeys {
			return errors.New(fmt.Sprint("multi sign account require at lest ", MinMultiSignKeys, " public keys"))
		}

		// Get M value
		M := context.Int("m")
		if M == 0 { // Use default M greater than half
			M = len(publicKeys)/2 + 1
		}
		if M < len(publicKeys)/2+1 || M > len(publicKeys) {
			return errors.New("M must be greater than half number of public keys, less than number of public keys")
		}

		programHash, err = wallet.AddMultiSignAccount(M, publicKeys...)
		if err != nil {
			return err
		}
	}

	// When add a new address, reset stored height to trigger synchronize blocks.
	wallet.CurrentHeight(ResetHeightCode)

	addrs, err := wallet.GetAddresses()
	if err != nil || len(addrs) == 0 {
		return errors.New("fail to load wallet addresses")
	}

	return ShowAccounts(addrs, programHash, wallet)
}

func getPublicKey(content string) (*crypto.PublicKey, error) {
	// Content can not be empty
	if content == "" {
		return nil, errors.New("content should be the public key file path or public key string")
	}

	// Get public key string
	if _, err := os.Stat(content); err == nil { // if content is a file

		file, err := os.OpenFile(content, os.O_RDONLY, 0666)
		if err != nil {
			return nil, errors.New("open public key file failed")
		}
		rawData, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, errors.New("read public key file failed")
		}
		content = string(rawData)
	}

	// Get public key
	keyBytes, err := HexStringToBytes(strings.TrimSpace(content))
	if err != nil {
		return nil, err
	}
	publicKey, err := crypto.DecodePoint(keyBytes)
	if err != nil {
		return nil, err
	}

	return publicKey, nil
}

func getPublicKeys(content string) ([]*crypto.PublicKey, error) {
	// Content can not be empty
	if content == "" {
		return nil, errors.New("content should be the public key[s] file path or public key strings separated by comma")
	}

	// Get public key strings
	var publicKeyStrings []string
	if _, err := os.Stat(content); err == nil { // if content is a file

		file, err := os.OpenFile(content, os.O_RDONLY, 0666)
		if err != nil {
			return nil, errors.New("open public key file failed")
		}
		rawData, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, errors.New("read public key file failed")
		}
		publicKeyStrings = strings.Split(strings.TrimSpace(string(rawData)), "\n")
	} else {
		publicKeyStrings = strings.Split(strings.TrimSpace(content), ",")
	}

	// Check if have duplicate public keys
	keyMap := map[string]string{}
	for _, publicKeyString := range publicKeyStrings {
		if keyMap[publicKeyString] == "" {
			keyMap[publicKeyString] = publicKeyString
		} else {
			return nil, errors.New(fmt.Sprint("duplicate public key:", publicKeyString))
		}
	}

	// Decode public keys from public key strings
	var publicKeys []*crypto.PublicKey
	for _, v := range publicKeyStrings {
		keyBytes, err := HexStringToBytes(strings.TrimSpace(v))
		if err != nil {
			return nil, err
		}
		publicKey, err := crypto.DecodePoint(keyBytes)
		if err != nil {
			return nil, err
		}
		publicKeys = append(publicKeys, publicKey)
	}

	return publicKeys, nil
}

func deleteAccount(wallet Wallet, address string) error {
	programHash, err := Uint168FromAddress(address)
	if err != nil {
		return err
	}

	err = wallet.DeleteAddress(programHash)
	if err != nil {
		return err
	}

	fmt.Println(address)
	return nil
}
