package model_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.prod.gitlab.aws.cloud.huk.de/huk/ds/core/golang-code-challenge/src/internal/domain/model"
)

func TestGreetingRequest_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		request     model.GreetingRequest
		expectedErr error
	}{
		{
			name:    "valid name",
			request: model.GreetingRequest{Name: "Welt"},
		},
		{
			name:    "exactly 100 characters is valid",
			request: model.GreetingRequest{Name: strings.Repeat("a", 100)},
		},
		{
			name:    "exactly 100 multibyte characters is valid",
			request: model.GreetingRequest{Name: strings.Repeat("ü", 100)},
		},
		{
			name:        "101 characters is rejected",
			request:     model.GreetingRequest{Name: strings.Repeat("a", 101)},
			expectedErr: model.ErrInvalidRequest,
		},
		{
			name:        "empty name is rejected",
			request:     model.GreetingRequest{Name: ""},
			expectedErr: model.ErrInvalidRequest,
		},
		{
			name:        "whitespace only name is rejected",
			request:     model.GreetingRequest{Name: "   "},
			expectedErr: model.ErrInvalidRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.request.Validate()

			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
