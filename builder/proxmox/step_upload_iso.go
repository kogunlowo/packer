package proxmox

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// stepUploadISO uploads an ISO file to Proxmox so we can boot from it
type stepUploadISO struct{}

var _ uploader = &proxmox.Client{}

func (s *stepUploadISO) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	client := state.Get("proxmoxClient").(uploader)
	c := state.Get("config").(*Config)

	if !c.shouldUploadISO {
		state.Put("iso_file", c.ISOFile)
		return multistep.ActionContinue
	}

	p := state.Get(downloadPathKey).(string)
	if p == "" {
		err := fmt.Errorf("Path to downloaded ISO was empty")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// All failure cases in resolving the symlink are caught anyway in os.Open
	isoPath, _ := filepath.EvalSymlinks(p)
	r, err := os.Open(isoPath)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	filename := filepath.Base(c.ISOUrls[0])
	err = client.Upload(c.Node, c.ISOStoragePool, "iso", filename, r)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	isoStoragePath := fmt.Sprintf("%s:iso/%s", c.ISOStoragePool, filename)
	state.Put("iso_file", isoStoragePath)
	return multistep.ActionContinue
}

func (s *stepUploadISO) Cleanup(state multistep.StateBag) {
}
