package wrapped

import "context"

type Wrapper func(ctx context.Context, funcname string, f func(ctx context.Context) error) error
type SimpleWrapper func(funcname string, f func() error) error
