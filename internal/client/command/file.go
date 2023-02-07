package command

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultclient"
	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultcrypt"
	"github.com/shreyner/gophkeeper/internal/client/storage"
)

type FileCommand struct {
	vclient     *vaultclient.Client
	vaultCrypt  *vaultcrypt.VaultCrypt
	fileStorage *storage.FileVaultStorage
}

func NewFileCommand(
	vclient *vaultclient.Client,
	vaultCrypt *vaultcrypt.VaultCrypt,
	fileStorage *storage.FileVaultStorage,
) *FileCommand {
	command := FileCommand{
		vclient:     vclient,
		vaultCrypt:  vaultCrypt,
		fileStorage: fileStorage,
	}

	return &command
}

func (c *FileCommand) RunView(_ context.Context, _ []string) {
	arr := c.fileStorage.GetAll()

	for _, model := range arr {
		fmt.Printf(
			"ID: %v, IsUpdate: %v, IsDeleted: %v, FileName: %v\n",
			model.ID,
			model.IsUpdate && model.IsNew,
			model.IsDelete,
			model.GetFileName(),
		)
	}
}

func (c *FileCommand) RunUpload(ctx context.Context, args []string) {
	if len(args) < 1 {
		fmt.Println("incorrect login and password")
		return
	}

	filePath := args[0]

	if len(filePath) < 1 {
		fmt.Println("incorrect path to fileOut")
		return
	}

	fileOut, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fileOut.Close()

	err = c.fileStorage.UploadFile(ctx, fileOut)
	if err != nil {
		fmt.Println(err)
		return
	}

	//reader, writer := io.Pipe()
	//defer reader.Close()
	//
	//w, err := c.vaultCrypt.EncryptStream(writer, []byte("equnPrKfGSYVSRxKNGluRthXe71KQ5q35mTu6QLa"))
	//
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//
	//go func() {
	//	defer writer.Close()
	//	_, err = io.Copy(w, fileOut)
	//
	//	if err != nil {
	//		fmt.Println("Error: ", err)
	//	}
	//}()
	//
	//result, err := c.vclient.VaultUpload(ctx, reader)
	//
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//
	//fmt.Println(result)

	//body, err := c.vclient.VaultDownload(ctx, "http://127.0.0.1:9000/vault/ZY24ImVBA4ARL0zrCx0b5IR7i_TvtmjG7gk2")
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//defer body.Close()
	//
	//r, err := c.vaultCrypt.DecryptStream(body, []byte("equnPrKfGSYVSRxKNGluRthXe71KQ5q35mTu6QLa"))
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//
	//fileIn, err := os.Create("./decripted.png")
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//defer fileIn.Sync()
	//defer fileIn.Close()
	//
	//_, err = io.Copy(fileIn, r)
}

func (c *FileCommand) RunDownload(ctx context.Context, args []string) {
	if len(args) < 2 {
		fmt.Println("incorrect login and password")
		return
	}

	fileID, filePath := args[0], args[1]

	if len(filePath) < 1 {
		fmt.Println("incorrect path to fileOut")
		return
	}

	ID, err := strconv.ParseUint(fileID, 10, 32)
	if err != nil {
		fmt.Println("Invalid ID")
		return
	}

	file, err := os.Create(filePath)

	if err != nil {
		fmt.Println(err)
		return
	}

	defer file.Close()

	err = c.fileStorage.DownloadFile(ctx, uint32(ID), file)

	if err != nil {
		fmt.Println(err)
		return
	}
}
