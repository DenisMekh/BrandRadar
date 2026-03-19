package healthcheck

import "testing"

func TestCheckersConstructorsAndNames(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		getName  func() string
		expected string
	}{
		{
			name: "postgres",
			getName: func() string {
				return NewPostgresChecker(nil).Name()
			},
			expected: "postgres",
		},
		{
			name: "redis",
			getName: func() string {
				return NewRedisChecker(nil).Name()
			},
			expected: "redis",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.getName(); got != tc.expected {
				t.Fatalf("unexpected checker name: got=%q want=%q", got, tc.expected)
			}
		})
	}
}
