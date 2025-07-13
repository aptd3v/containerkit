package health_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/aptd3v/go-contain/pkg/create/config/cc/health"
	"github.com/aptd3v/go-contain/pkg/create/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
)

func TestAssignments(t *testing.T) {
	tests := []struct {
		config   *container.HealthConfig
		setFn    health.SetHealthcheckConfig
		field    string
		wantErr  bool
		message  string
		expected any
	}{
		{
			config:   &container.HealthConfig{},
			setFn:    health.WithStartPeriod(10),
			field:    "StartPeriod",
			wantErr:  false,
			message:  "WithStartPeriod ok",
			expected: 10 * time.Second,
		},
		{
			config:   &container.HealthConfig{},
			setFn:    health.WithStartPeriod(0),
			field:    "StartPeriod",
			wantErr:  false,
			message:  "WithStartPeriod inherit",
			expected: 0 * time.Second,
		},
		{
			config:   &container.HealthConfig{},
			setFn:    health.WithStartPeriod(-1),
			field:    "StartPeriod",
			wantErr:  true,
			message:  "WithStartPeriod negative",
			expected: 0 * time.Second,
		},
		{
			config:   &container.HealthConfig{},
			setFn:    health.WithTimeout(10),
			field:    "Timeout",
			wantErr:  false,
			message:  "WithTimeout ok",
			expected: 10 * time.Second,
		},
		{
			config:   &container.HealthConfig{},
			setFn:    health.WithTimeout(-1),
			field:    "Timeout",
			wantErr:  true,
			message:  "WithTimeout negative",
			expected: 0 * time.Second,
		},
		{
			config:   &container.HealthConfig{},
			setFn:    health.WithInterval(10),
			field:    "Interval",
			wantErr:  false,
			message:  "WithInterval ok",
			expected: 10 * time.Second,
		},
		{
			config:   &container.HealthConfig{},
			setFn:    health.WithInterval(-1),
			field:    "Interval",
			wantErr:  true,
			message:  "WithInterval negative",
			expected: 0 * time.Second,
		},
		{
			config:   &container.HealthConfig{},
			setFn:    health.WithRetries(10),
			field:    "Retries",
			wantErr:  false,
			message:  "WithRetries ok",
			expected: 10,
		},
		{
			config:   &container.HealthConfig{},
			setFn:    health.WithRetries(-1),
			field:    "Retries",
			wantErr:  true,
			message:  "WithRetries negative",
			expected: 0,
		},
		{
			config:  &container.HealthConfig{},
			setFn:   health.WithTest("CMD-SHELL", "curl -f http://localhost:8080/health || exit 1"),
			field:   "Test",
			wantErr: false,
			message: "WithTest ok",
			expected: []string{
				"CMD-SHELL",
				"curl -f http://localhost:8080/health || exit 1",
			},
		},
		{
			config:   &container.HealthConfig{},
			setFn:    health.WithTest(),
			field:    "Test",
			wantErr:  false,
			message:  "WithTest empty",
			expected: []string{},
		},
		{
			config:   &container.HealthConfig{},
			setFn:    health.Fail(errors.New("test error")),
			field:    "",
			wantErr:  true,
			message:  "Fail error",
			expected: nil,
		},
		{
			config:  &container.HealthConfig{},
			setFn:   health.Failf("test error %s", "test"),
			field:   "",
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.message, func(t *testing.T) {
			err := test.setFn(test.config)
			if test.wantErr {
				assert.Error(t, err)
				assert.True(t, errdefs.IsContainerConfigError(err), "expected container config error")
			} else {
				assert.NoError(t, err)
			}
			if test.field != "" {
				assert.Equal(t,
					test.expected,
					reflect.ValueOf(*test.config).FieldByName(test.field).Interface(),
					test.message,
				)
			}
		})
	}
}
