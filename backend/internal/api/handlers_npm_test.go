package api

import (
	"testing"

	"github.com/vape-group/backend/internal/service"
)

func TestShouldBlockTenantDomainSaveForNPMResult(t *testing.T) {
	tests := []struct {
		name   string
		result *service.NPMResult
		want   bool
	}{
		{
			name: "nil result blocks save",
			want: true,
		},
		{
			name: "error status blocks save",
			result: &service.NPMResult{
				Status:     "error",
				Message:    "NPM sync failed",
				NPMUpdated: false,
			},
			want: true,
		},
		{
			name: "success without update blocks save",
			result: &service.NPMResult{
				Status:     "success",
				Message:    "no changes applied",
				NPMUpdated: false,
			},
			want: true,
		},
		{
			name: "warning does not block save",
			result: &service.NPMResult{
				Status:     "warning",
				Message:    "No NPM proxy host found, only database was updated",
				NPMUpdated: false,
			},
			want: false,
		},
		{
			name: "successful update allows save",
			result: &service.NPMResult{
				Status:         "success",
				Message:        "NPM proxy host updated successfully",
				NPMUpdated:     true,
				UpdatedDomains: []string{"example.com", "www.example.com"},
			},
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := shouldBlockTenantDomainSaveForNPMResult(test.result)
			if got != test.want {
				t.Fatalf("shouldBlockTenantDomainSaveForNPMResult() = %v, want %v", got, test.want)
			}
		})
	}
}
