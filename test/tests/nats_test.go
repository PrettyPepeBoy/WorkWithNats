package tests

import (
	"TestTaskNats/internal/services/validation"
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestValid(t *testing.T) {
	t.Run("nats server sends to us valid data", func(t *testing.T) {
		type testCase struct {
			data     []byte
			expected bool
		}

		testCases := []testCase{{
			data:     []byte(`{"name":"TOYOTA", "price":500000, "amount":70}`),
			expected: true,
		}, {
			data:     []byte(`{"name":"TOYOTAOBJFOEFOEJIEJFIEHFUEHFUEHFUEHUEFHEUFUEHF", "price":500000, "amount":70}`),
			expected: false,
		}, {
			data:     []byte(``),
			expected: false,
		},
		}

		for _, rc := range testCases {
			check := validation.Valid(rc.data)
			assert.Equal(t, check, rc.expected)
		}
	})

}
