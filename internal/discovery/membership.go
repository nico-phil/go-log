package discovery

import "go.uber.org/zap"

type Membership struct {
	Config
	logger *zap.Logger
}

type Config struct {
}

func New() (*Membership, error) {
	return &Membership{
		logger: zap.L().Named("membership"),
	}, nil
}
