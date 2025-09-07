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
    // Harden options with sensible defaults
    if opts.DefaultLimit <= 0 {
        opts.DefaultLimit = DefaultOptions().DefaultLimit
    }
    if opts.MaxLimit <= 0 {
        opts.MaxLimit = DefaultOptions().MaxLimit
    }
    if opts.LogLevel == "" {
        opts.LogLevel = DefaultOptions().LogLevel
    }
    return &Pager{
        opts:   opts,
        logger: newDefaultLogger(opts.LogLevel),
    }
}

// SetLogger replaces the pager's logger. Nil is ignored. Returns the pager for chaining.
func (p *Pager) SetLogger(l Logger) *Pager {
    if l != nil { p.logger = l }
    return p
}
