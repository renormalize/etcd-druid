// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"testing"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/gomega"
)

func TestMergeMaps(t *testing.T) {
	testCases := []struct {
		name       string
		sourceMaps []map[string]string
		expected   map[string]string
	}{
		{
			name:       "nil source maps",
			sourceMaps: nil,
			expected:   nil,
		},
		{
			name: "one of the source maps is nil",
			sourceMaps: []map[string]string{
				{"key1": "value1"},
				nil,
			},
			expected: map[string]string{"key1": "value1"},
		},
		{
			name: "source maps are empty",
			sourceMaps: []map[string]string{
				{}, {},
			},
			expected: map[string]string{},
		},
		{
			name: "source maps are neither nil nor empty",
			sourceMaps: []map[string]string{
				{"key1": "value1"},
				{"key2": "value2"},
				{"key1": "value3"},
			},
			expected: map[string]string{
				"key1": "value3",
				"key2": "value2",
			},
		},
	}

	g := NewWithT(t)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g.Expect(MergeMaps(tc.sourceMaps...)).To(Equal(tc.expected))
		})
	}
}

func TestKey(t *testing.T) {
	testCases := []struct {
		name            string
		namespaceOrName string
		nameOpt         []string
		expected        client.ObjectKey
	}{
		{
			name:            "Namespace and name",
			namespaceOrName: "test-namespace",
			nameOpt:         []string{"test-name"},
			expected:        client.ObjectKey{Namespace: "test-namespace", Name: "test-name"},
		},
		{
			name:            "Only name",
			namespaceOrName: "test-name",
			nameOpt:         nil,
			expected:        client.ObjectKey{Name: "test-name"},
		},
	}

	g := NewWithT(t)
	t.Parallel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g.Expect(Key(tc.namespaceOrName, tc.nameOpt...)).To(Equal(tc.expected))
		})
	}
}

func TestIsEmptyString(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: true,
		},
		{
			name:     "Whitespace string",
			input:    "   ",
			expected: true,
		},
		{
			name:     "Non-empty string",
			input:    "non-empty",
			expected: false,
		},
	}

	g := NewWithT(t)
	t.Parallel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g.Expect(IsEmptyString(tc.input)).To(Equal(tc.expected))
		})
	}
}

func TestIfConditionOr(t *testing.T) {
	testCases := []struct {
		name      string
		condition bool
		trueVal   string
		falseVal  string
		expected  string
	}{
		{
			name:      "True condition",
			condition: true,
			trueVal:   "trueVal",
			falseVal:  "falseVal",
			expected:  "trueVal",
		},
		{
			name:      "False condition",
			condition: false,
			trueVal:   "trueVal",
			falseVal:  "falseVal",
			expected:  "falseVal",
		},
	}

	g := NewWithT(t)
	t.Parallel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g.Expect(IfConditionOr(tc.condition, tc.trueVal, tc.falseVal)).To(Equal(tc.expected))
		})
	}
}

func TestComputeScheduleInterval(t *testing.T) {
	testCases := []struct {
		name         string
		cronSchedule string
		expected     time.Duration
		expectError  bool
	}{
		{
			name:         "Valid cron schedule",
			cronSchedule: "0 0 * * *",
			expected:     24 * time.Hour,
			expectError:  false,
		},
		{
			name:         "Valid cron schedule",
			cronSchedule: "0 0 * * 1",
			expected:     24 * time.Hour * 7,
			expectError:  false,
		},
		{
			name:         "Valid cron schedule",
			cronSchedule: "*/1 * * * *",
			expected:     1 * time.Minute,
			expectError:  false,
		},
		{
			name:         "Valid cron schedule",
			cronSchedule: "0 */1 * * *",
			expected:     1 * time.Hour,
			expectError:  false,
		},
		{
			name:         "Invalid cron schedule",
			cronSchedule: "invalid-cron",
			expected:     0,
			expectError:  true,
		},
	}

	g := NewWithT(t)
	t.Parallel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			duration, err := ComputeScheduleInterval(tc.cronSchedule)
			if tc.expectError {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(duration).To(Equal(tc.expected))
			}
		})
	}
}
