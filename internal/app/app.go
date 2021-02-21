package app

import "context"

func Start(ctx context.Context) (done chan struct{}) {
	return make(chan struct{})
}
