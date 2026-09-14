package application

import "testing"

func TestKnownMetricIncludesRegisteredEvidenceSeries(t *testing.T) {
	metrics := []string{
		"episode_reward",
		"moving_average_reward",
		"epsilon",
		"loss",
		"episode_steps",
		"collisions",
		"success_rate",
		"action_up",
		"action_right",
		"action_down",
		"action_left",
		"training_examples",
		"prediction_error",
		"accumulated_error",
		"rollout_risk",
	}
	for _, metric := range metrics {
		if !knownMetric(metric) {
			t.Fatalf("registered metric %q was rejected", metric)
		}
	}
	if knownMetric("arbitrary_metric") {
		t.Fatal("arbitrary metric was accepted")
	}
}
