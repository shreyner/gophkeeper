package main

import (
	"fmt"
	"log"

	"github.com/c-bata/go-prompt"
	"github.com/shreyner/gophkeeper/internal/client/app"
	"github.com/shreyner/gophkeeper/internal/client/command"
	"github.com/shreyner/gophkeeper/internal/client/pkg/promptcmd"
	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultcrypt"
	"github.com/shreyner/gophkeeper/internal/client/state"
	"github.com/shreyner/gophkeeper/internal/client/vaultclient"
	pb "github.com/shreyner/gophkeeper/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Build version: %v\nBuild date: %v\nBuild commit: %v\n", buildVersion, buildDate, buildCommit)

	conn, err := grpc.Dial(":3200", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatal(err)
		return
	}

	defer conn.Close()

	gophKeeperClient := pb.NewGophkeeperClient(conn)

	application := app.New()
	appState := state.New()
	vclient := vaultclient.New(appState, gophKeeperClient)
	_ = vaultcrypt.New() // vcrypt

	commands := command.NewCommands(application, vclient)

	promptcmd := promptcmd.New(
		commands,
	)

	t := prompt.New(
		promptcmd.Executor,
		promptcmd.Completer,
		prompt.OptionTitle("Gophkeeper"),
		prompt.OptionPrefix("> "),
		prompt.OptionSetExitCheckerOnInput(promptcmd.ExitChecker),
	)

	t.Run()

	//requestLoginCtxTimeout, cancelLoginRequest := context.WithTimeout(ctx, time.Second)
	//defer cancelLoginRequest()
	//
	//requestLogin := pb.LoginRequest{
	//	Login:    "Alex2",
	//	Password: "1234",
	//}
	//
	//loginResponse, err := gophKeeperClient.Login(requestLoginCtxTimeout, &requestLogin)
	//
	//if err != nil {
	//	fmt.Println("Error reposne", err)
	//	return
	//}
	//
	//fmt.Println("Token: ", loginResponse.AuthToken)
	//
	//md := metadata.New(map[string]string{
	//	"token": loginResponse.AuthToken,
	//})
	//
	//contextRequest := metadata.NewOutgoingContext(ctx, md)
	//
	//requestCheckAuthTimeout, cancelCheckAuthRequest := context.WithTimeout(contextRequest, 5*time.Second)
	//defer cancelCheckAuthRequest()
	//
	//responseCheckAuth, err := gophKeeperClient.CheckAuth(requestCheckAuthTimeout, new(empty.Empty))
	//
	//if err != nil {
	//	fmt.Println("Error responseCheckAuth", err)
	//	return
	//}
	//
	//fmt.Println("Response CheckAuth", responseCheckAuth.Message)
	//
	//randomBytes := make([]byte, 32)
	//_, err = rand.Read(randomBytes)
	//if err != nil {
	//	fmt.Println("Error generate randomBytes")
	//	return
	//}

	//requestCheckVault, cancelCheckVaultRequest := context.WithTimeout(contextRequest, 5*time.Second)
	//defer cancelCheckVaultRequest()
	//
	//requestVaultCreate := pb.VaultCreateRequest{Vault: randomBytes}
	//
	//responseVaultCreate, err := gophKeeperClient.VaultCreate(requestCheckVault, &requestVaultCreate)
	//
	//if err != nil {
	//	fmt.Println("Error responseVaultCreate", err)
	//	return
	//}
	//
	//fmt.Println("Response responseVaultCreate", responseVaultCreate.Id)

	//requestCheckVault, cancelCheckVaultRequest := context.WithTimeout(contextRequest, 10000*time.Second)
	//defer cancelCheckVaultRequest()
	//
	//requestVaultCreate := pb.VaultUpdateRequest{
	//	Id:      "21b498ef-db70-4be7-a7c4-52e637e8132f",
	//	Vault:   randomBytes,
	//	Version: 0,
	//}
	//
	//responseVaultUpdate, err := gophKeeperClient.VaultUpdate(requestCheckVault, &requestVaultCreate)
	//
	//if err != nil {
	//	fmt.Println("Error responseVaultCreate", err)
	//	return
	//}
	//
	//fmt.Println("Response responseVaultCreate", responseVaultUpdate.Version)

	//requestCheckVault, cancelCheckVaultRequest := context.WithTimeout(contextRequest, 10000*time.Second)
	//defer cancelCheckVaultRequest()
	//
	//requestVaultCreate := pb.VaultDeleteRequest{
	//	Id:      "7d17607f-92b1-490b-bc23-f396b2ea7935",
	//	Version: 2,
	//}
	//
	//responseVaultUpdate, err := gophKeeperClient.VaultDelete(requestCheckVault, &requestVaultCreate)
	//
	//if err != nil {
	//	fmt.Println("Error responseVaultCreate", err)
	//	return
	//}
	//
	//fmt.Println("Response responseVaultCreate", responseVaultUpdate.String())

	//requestCheckVault, cancelCheckVaultRequest := context.WithTimeout(contextRequest, 10000*time.Second)
	//defer cancelCheckVaultRequest()
	//
	//requestVaultCreate := pb.VaultSyncRequest{
	//	VaultVersions: []*pb.VaultSyncRequest_VaultVersion{
	//		{
	//			Id:      "7934fcb9-a503-4a0d-94b2-3c5aa84f169e",
	//			Version: 3,
	//		},
	//		{
	//			Id:      "21b498ef-db70-4be7-a7c4-52e637e8132f",
	//			Version: 1,
	//		},
	//	},
	//}
	//
	//responseVaultUpdate, err := gophKeeperClient.VaultSync(requestCheckVault, &requestVaultCreate)
	//
	//if err != nil {
	//	fmt.Println("Error responseVaultCreate", err)
	//	return
	//}
	//
	//fmt.Println("Response responseVaultCreate", len(responseVaultUpdate.UpdatedVaults))
	//for _, vault := range responseVaultUpdate.UpdatedVaults {
	//	fmt.Println(vault.Id, vault.Version, vault.IsDeleted)
	//}

	//vcrypt := vaultcrypt.New()
	//
	//if err := vcrypt.SetKey([]byte("123")); err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//
	//encryptedData, err := vcrypt.Encrypt([]byte("Hello world"))
	//
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//
	//fmt.Println(hex.EncodeToString(encryptedData))
	//
	//decrypted, err := vcrypt.Decrypt(encryptedData)
	//
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//
	//fmt.Println(string(decrypted))

}
