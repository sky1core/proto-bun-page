package pager

type Options struct {
    DefaultLimit int
    MaxLimit     int
    LogLevel     string
    AllowedOrderKeys []string
    DefaultOrderSpecs []OrderSpec
}

func DefaultOptions() *Options {
	return &Options{
		DefaultLimit: 20,
		MaxLimit:     100,
		LogLevel:     "warn",
	}
}

type Pager struct {
    opts   *Options
    logger Logger
}

func New(opts *Options) *Pager {
    if opts == nil {
        opts = DefaultOptions()
    }
    return &Pager{
        opts:   opts,
        logger: newDefaultLogger(opts.LogLevel),
    }
}
