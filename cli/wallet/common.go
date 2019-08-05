package wallet

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	walt "github.com/elastos/Elastos.ELA.Client.SideChain/wallet"

	. "github.com/elastos/Elastos.ELA/common"
	"github.com/howeyc/gopass"
)

func GetPassword(password []byte, confirmed bool) ([]byte, error) {
	if len(password) > 0 {
		return []byte(password), nil
	}

	fmt.Print("INPUT PASSWORD:")

	password, err := gopass.GetPasswd()
	if err != nil {
		return nil, err
	}

	if !confirmed {
		return password, nil
	} else {

		fmt.Print("CONFIRM PASSWORD:")

		confirm, err := gopass.GetPasswd()
		if err != nil {
			return nil, err
		}

		if !bytes.Equal(password, confirm) {
			return nil, errors.New("input password unmatched")
		}
	}

	return password, nil
}

func ShowAccountInfo(name string, password []byte) error {
	var err error
	password, err = GetPassword(password, false)
	if err != nil {
		return err
	}

	keyStore, err := walt.OpenKeystore(name, password)
	if err != nil {
		return err
	}

	// print header
	fmt.Printf("%-34s %-66s\n", "ADDRESS", "PUBLIC KEY")
	fmt.Println(strings.Repeat("-", 34), strings.Repeat("-", 66))

	// print account
	publicKey := keyStore.GetPublicKey()
	publicKeyBytes, _ := publicKey.EncodePoint(true)
	fmt.Printf("%-34s %-66s\n", keyStore.Address(), BytesToHexString(publicKeyBytes))
	// print divider line
	fmt.Println(strings.Repeat("-", 34), strings.Repeat("-", 66))

	return nil
}

func SelectAccount(wallet walt.Wallet) (string, error) {
	addrs, err := wallet.GetAddresses()
	if err != nil || len(addrs) == 0 {
		return "", errors.New("fail to load wallet addresses")
	}

	// only one address return it
	if len(addrs) == 1 {
		return addrs[0].Address, nil
	}

	// show accounts
	err = ShowAccounts(addrs, nil, wallet)
	if err != nil {
		return "", err
	}

	// select address by index input
	fmt.Println("Please input the address INDEX you want to use and press enter")

	index := -1
	for index == -1 {
		index = getInput(len(addrs))
	}

	return addrs[index].Address, nil
}

func ShowAccounts(addrs []*walt.Address, newAddr *Uint168, wallet walt.Wallet) error {
	// print header
	fmt.Printf("%5s %34s %-20s%22s %6s\n", "INDEX", "ADDRESS", "BALANCE", "(LOCKED)", "TYPE")
	fmt.Println("-----", strings.Repeat("-", 34), strings.Repeat("-", 42), "------")

	currentHeight := wallet.CurrentHeight(walt.QueryHeightCode)
	for i, addr := range addrs {
		available := Fixed64(0)
		locked := Fixed64(0)
		UTXOs, err := wallet.GetAddressUTXOs(addr.ProgramHash)
		if err != nil {
			return errors.New("get " + addr.Address + " UTXOs failed")
		}
		for _, utxo := range UTXOs {
			if utxo.LockTime < currentHeight {
				available += *utxo.Amount
			} else {
				locked += *utxo.Amount
			}
		}
		var format = "%5d %34s %-20s%22s %6s\n"
		if newAddr != nil && newAddr.IsEqual(*addr.ProgramHash) {
			format = "\033[0;32m" + format + "\033[m"
		}

		fmt.Printf(format, i+1, addr.Address, available.String(), "("+locked.String()+")", addr.TypeName())
		fmt.Println("-----", strings.Repeat("-", 34), strings.Repeat("-", 42), "------")
	}

	return nil
}

func getInput(max int) int {
	fmt.Print("INPUT INDEX: ")
	input, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		fmt.Println("read input falied")
		return -1
	}

	// trim space
	input = strings.TrimSpace(input)

	index, err := strconv.ParseInt(input, 10, 32)
	if err != nil {
		fmt.Println("please input a positive integer")
		return -1
	}

	if int(index) > max {
		fmt.Println("INDEX should between 1 ~", max)
		return -1
	}

	return int(index) - 1
}
