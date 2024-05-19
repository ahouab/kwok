package cel

import (
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/stretchr/testify/assert"
)

func TestResourceList_Contains(t *testing.T) {
	tests := []struct {
		name    string
		r       ResourceList
		index   ref.Val
		want    ref.Val
		wantErr bool
	}{
		{
			name:  "ExistingResource",
			r:     NewResourceList(corev1.ResourceList{"cpu": resource.MustParse("100m")}),
			index: types.String("cpu"),
			want:  types.True,
		},
		{
			name:  "NonExistingResource",
			r:     NewResourceList(corev1.ResourceList{"cpu": resource.MustParse("100m")}),
			index: types.String("memory"),
			want:  types.False,
		},
		// Add more test cases as needed
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.r.Contains(tt.index)
			if tt.wantErr {
				assert.Error(t, fmt.Errorf(""), "Expected an error")
			} else {
				assert.Equal(t, tt.want, got, "Expected value doesn't match")
			}
		})
	}
}

func TestResourceList_Get(t *testing.T) {
	// Write test cases to cover the Get method
}

func TestResourceList_IsZeroValue(t *testing.T) {
	tests := []struct {
		name string
		r    ResourceList
		want bool
	}{
		{
			name: "NonEmptyList",
			r:    NewResourceList(corev1.ResourceList{"cpu": resource.MustParse("100m")}),
			want: false,
		},
		{
			name: "EmptyList",
			r:    ResourceList{},
			want: true,
		},
		// Add more test cases as needed
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.r.IsZeroValue()
			assert.Equal(t, tt.want, got, "Expected value doesn't match")
		})
	}
}

func TestResourceList_Size(t *testing.T) {
	tests := []struct {
		name string
		r    ResourceList
		want ref.Val
	}{
		{
			name: "NonEmptyList",
			r:    NewResourceList(corev1.ResourceList{"cpu": resource.MustParse("100m")}),
			want: types.Int(1),
		},
		{
			name: "EmptyList",
			r:    ResourceList{},
			want: types.Int(0),
		},
		// Add more test cases as needed
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.r.Size()
			assert.Equal(t, tt.want, got, "Expected value doesn't match")
		})
	}
}
