/*
Copyright © 2021 SUSE LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package action

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/zloylos/grsync"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type InstallAction struct {
	Device string
	Distro string  // Source of data to install
	Target string  // Target to install
}

func NewInstallAction(device string) *InstallAction{
	return &InstallAction{Device: device}
}

func (i InstallAction) Run() error {
	fmt.Println("InstallAction called")
	// Rough steps (then they have multisteps inside)
	// Remember to hook the yip hooks (before-install, after-install-chroot, after-install)
	// Check device valid
	// partition device
	// check source to install
	// install Active
	// install grub
	// Relabel SELinux
	// Unmount everything
	// install Recovery
	// install Secondary
	// Rebrand
	// ????
	// profit!
	return nil
}

func (i InstallAction) doCopy() error {
	fmt.Printf("Copying cOS..")
	// 1 - rsync all the system from source to target
	task := grsync.NewTask(
		i.Distro,
		i.Target,
		grsync.RsyncOptions{
			Quiet: true,
			Archive: true,
			XAttrs: true,
			ACLs: true,
			Exclude: []string{"mnt", "proc", "sys", "dev", "tmp"},
		},
	)
	go func() {
		for {
			state := task.State()
			fmt.Printf(
				"progress: %.2f / rem. %d / tot. %d / sp. %s \n",
				state.Progress,
				state.Remain,
				state.Total,
				state.Speed,
			)
			<- time.After(time.Second)
		}
	}()
	if err := task.Run(); err != nil {
		return err
	}
	// 2 - if we got a cloud config file get it and store in the target
	if viper.GetString("cloudInit") != "" {
		customConfig := fmt.Sprintf("%s/oem/99_custom.yaml", i.Target)

		if err := i.getUrl(viper.GetString("cloudInit"), customConfig); err != nil {
			return err
		}

		if err := os.Chmod(customConfig, 0600); err != nil {
			return err
		}
	}
	return nil
}

func (i InstallAction) getUrl(url string, destination string) error {
	var source io.Reader
	var err error

	switch {
	case strings.HasPrefix(url, "http"):
	case strings.HasPrefix(url, "ftp"):
	case strings.HasPrefix(url, "tftp"):
		fmt.Printf("Downloading from %s to %s", url, destination)
		resp, err := http.Get(url)
		if err != nil {return err}
		source = resp.Body
		defer resp.Body.Close()
	default:
		fmt.Printf("Copying from %s to %s", url, destination)
		file, err := os.Open(url)
		if err != nil {return err}
		source = file
		defer file.Close()
	}

	dest, err := os.Create(destination)
	defer dest.Close()
	if err != nil {return err}
	nBytes, err := io.Copy(dest, source)
	if err != nil {return err}
	fmt.Printf("Copied %d bytes", nBytes)

	return nil
}