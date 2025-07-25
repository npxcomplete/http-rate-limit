//go:build testing

package clients

import (
	"context"
	"github.com/golang/mock/gomock"
)

type MockMetadataClient struct {
	Ctrl                       *gomock.Controller
	GetMetadataWithContextFunc func(ctx context.Context, path string) (string, error)
}

func (m *MockMetadataClient) GetMetadataWithContext(ctx context.Context, path string) (string, error) {
	if m.GetMetadataWithContextFunc != nil {
		return m.GetMetadataWithContextFunc(ctx, path)
	}
	return "", nil
}
