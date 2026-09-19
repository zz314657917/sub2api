package service

import "testing"

func TestPelicanTimeoutValidation(t *testing.T) {
	for _, seconds := range []int{0, 30, 180, 1200, 3600, -1, 29, 3601} {
		input := PelicanPlanInput{GroupID: 1, ModelID: "gpt-test", IntervalMinutes: 60, MaxResults: 20, MinChars: 100, TimeoutSeconds: seconds}
		valid := seconds == 0 || (seconds >= 30 && seconds <= 3600)
		if (validPelicanInput(input) == nil) != valid {
			t.Fatalf("unexpected validation for %d", seconds)
		}
	}
}
