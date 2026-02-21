// SPDX-License-Identifier: GPL-3.0-only
// Copyright (c) 2021 by Rafael Osipov <rafael.osipov@outlook.com>

package testutil

// This package is not referenced from production code, so it should not be included to the executable file.
// It is referenced only as a helper from tests.

import "time"

// Mock implementation of Clock interface for testing purposes.
type MockClock struct {
	FixedTime time.Time
}

// Now gets the actual time for the MockClock instance.
func (m MockClock) Now() time.Time { return m.FixedTime }
