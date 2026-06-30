package labs

import (
	"encoding/json"
	"reflect"
	"testing"
)

// TestUpdateAlertRequestDiscriminatorSerialization is a regression test for the forked
// union template (internal/clients/oapi-templates/union.tmpl).
//
// The labs Update*Alert* variants declare their discriminator as an optional pointer
// (`Type *...Type `json:"type,omitempty"“), so a caller can build a variant WITHOUT setting
// Type. The fork injects the discriminator into the marshaled JSON instead of assigning a
// struct field, so the union still serializes the correct "type" and round-trips through
// ValueByDiscriminator(). With the upstream (unforked) template these unions either fail to
// compile or silently drop "type" when the caller leaves it nil.
//
// Each case below intentionally leaves Type nil to exercise exactly that path.
func TestUpdateAlertRequestDiscriminatorSerialization(t *testing.T) {
	tests := []struct {
		name       string
		build      func() (UpdateAlertRequest, error)
		wantType   string
		wantGoType reflect.Type
	}{
		{
			name: "DagDuration",
			build: func() (UpdateAlertRequest, error) {
				var u UpdateAlertRequest
				return u, u.FromUpdateDagDurationAlertRequest(UpdateDagDurationAlertRequest{})
			},
			wantType:   "DAG_DURATION",
			wantGoType: reflect.TypeOf(UpdateDagDurationAlertRequest{}),
		},
		{
			name: "DagFailure",
			build: func() (UpdateAlertRequest, error) {
				var u UpdateAlertRequest
				return u, u.FromUpdateDagFailureAlertRequest(UpdateDagFailureAlertRequest{})
			},
			wantType:   "DAG_FAILURE",
			wantGoType: reflect.TypeOf(UpdateDagFailureAlertRequest{}),
		},
		{
			name: "DagSuccess",
			build: func() (UpdateAlertRequest, error) {
				var u UpdateAlertRequest
				return u, u.FromUpdateDagSuccessAlertRequest(UpdateDagSuccessAlertRequest{})
			},
			wantType:   "DAG_SUCCESS",
			wantGoType: reflect.TypeOf(UpdateDagSuccessAlertRequest{}),
		},
		{
			name: "DagTimeliness",
			build: func() (UpdateAlertRequest, error) {
				var u UpdateAlertRequest
				return u, u.FromUpdateDagTimelinessAlertRequest(UpdateDagTimelinessAlertRequest{})
			},
			wantType:   "DAG_TIMELINESS",
			wantGoType: reflect.TypeOf(UpdateDagTimelinessAlertRequest{}),
		},
		{
			name: "TaskDuration",
			build: func() (UpdateAlertRequest, error) {
				var u UpdateAlertRequest
				return u, u.FromUpdateTaskDurationAlertRequest(UpdateTaskDurationAlertRequest{})
			},
			wantType:   "TASK_DURATION",
			wantGoType: reflect.TypeOf(UpdateTaskDurationAlertRequest{}),
		},
		{
			name: "TaskFailure",
			build: func() (UpdateAlertRequest, error) {
				var u UpdateAlertRequest
				return u, u.FromUpdateTaskFailureAlertRequest(UpdateTaskFailureAlertRequest{})
			},
			wantType:   "TASK_FAILURE",
			wantGoType: reflect.TypeOf(UpdateTaskFailureAlertRequest{}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := tt.build()
			if err != nil {
				t.Fatalf("From%s returned error: %v", tt.name, err)
			}

			// 1. The marshaled union carries the injected discriminator, even though the
			//    caller never set the (optional) Type field.
			b, err := json.Marshal(u)
			if err != nil {
				t.Fatalf("json.Marshal(union): %v", err)
			}
			var got struct {
				Type string `json:"type"`
			}
			if err := json.Unmarshal(b, &got); err != nil {
				t.Fatalf("json.Unmarshal: %v", err)
			}
			if got.Type != tt.wantType {
				t.Errorf("marshaled type = %q, want %q (json: %s)", got.Type, tt.wantType, b)
			}

			// 2. Discriminator() reads the wire value back.
			disc, err := u.Discriminator()
			if err != nil {
				t.Fatalf("Discriminator(): %v", err)
			}
			if disc != tt.wantType {
				t.Errorf("Discriminator() = %q, want %q", disc, tt.wantType)
			}

			// 3. ValueByDiscriminator() round-trips to the matching concrete variant.
			val, err := u.ValueByDiscriminator()
			if err != nil {
				t.Fatalf("ValueByDiscriminator(): %v", err)
			}
			if reflect.TypeOf(val) != tt.wantGoType {
				t.Errorf("ValueByDiscriminator() type = %T, want %s", val, tt.wantGoType)
			}
		})
	}
}
