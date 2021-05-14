// Package git is intended for git access from FuseML client
// It needs GITEA_URL set as environment variable.
package git

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/pkg/errors"

	config "github.com/fuseml/fuseml-core/pkg/core/config"
	dircopy "github.com/otiai10/copy"
)

// Push the code from local dir to remote repo
func Push(org, name, location string) error {
	log.Printf("Pushing the code to the git repository...")

	gitURL, exists := os.LookupEnv("GITEA_URL")
	if !exists {
		return errors.New("Value for gitea URL (GITEA_URL) was not provided")
	}

	tmpDir, err := ioutil.TempDir("", "codeset-source")
	if err != nil {
		return errors.Wrap(err, "can't create temp directory")
	}
	defer os.Remove(tmpDir)
	err = dircopy.Copy(location, tmpDir)
	if err != nil {
		return errors.Wrap(err, "can't copy source directory to temp")
	}

	u, err := url.Parse(gitURL)
	if err != nil {
		return errors.Wrap(err, "Failed to parse gitea url")
	}
	// TODO username+password are created by server action before git push
	// The values could be returned from some POST if we had some init command, but
	// for now let's assume the values are fixed
	username := config.DefaultUserName(org)
	password := config.DefaultUserPassword

	u.User = url.UserPassword(username, password)
	u.Path = path.Join(u.Path, org, name)

	// TODO: use some real API instead...
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf(`
cd "%s"
git init
git config user.name "Fuseml"
git config user.email ci@fuseml
git remote add fuseml "%s"
git fetch --all
git reset --soft fuseml/main
git add --all
git commit --no-gpg-sign -m "pushed at %s"
git push fuseml master:main
`, tmpDir, u.String(), time.Now().Format("20060102150405")))

	_, err = cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(err, "Pushing the code has failed")
	}
	return nil
}
