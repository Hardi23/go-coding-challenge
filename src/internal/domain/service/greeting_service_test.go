package service_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.prod.gitlab.aws.cloud.huk.de/huk/ds/core/golang-code-challenge/src/internal/domain/model"
	"gitlab.prod.gitlab.aws.cloud.huk.de/huk/ds/core/golang-code-challenge/src/internal/domain/service"
)

func TestGreetingService_Greet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		input           string
		expectedMessage string
		expectedErr     error
	}{
		{
			name:            "happy path",
			input:           "Welt",
			expectedMessage: "Hallo, Welt!",
		},
		{
			name:            "trims surrounding whitespace",
			input:           "   Welt   ",
			expectedMessage: "Hallo, Welt!",
		},
		{
			name:            "handles umlauts",
			input:           "Müller",
			expectedMessage: "Hallo, Müller!",
		},
		{
			name:            "accepts name at the 100 character limit",
			input:           strings.Repeat("a", 100),
			expectedMessage: "Hallo, " + strings.Repeat("a", 100) + "!",
		},
		{
			name:        "rejects empty name",
			input:       "",
			expectedErr: model.ErrInvalidRequest,
		},
		{
			name:        "rejects whitespace only name",
			input:       "   \t\n  ",
			expectedErr: model.ErrInvalidRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sut := service.NewGreetingService()

			result, err := sut.Greet(context.Background(), tc.input)

			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
				assert.Nil(t, result, "no result expected on error")
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tc.expectedMessage, result.Message)
		})
	}
}

func TestGreetingService_Greet_RespectsContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	sut := service.NewGreetingService()

	result, err := sut.Greet(ctx, "Welt")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Hallo, Welt!", result.Message)
}
