package config

import (
	"sync"

	"github.com/rs/zerolog/log"

	libconfig "github.com/lk153/import-gsheet/lib/configs"
	libenv "github.com/lk153/import-gsheet/lib/env"
)

var (
	SDBEnv SDBStrEnv
	cfg    libconfig.Config
	once   sync.Once
)

type SDBStrEnv struct {
	libenv.NVBaseEnv
}

func init() {
	GetCfg()
}

func GetCfg() libconfig.Config {
	once.Do(func() {
		SDBEnv = SDBStrEnv{}
		cfg = libconfig.New()
		if err := libenv.Init(cfg, &SDBEnv); err != nil {
			log.Err(err).Msg("Error while initializing environment variables")
			panic(err)
		}
	})
	if cfg == nil {
		panic("Can not load env configuration")
	}

	return cfg
}
