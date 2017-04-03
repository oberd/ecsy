package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/kardianos/osext"
	"github.com/spf13/cobra"
)

type release struct {
	Version string `json:"tag_name"`
}

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "self-update",
	Short: "Update the ecsy cli binary on your system",
	Long:  `Update the ecsy cli binary on your system`,
	Run: func(cmd *cobra.Command, args []string) {
		executable, err := osext.Executable()
		failOnError(err, "Could not find executable")
		response, err := http.Get("https://api.github.com/repos/oberd/ecsy/releases")
		failOnError(err, "Fail finding newest release")
		body, err := ioutil.ReadAll(response.Body)
		failOnError(err, "Problem parsing json from github")
		response.Body.Close()
		releases := make([]release, 0)
		json.Unmarshal(body, &releases)
		suffix := "darwin-amd64"
		if runtime.GOOS == "darwin" {
			suffix = "darwin-amd64"
		}
		var version string
		if len(args) == 0 {
			version = releases[0].Version
		} else {
			version = args[1]
		}
		url := fmt.Sprintf("https://github.com/oberd/ecsy/releases/download/%s/ecsy-%s-%s", version, version, suffix)
		tmp, err := ioutil.TempFile("", "ecsy")
		defer os.Remove(tmp.Name())
		failOnError(err, "Problem allocating temp file")
		fmt.Printf("Downloading version %s...\n", version)
		resp, err := http.Get(url)
		failOnError(err, "Problem downloading binary")
		defer resp.Body.Close()
		failOnError(err, "Problem opening temp file")
		_, err = io.Copy(tmp, resp.Body)
		failOnError(err, "Problem writing temp file")
		err = os.Chmod(tmp.Name(), 0755)
		failOnError(err, "Problem changing permissions")
		mover := exec.Command("mv", tmp.Name(), executable)
		out, err := mover.Output()
		failOnError(err, fmt.Sprintf("Problem updating binary (possibly permissions issue), %s", out))
		fmt.Printf("Successfully updated ecsy to version %s. Enjoy!\n", version)
	},
}

func init() {
	RootCmd.AddCommand(updateCmd)
}
