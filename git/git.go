package git

import (
	"fmt"
	gogit "github.com/go-git/go-git"
	"github.com/go-git/go-git/plumbing"
	"github.com/go-git/go-git/plumbing/transport"
	"github.com/go-git/go-git/plumbing/transport/ssh"
	"github.com/op/go-logging"
	"io/ioutil"
	"os"
)

var log = logging.MustGetLogger("gitops")

type DeplyomentConfig struct {
	Remote         string
	Repository     *gogit.Repository
	SshPrivKeyFile string
}

func Open(path string, remote string, sshPrivKeyFile string) (*DeplyomentConfig, error) {
	repo, err := gogit.PlainOpen(path)
	if nil != err {
		log.Errorf("Unable to open '%s' as git repository: '%s'", path, err)
		return nil, err
	}

	return &DeplyomentConfig{
		Remote:         remote,
		Repository:     repo,
		SshPrivKeyFile: sshPrivKeyFile,
	}, nil
}

func (deploymentConfig *DeplyomentConfig) Apply(ref *plumbing.Reference) (*plumbing.Reference, error) {
	currentRef, err := deploymentConfig.LocalRef()
	if nil != err {
		return nil, err
	}

	worktree, err := deploymentConfig.Repository.Worktree()
	if nil != err {
		log.Errorf("Unable to get worktree: '%s'", err)
		return nil, err
	}

	err = worktree.Reset(&gogit.ResetOptions{
		Commit: ref.Hash(),
		Mode:   gogit.HardReset,
	})
	if nil != err {
		log.Errorf(
			"Unable to update %s (%s) to %s (%s): ''",
			currentRef.Name().Short(),
			currentRef.Hash().String(),
			ref.Name().Short(),
			ref.Hash().String(),
			err,
		)
		return nil, err
	}

	log.Infof(
		"Successfully updated %s (%s) to %s (%s)",
		currentRef.Name().Short(),
		currentRef.Hash().String(),
		ref.Name().Short(),
		ref.Hash().String(),
	)
	return currentRef, nil
}

func (deploymentConfig *DeplyomentConfig) LocalRef() (*plumbing.Reference, error) {
	localRef, err := deploymentConfig.Repository.Head()
	if nil != err {
		log.Errorf("Unable to determine local ref: '%s'", err)
		return nil, err
	}
	return localRef, nil
}

func (deploymentConfig *DeplyomentConfig) RemoteRef() (*plumbing.Reference, error) {
	localRef, err := deploymentConfig.LocalRef()
	if nil != err {
		return nil, err
	}
	refspec := fmt.Sprintf("refs/remotes/%s/%s", deploymentConfig.Remote, localRef.Name().Short())
	remoteRef, err := deploymentConfig.Repository.Reference(plumbing.ReferenceName(refspec), true)
	if nil != err {
		log.Errorf("Unable to determine remote ref of '%s': '%s'", refspec, err)
		return nil, err
	}
	return remoteRef, nil
}

func (deploymentConfig *DeplyomentConfig) Fetch() (bool, error) {
	err := deploymentConfig.Repository.Fetch(&gogit.FetchOptions{
		RemoteName: deploymentConfig.Remote,
		Auth:       deploymentConfig.createAuth(),
		Progress:   os.Stdout,
		Force:      true,
	})

	if err == gogit.NoErrAlreadyUpToDate {
		log.Debugf("No changes detected on remote '%s'.", deploymentConfig.Remote)
		return false, nil
	} else if nil != err {
		log.Warningf("An error occurred during gogit fetch: %s", err)
	}
	return true, err
}

func (deploymentConfig *DeplyomentConfig) Update() (*plumbing.Reference, error) {
	hasUpdate, err := deploymentConfig.Fetch()
	if nil != err {
		return nil, err
	}

	if !hasUpdate {
		return deploymentConfig.LocalRef()
	}

	ref, err := deploymentConfig.RemoteRef()
	if nil != err {
		return nil, err
	}
	return deploymentConfig.Apply(ref)
}

func (deploymentConfig *DeplyomentConfig) createAuth() transport.AuthMethod {
	sshKey, err := ioutil.ReadFile(deploymentConfig.SshPrivKeyFile)
	if nil != err {
		log.Warningf("Unable to open SSH privkey '%s': '%s", deploymentConfig.SshPrivKeyFile, err)
		return nil
	}
	publicKey, err := ssh.NewPublicKeys("git", sshKey, "")
	if nil != err {
		log.Warningf("Unable to generate SSH pubkey with '%s': '%s", deploymentConfig.SshPrivKeyFile, err)
	}
	return publicKey
}
