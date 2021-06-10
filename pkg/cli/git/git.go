// Package git is intended for git access from FuseML client
package git

import (
	"fmt"

	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"time"

	git "github.com/go-git/go-git/v5"
	gitobject "github.com/go-git/go-git/v5/plumbing/object"
	"github.com/pkg/errors"

	config "github.com/fuseml/fuseml-core/pkg/core/config"
	dircopy "github.com/otiai10/copy"
)

// check if the location is indeed pointing to local directory and if so, copy it to target dir
func prepareLocalDirectory(location, target string) error {
	info, err := os.Stat(location)
	if os.IsNotExist(err) {
		return err
	}
	if !info.IsDir() {
		return errors.New(fmt.Sprintf("input path (%s) is not a directory", location))
	}

	err = dircopy.Copy(location, target)
	if err != nil {
		return errors.Wrap(err, "can't copy source directory "+location+" to "+target)
	}
	return nil
}

func fetchRemoteRepository(path, target string) error {

	_, err := git.PlainClone(target, false, &git.CloneOptions{
		URL: path,
	})
	if err != nil {
		return errors.Wrap(err, "failed fetching remote repository")
	}
	return nil
}

// Push the code from local dir to remote repo
// If username or password is not provided, use default values
func Push(org, name, location, gitURL string, uname, pass *string, debug bool) error {
	log.Printf("Pushing the code to the git repository...")

	tmpDir, err := ioutil.TempDir("", "codeset-source")
	if err != nil {
		return errors.Wrap(err, "can't create temp directory "+tmpDir)
	}
	defer os.Remove(tmpDir)

	// if location is URL pointing to git repo, clone the content localy
	loc, err := url.Parse(location)
	if err == nil && loc.IsAbs() && loc.Scheme != "" && loc.Host != "" {
		if err := fetchRemoteRepository(location, tmpDir); err != nil {
			return err
		}
	} else {
		if err := prepareLocalDirectory(location, tmpDir); err != nil {
			return err
		}
	}

	// ignore the fact that the source dir might also be a git repository
	os.RemoveAll(filepath.Join(tmpDir, ".git"))

	// prepare the full URL for the remote (target) repository
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

	u.User = url.UserPassword(username, password)

	// Clone new repository so we can push new content
	cloneDir, err := ioutil.TempDir("", "codeset-clone")
	if err != nil {
		return errors.Wrap(err, "can't create temp directory "+cloneDir)
	}
	defer os.Remove(cloneDir)

	r, err := git.PlainClone(cloneDir, false, &git.CloneOptions{
		URL: u.String(),
	})

	if err != nil {
		return errors.Wrap(err, "failed cloning repository")
	}

	w, err := r.Worktree()
	if err != nil {
		return errors.Wrap(err, "failed fetching worktree of repository")
	}

	err = w.RemoveGlob("*")
	if err != nil {
		return errors.Wrap(err, "failed marking existing files for removal")
	}

	// tmpDir has .git removed, so we can move just data to the real repository clone
	err = dircopy.Copy(tmpDir, cloneDir)
	if err != nil {
		return errors.Wrap(err, "failed copying directory content")
	}

	err = w.AddWithOptions(&git.AddOptions{All: true})
	if err != nil {
		return errors.Wrap(err, "failed adding new directory content")
	}
	_, err = w.Commit(
		fmt.Sprintf("New codeset update from %s", location),
		&git.CommitOptions{
			Author: &gitobject.Signature{
				Name:  "FuseML core user",
				Email: "fuseml-core@fuseml",
				When:  time.Now(),
			},
		})
	if err != nil {
		return errors.Wrap(err, "failed commiting changes")
	}

	pushOpts := &git.PushOptions{}
	if debug {
		pushOpts.Progress = os.Stdout
	}

	err = r.Push(pushOpts)
	if err != nil {
		return errors.Wrap(err, "failed pushing commits")
	}
	return nil
}
