package util

import (
	"context"
	"fmt"

	"github.com/stretchr/testify/mock"

	"gitlab.prod.gitlab.aws.cloud.huk.de/huk/ds/core/golang-code-challenge/src/internal/domain/model"
	"gitlab.prod.gitlab.aws.cloud.huk.de/huk/ds/core/golang-code-challenge/src/internal/domain/service"
)

var _ service.GreetingService = (*GreetingServiceMock)(nil)

type GreetingServiceMock struct {
	mock.Mock
}

func NewGreetingServiceMock() *GreetingServiceMock {
	return new(GreetingServiceMock)
}

func (m *GreetingServiceMock) Greet(ctx context.Context, name string) (*model.GreetingResponse, error) {
	args := m.Called(ctx, name)

	result, ok := args.Get(0).(*model.GreetingResponse)
	if !ok {
		return nil, fmt.Errorf("could not cast mock argument: %+v", result)
	}
	return result, args.Error(1)
}
