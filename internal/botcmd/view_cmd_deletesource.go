package botcmd

import "context"

type SourceDeleter interface {
	Delete(ctx context.Context, id int64) error
}
