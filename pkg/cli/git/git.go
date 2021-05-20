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
// If username or password is not provided, use default values, but password
// provided from env variable GITEA_PROJECT_PASSWORD has the highest priority.
func Push(org, name, location, gitURL string, uname, pass *string, debug bool) error {
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
	username := config.DefaultUserName(org)
	password := config.DefaultUserPassword

	if uname != nil {
		username = *uname
	}
	if pass != nil {
		password = *pass
	}
	envPassword, exists := os.LookupEnv("FUSEML_PROJECT_PASSWORD")
	if exists {
		password = envPassword
	}

	u.User = url.UserPassword(username, password)

	// TODO: use some real API instead...
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf(`
cd "%s"
rm -rf .git
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
