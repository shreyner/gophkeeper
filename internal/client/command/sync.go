package command

import (
	"fmt"

	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultsync"
	"golang.org/x/net/context"
)

type SyncCommand struct {
	vsync *vaultsync.VaultSync
}

func NewSyncCommand(
	vsync *vaultsync.VaultSync,
) *SyncCommand {
	command := SyncCommand{
		vsync: vsync,
	}

	return &command
}

func (c *SyncCommand) Run(ctx context.Context, args []string) {
	fmt.Println("Start full sync...")

	err := c.vsync.Sync()

	if err != nil {
		fmt.Println("Error first sync storage: ", err)
		return
	}

	fmt.Println("Success full sync 🎉")

}
