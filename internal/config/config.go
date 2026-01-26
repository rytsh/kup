package config

import (
	"context"

	"github.com/rakunlabs/chu"
	"github.com/rakunlabs/chu/loader"
	"github.com/rakunlabs/chu/loader/loaderenv"
	"github.com/rakunlabs/chu/loader/loaderfile"
)

type Config struct {
}

func Load(ctx context.Context) (*Config, error) {
	cfg := &Config{}
	if err := chu.Load(
		ctx, "kup", cfg,
		chu.WithDisableLoader(loader.NameHTTP),
		chu.WithLoaderOption(loaderenv.New(
			loaderenv.WithPrefix("KUP_"),
			loaderenv.WithEnvFile(),
			loaderenv.WithCheckConfigEnvFile(false),
		)),
		chu.WithLoaderOption(loaderfile.New(
			loaderfile.WithCheckCurrentFolder(false),
			loaderfile.WithCheckEnv(false),
			loaderfile.WithFolders("~/.config/kup"),
		)),
	); err != nil {
		return nil, err
	}

	return cfg, nil
}
