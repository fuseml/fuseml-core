// Package git is intended for git access from FuseML client
package git

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"time"

	"github.com/pkg/errors"

	config "github.com/fuseml/fuseml-core/pkg/core/config"
	dircopy "github.com/otiai10/copy"
)

// Push the code from local dir to remote repo
func Push(org, name, location, gitURL string, debug bool) error {
	log.Printf("Pushing the code to the git repository...")

	tmpDir, err := ioutil.TempDir("", "codeset-source")
	if err != nil {
		return errors.Wrap(err, "can't create temp directory "+tmpDir)
	}
	defer os.Remove(tmpDir)
	err = dircopy.Copy(location, tmpDir)
	if err != nil {
		return errors.Wrap(err, "can't copy source directory "+location+" to "+tmpDir)
	}

	u, err := url.Parse(gitURL)
	if err != nil {
		return errors.Wrap(err, "Failed to parse git url")
	}
	// TODO username+password are created by server action before git push
	// The values could be returned from some POST if we had some init command, but
	// for now let's assume the values are fixed
	username := config.DefaultUserName(org)
	password := config.DefaultUserPassword

	u.User = url.UserPassword(username, password)

	// TODO: use some real API instead...
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf(`
cd "%s"
git init
git config user.name "Fuseml"
git config user.email cli@fuseml
git remote add fuseml "%s"
git fetch --all
git reset --soft fuseml/main
git add --all
git commit --no-gpg-sign -m "pushed at %s"
git push fuseml master:main
`, tmpDir, u.String(), time.Now().Format("20060102150405")))

	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(err,
			fmt.Sprintf("Pushing the code has failed with:\n%s\n",
				string(stdout)))
	}
	if debug {
		fmt.Println(string(stdout))
	}
	return nil
}