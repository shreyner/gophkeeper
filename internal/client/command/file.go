package command

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultclient"
	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultcrypt"
)

type FileCommand struct {
	vclient    *vaultclient.Client
	vaultCrypt *vaultcrypt.VaultCrypt
}

func NewFileCommand(
	vclient *vaultclient.Client,
	vaultCrypt *vaultcrypt.VaultCrypt,
) *FileCommand {
	command := FileCommand{
		vclient:    vclient,
		vaultCrypt: vaultCrypt,
	}

	return &command
}

func (c *FileCommand) RunUpload(ctx context.Context, args []string) {
	if len(args) < 1 {
		fmt.Println("incorrect login and password")
		return
	}

	filePath := args[0]

	if len(filePath) < 1 {
		fmt.Println("incorrect path to file")
		return
	}

	file, err := os.Open(filePath)

	if err != nil {
		fmt.Println(err)
		return
	}

	defer file.Close()

	//fileIn, err := os.Create("./encrypted-file")
	//
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//
	//w, err := c.vaultCrypt.EncryptStream(fileIn, []byte("equnPrKfGSYVSRxKNGluRthXe71KQ5q35mTu6QLa"))
	//
	//io.Copy(w, file)

	reader, writer := io.Pipe()

	w, err := c.vaultCrypt.EncryptStream(writer, []byte("equnPrKfGSYVSRxKNGluRthXe71KQ5q35mTu6QLa"))

	if err != nil {
		fmt.Println(err)
		return
	}

	go func() {
		io.Copy(w, file)
	}()

	//result, err := c.vclient.VaultUpload(ctx, file)
	//
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//
	//fmt.Println(result)

	result, err := c.vclient.VaultUpload(ctx, reader)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(result)

	//err = c.vaultCrypt.DecryptStream(file, fileIn, []byte("equnPrKfGSYVSRxKNGluRthXe71KQ5q35mTu6QLa"))
	//
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}

}
