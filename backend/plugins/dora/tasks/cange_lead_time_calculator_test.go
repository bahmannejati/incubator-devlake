/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tasks

import (
	"testing"
	"time"
)

func TestCalculateTime(t *testing.T) {
	start := time.Date(2022, time.January, 1, 9, 0, 0, 0, time.UTC)
	end := time.Date(2022, time.January, 5, 15, 0, 0, 0, time.UTC)

	expectedDuration := 45 * time.Hour

	duration := calculateTime(start, end)

	if duration != expectedDuration {
		t.Errorf("Expected duration: %v, but got: %v", expectedDuration, duration)
	}
}

func TestCalculateTime_Weekend(t *testing.T) {
	start := time.Date(2022, time.January, 5, 9, 0, 0, 0, time.UTC)
	end := time.Date(2022, time.January, 8, 15, 0, 0, 0, time.UTC)

	expectedDuration := 15 * time.Hour

	duration := calculateTime(start, end)

	if duration != time.Duration(expectedDuration) {
		t.Errorf("Expected duration: %v, but got: %v", expectedDuration, duration)
	}
}

func TestCalculateTime_NonWorkingHour(t *testing.T) {
	start := time.Date(2022, time.January, 2, 8, 0, 0, 0, time.UTC)
	end := time.Date(2022, time.January, 4, 22, 0, 0, 0, time.UTC)

	expectedDuration := 30

	duration := calculateTime(start, end)

	if duration != time.Duration(expectedDuration)*time.Hour {
		t.Errorf("Expected duration: %v, but got: %v", expectedDuration, duration)
	}
}
