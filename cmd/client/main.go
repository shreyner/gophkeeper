package main

import (
	"fmt"
	"log"

	"github.com/c-bata/go-prompt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/shreyner/gophkeeper/internal/client/command"
	"github.com/shreyner/gophkeeper/internal/client/pkg/promptcmd"
	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultclient"
	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultcrypt"
	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultsync"
	"github.com/shreyner/gophkeeper/internal/client/state"
	"github.com/shreyner/gophkeeper/internal/client/storage"
	pb "github.com/shreyner/gophkeeper/proto"
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

	appState := state.New()

	vclient := vaultclient.New(appState, gophKeeperClient)
	vcrypt := vaultcrypt.New()

	loginVaultStorage := storage.NewLoginVaultStorage(vcrypt)
	fileVaultStorage := storage.NewFileVaultStorage(vcrypt, vclient)

	err = loginVaultStorage.LoadFromLocalFile("./data/site-login.db")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer func() {
		err = loginVaultStorage.SaveToFile("./data/site-login.db")
		if err != nil {
			log.Println("error saved data to file", err)
			return
		}
	}()
	err = fileVaultStorage.LoadFromLocalFile("./data/file.db")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer func() {
		err = fileVaultStorage.SaveToFile("./data/file.db")
		if err != nil {
			log.Println("error saved data to file", err)
			return
		}
	}()

	vsync := vaultsync.New(
		vcrypt,
		vclient,
		[]vaultsync.StorageSyncer{
			loginVaultStorage,
			fileVaultStorage,
		},
	)

	commands := command.NewCommands(
		vclient,
		vcrypt,
		vsync,
		loginVaultStorage,
		fileVaultStorage,
	)

	if err != nil {
		fmt.Println("Error first sync sync: ", err)
		return
	}

	promptcmd := promptcmd.New(
		appState,
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
}
