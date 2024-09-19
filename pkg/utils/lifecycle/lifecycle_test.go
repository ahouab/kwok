/*
Copyright 2024 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package lifecycle

import (
	"context"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/labels"

	"sigs.k8s.io/kwok/pkg/apis/internalversion"
	"sigs.k8s.io/kwok/pkg/utils/format"
)

func TestListAllPossibleStages(t *testing.T) {
	stages := []*internalversion.Stage{
		{
			Spec: internalversion.StageSpec{
				Selector: &internalversion.StageSelector{
					MatchLabels: map[string]string{
						"a": "b",
						"c": "d",
					},
				},
			},
		},
		{
			Spec: internalversion.StageSpec{
				Selector: &internalversion.StageSelector{
					MatchAnnotations: map[string]string{
						"e": "f",
						"g": "h",
					},
				},
			},
		},
	}
	lc, err := NewLifecycle(stages)
	if err != nil {
		t.Fatal("Could not create a new lifecycle:", err)
	}
	label := labels.Set{
		"a": "b",
		"c": "d",
	}
	annotation := labels.Set{
		"e": "f",
		"g": "h",
	}
	var data interface{}
	var possibleStages []*Stage
	possibleStages, err = lc.ListAllPossible(label, annotation, data)
	if err != nil {
		t.Fatal("Could not list all possible Stages:", err)
	}
	if possibleStages == nil {
		t.Fatal("Did not expect the list of all possible stages to be empty")
	}
	for i := 0; i < len(possibleStages); i++ {
		stageLabels := possibleStages[i].matchLabels
		stageAnnotation := possibleStages[i].matchAnnotations
		if stageLabels != nil {
			if stageLabels.String() != label.String() {
				t.Fatal("StageLabels and label do not match")
			}
		}
		if stageAnnotation != nil {
			if stageAnnotation.String() != annotation.String() {
				t.Fatal("StageAnnotations and annotation do not match")
			}
		}
	}
}

func TestLifecycleMatch(t *testing.T) {
	stages := []*internalversion.Stage{
		{
			Spec: internalversion.StageSpec{
				Selector: &internalversion.StageSelector{
					MatchLabels: map[string]string{
						"a": "b",
						"c": "d",
					},
				},
			},
		},
		{
			Spec: internalversion.StageSpec{
				Selector: &internalversion.StageSelector{
					MatchAnnotations: map[string]string{
						"e": "f",
						"g": "h",
					},
				},
			},
		},
	}
	lc, err := NewLifecycle(stages)
	if err != nil {
		t.Fatal("Could not create a new lifecycle:", err)
	}
	label := labels.Set{
		"a": "b",
		"c": "d",
	}
	annotation := labels.Set{
		"e": "f",
		"g": "h",
	}
	var data interface{}
	_, err = lc.Match(label, annotation, data)
	if err != nil {
		t.Fatal("Could not match Stage:", err)
	}
}

func TestStageMatch(t *testing.T) {
	for _, tc := range []struct {
		Scenario   string
		Stage      *internalversion.Stage
		Label      labels.Set
		Annotation labels.Set
		Expected   bool
	}{
		{
			Scenario: "Test MatchLabels",
			Stage: &internalversion.Stage{
				Spec: internalversion.StageSpec{
					Selector: &internalversion.StageSelector{
						MatchLabels: map[string]string{
							"a": "b",
							"c": "d",
						},
					},
				},
			},
			Label: labels.Set{
				"a": "b",
				"c": "d",
			},
			Expected: true,
		},
		{
			Scenario: "Test MatchLabels with incorrect labels",
			Stage: &internalversion.Stage{
				Spec: internalversion.StageSpec{
					Selector: &internalversion.StageSelector{
						MatchLabels: map[string]string{
							"a": "b",
							"c": "d",
						},
					},
				},
			},
			Label: labels.Set{
				"a": "b",
			},
			Expected: false,
		},
		{
			Scenario: "Test MatchAnnotations",
			Stage: &internalversion.Stage{
				Spec: internalversion.StageSpec{
					Selector: &internalversion.StageSelector{
						MatchAnnotations: map[string]string{
							"a": "b",
							"c": "d",
						},
					},
				},
			},
			Annotation: labels.Set{
				"a": "b",
				"c": "d",
			},
			Expected: true,
		},
		{
			Scenario: "Test MatchAnnotations with incorrect annotations",
			Stage: &internalversion.Stage{
				Spec: internalversion.StageSpec{
					Selector: &internalversion.StageSelector{
						MatchAnnotations: map[string]string{
							"a": "b",
							"c": "d",
						},
					},
				},
			},
			Annotation: labels.Set{
				"a": "b",
				"c": "e",
			},
			Expected: false,
		},
		{
			Scenario: "Test MatchExpressions that dont match",
			Stage: &internalversion.Stage{
				Spec: internalversion.StageSpec{
					Selector: &internalversion.StageSelector{
						MatchExpressions: []internalversion.SelectorRequirement{
							{
								Key:      ".test",
								Operator: internalversion.SelectorOpIn,
								Values:   []string{"b", "c"},
							},
							{
								Key:      ".test",
								Operator: internalversion.SelectorOpNotIn,
								Values:   []string{"b", "f"},
							},
						},
					},
				},
			},
			Expected: false,
		},
		{
			Scenario: "Test MatchExpressions that match",
			Stage: &internalversion.Stage{
				Spec: internalversion.StageSpec{
					Selector: &internalversion.StageSelector{
						MatchExpressions: []internalversion.SelectorRequirement{
							{
								Key:      ".test",
								Operator: internalversion.SelectorOpNotIn,
								Values:   []string{"b", "c"},
							},
							{
								Key:      ".test",
								Operator: internalversion.SelectorOpNotIn,
								Values:   []string{"b", "f"},
							},
						},
					},
				},
			},
			Expected: true,
		},
	} {
		t.Run(tc.Scenario, func(t *testing.T) {
			var actual bool
			var matchData interface{}
			stage, err := NewStage(tc.Stage)
			if err != nil {
				t.Fatalf("Could not create new stage: %v", err)
			}
			actual, err = stage.match(tc.Label, tc.Annotation, matchData)
			if err != nil {
				t.Fatalf("Could not create new stage: %v", err)
			}
			if tc.Expected && !actual {
				t.Error("Expected a match")
			}
			if !tc.Expected && actual {
				t.Error("Did not expect a match")
			}
		})
	}
}

func TestStageDelay(t *testing.T) {
	for _, tc := range []struct {
		Scenario string
		Stage    *internalversion.Stage
		Expected bool
	}{
		{
			Scenario: "Duration is not nil and jitterDuration is nil",
			Stage: &internalversion.Stage{
				Spec: internalversion.StageSpec{
					Selector: &internalversion.StageSelector{},
					Delay: &internalversion.StageDelay{
						DurationMilliseconds: format.Ptr(int64(9)),
					},
				},
			},
			Expected: true,
		},
		{
			Scenario: "Duration is not nil and jitterDuration is not nil",
			Stage: &internalversion.Stage{
				Spec: internalversion.StageSpec{
					Selector: &internalversion.StageSelector{},
					Delay: &internalversion.StageDelay{
						DurationMilliseconds:       format.Ptr(int64(9)),
						JitterDurationMilliseconds: format.Ptr(int64(4)),
					},
				},
			},
			Expected: true,
		},
		{
			Scenario: "Duration is nil and jitterDuration is not nil",
			Stage: &internalversion.Stage{
				Spec: internalversion.StageSpec{
					Selector: &internalversion.StageSelector{},
					Delay: &internalversion.StageDelay{
						JitterDurationMilliseconds: format.Ptr(int64(4)),
					},
				},
			},
			Expected: true,
		},
		{
			Scenario: "Duration is nil and jitterDuration is also nil",
			Stage: &internalversion.Stage{
				Spec: internalversion.StageSpec{
					Selector: &internalversion.StageSelector{},
				},
			},
			Expected: false,
		},
	} {
		t.Run(tc.Scenario, func(t *testing.T) {
			var actual bool
			var v interface{}
			stage, err := NewStage(tc.Stage)
			if err != nil {
				t.Fatalf("Could not create new stage: %v", err)
			}
			_, actual = stage.Delay(context.Background(), v, time.Now())
			if tc.Expected && !actual {
				t.Error("Expected a valid duration")
			}
			if !tc.Expected && actual {
				t.Error("Did not expect a valid duration")
			}
		})
	}
}
