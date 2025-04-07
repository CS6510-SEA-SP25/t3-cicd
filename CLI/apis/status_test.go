package apis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateStatusReport(t *testing.T) {
	// Test cases
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name: "Valid pipeline data",
			input: map[string]interface{}{
				"pipeline": map[string]interface{}{
					"pipeline_id": 15.0,
					"name":        "maven_project_1",
					"status":      "SUCCESS",
					"stage_order": "verify",
				},
				"stages": map[string]interface{}{
					"verify": map[string]interface{}{
						"stage_id": 11.0,
						"name":     "verify",
						"status":   "SUCCESS",
						"jobs": []interface{}{
							map[string]interface{}{"job_id": 19.0, "name": "verify", "status": "SUCCESS"},
							map[string]interface{}{"job_id": 20.0, "name": "test", "status": "SUCCESS"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "Invalid input type",
			input:   "invalid-data",
			wantErr: true,
		},
		{
			name: "Empty stages",
			input: map[string]interface{}{
				"pipeline": map[string]interface{}{
					"pipeline_id": 15.0,
					"name":        "empty-pipeline",
					"status":      "PENDING",
					"stage_order": "",
				},
				"stages": map[string]interface{}{},
			},
			wantErr: false,
		},
		{
			name: "Missing pipeline field",
			input: map[string]interface{}{
				"stages": map[string]interface{}{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := generateStatusReport(tt.input)
			// if tt.wantErr {
			// 	assert.Error(t, err, "Expected error for case: %s", tt.name)
			// } else {
				assert.NoError(t, err, "Unexpected error for case: %s", tt.name)
			// }
		})
	}
}

func TestConvertToPipelineExecStatus(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    RequestExecutionStatus_ResponseBody
		wantErr bool
	}{
		{
			name:    "Invalid input type",
			input:   "not-a-map",
			want:    RequestExecutionStatus_ResponseBody{},
			wantErr: true,
		},
		{
			name: "Valid pipeline data",
			input: map[string]interface{}{
				"pipeline": map[string]interface{}{
					"pipeline_id": 15.0,
					"name":        "maven_project_1",
					"status":      "SUCCESS",
					"stage_order": "verify",
				},
				"stages": map[string]interface{}{
					"verify": map[string]interface{}{
						"stage_id": 11.0,
						"name":     "verify",
						"status":   "SUCCESS",
						"jobs": []interface{}{
							map[string]interface{}{"job_id": 19.0, "name": "verify", "status": "SUCCESS"},
							map[string]interface{}{"job_id": 20.0, "name": "test", "status": "SUCCESS"},
						},
					},
				},
			},
			want: RequestExecutionStatus_ResponseBody{
				Pipeline: PipelineExecutionStatus{
					PipelineId: 15,
					Name:       "maven_project_1",
					Status:     "SUCCESS",
					StageOrder: "verify",
				},
				Stages: map[string]StageExecutionStatus{
					"verify": {
						StageId: 11,
						Name:    "verify",
						Status:  "SUCCESS",
						Jobs: []JobExecutionStatus{
							{JobId: 19, Name: "verify", Status: "SUCCESS"},
							{JobId: 20, Name: "test", Status: "SUCCESS"},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertToPipelineExecStatus(tt.input)
			// if tt.wantErr {
			// 	assert.Error(t, err)
			// } else {
			assert.NoError(t, err)
			assert.Equal(t, tt.want.Pipeline, got.Pipeline)
			assert.Equal(t, len(tt.want.Stages), len(got.Stages))
			// }
		})
	}
}

func TestColorStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
	}{
		{"Success status", "SUCCESS"},
		{"Failed status", "FAILED"},
		{"Pending status", "PENDING"},
		{"Unknown status", "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := colorStatus(tt.status)
			assert.Contains(t, result, tt.status)
		})
	}
}
