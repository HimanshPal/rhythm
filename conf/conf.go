package conf

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

type Conf struct {
	API         API
	Storage     Storage
	Coordinator Coordinator
	Secrets     Secrets
	Verbose     bool
	Mesos       Mesos
}

type API struct {
	Address string
	Auth    APIAuth
}

const (
	APIAuthBackendGitLab = "gitlab"
	APIAuthBackendNone   = "none"
)

type APIAuth struct {
	Backend string
	GitLab  APIAuthGitLab
}

type APIAuthGitLab struct {
	BaseURL string
}

type Storage struct {
	Backend   string
	ZooKeeper StorageZK
}

const StorageBackendZK = "zookeeper"

type StorageZK struct {
	BasePath string
	Servers  []string
	Timeout  time.Duration
}

const CoordinatorBackendZK = "zookeeper"

type Coordinator struct {
	Backend   string
	ZooKeeper CoordinatorZK
}

type CoordinatorZK struct {
	BasePath    string
	ElectionDir string
	Servers     []string
	Timeout     time.Duration
}

const SecretsBackendVault = "vault"

type Secrets struct {
	Backend string
	Vault   SecretsVault
}

type SecretsVault struct {
	Token   string
	Address string
	Timeout time.Duration
}

type Mesos struct {
	Auth            MesosAuth
	BaseURL         string
	Checkpoint      bool
	FailoverTimeout time.Duration
	Hostname        string
	User            string
	WebUiURL        string
	Principal       string
}

const (
	MesosAuthTypeBasic = "basic"
	MesosAuthTypeNone  = "none"
)

type MesosAuth struct {
	Type  string
	Basic MesosAuthBasic
}

type MesosAuthBasic struct {
	Username string
	Password string
}

func New(path string) (*Conf, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var conf = &Conf{
		API: API{
			Address: "localhost:8000",
			Auth: APIAuth{
				Backend: APIAuthBackendNone,
			},
		},
		Storage: Storage{
			Backend: StorageBackendZK,
			ZooKeeper: StorageZK{
				Servers:  []string{"127.0.0.1"},
				Timeout:  10000, // 10s
				BasePath: "/rhythm",
			},
		},
		Coordinator: Coordinator{
			Backend: CoordinatorBackendZK,
			ZooKeeper: CoordinatorZK{
				Servers:     []string{"127.0.0.1"},
				Timeout:     10000, // 10s
				BasePath:    "/rhythm",
				ElectionDir: "election",
			},
		},
		Secrets: Secrets{
			Backend: SecretsBackendVault,
			Vault: SecretsVault{
				Timeout: 3000, // 3s
			},
		},
		Verbose: false,
		Mesos: Mesos{
			BaseURL:         "http://127.0.0.1:5050",
			FailoverTimeout: time.Hour * 24 * 7,
		},
	}
	err = json.Unmarshal(file, conf)
	conf.Secrets.Vault.Timeout *= time.Millisecond
	conf.Storage.ZooKeeper.Timeout *= time.Millisecond
	conf.Coordinator.ZooKeeper.Timeout *= time.Millisecond
	if err != nil {
		return nil, err
	}
	return conf, nil
}
