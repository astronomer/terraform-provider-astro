package platform_v1

import (
	"encoding/json"
	"reflect"
	"testing"
)

// TestDeploymentRequestDiscriminatorSerialization is a regression test for the forked union
// template (internal/clients/oapi-templates/union.tmpl) as it applies to the Deployment
// unions, which are the first unions generated into this package.
//
// The v1 Create*DeploymentRequest variants declare their discriminator as an optional
// pointer (`Type *...Type `json:"type,omitempty"“) because v1 relaxed the create requests'
// required fields, so a caller can build a variant WITHOUT setting Type. The fork injects
// the discriminator into the marshaled JSON instead of assigning a struct field, so the
// union still serializes the correct "type" and round-trips through ValueByDiscriminator().
//
// Each case below intentionally leaves Type nil (create) or zero (update) to exercise
// exactly that path.
func TestDeploymentRequestDiscriminatorSerialization(t *testing.T) {
	createTests := []struct {
		name       string
		build      func() (CreateDeploymentRequest, error)
		wantType   string
		wantGoType reflect.Type
	}{
		{
			name: "Standard",
			build: func() (CreateDeploymentRequest, error) {
				var u CreateDeploymentRequest
				return u, u.FromCreateStandardDeploymentRequest(CreateStandardDeploymentRequest{})
			},
			wantType:   "STANDARD",
			wantGoType: reflect.TypeOf(CreateStandardDeploymentRequest{}),
		},
		{
			name: "Dedicated",
			build: func() (CreateDeploymentRequest, error) {
				var u CreateDeploymentRequest
				return u, u.FromCreateDedicatedDeploymentRequest(CreateDedicatedDeploymentRequest{})
			},
			wantType:   "DEDICATED",
			wantGoType: reflect.TypeOf(CreateDedicatedDeploymentRequest{}),
		},
		{
			name: "Hybrid",
			build: func() (CreateDeploymentRequest, error) {
				var u CreateDeploymentRequest
				return u, u.FromCreateHybridDeploymentRequest(CreateHybridDeploymentRequest{})
			},
			wantType:   "HYBRID",
			wantGoType: reflect.TypeOf(CreateHybridDeploymentRequest{}),
		},
	}

	for _, tt := range createTests {
		t.Run("Create"+tt.name, func(t *testing.T) {
			u, err := tt.build()
			if err != nil {
				t.Fatalf("FromCreate%sDeploymentRequest returned error: %v", tt.name, err)
			}
			assertDiscriminator(t, u, u.Discriminator, u.ValueByDiscriminator, tt.wantType, tt.wantGoType)
		})
	}

	updateTests := []struct {
		name       string
		build      func() (UpdateDeploymentRequest, error)
		wantType   string
		wantGoType reflect.Type
	}{
		{
			name: "Standard",
			build: func() (UpdateDeploymentRequest, error) {
				var u UpdateDeploymentRequest
				return u, u.FromUpdateStandardDeploymentRequest(UpdateStandardDeploymentRequest{})
			},
			wantType:   "STANDARD",
			wantGoType: reflect.TypeOf(UpdateStandardDeploymentRequest{}),
		},
		{
			name: "Dedicated",
			build: func() (UpdateDeploymentRequest, error) {
				var u UpdateDeploymentRequest
				return u, u.FromUpdateDedicatedDeploymentRequest(UpdateDedicatedDeploymentRequest{})
			},
			wantType:   "DEDICATED",
			wantGoType: reflect.TypeOf(UpdateDedicatedDeploymentRequest{}),
		},
		{
			name: "Hybrid",
			build: func() (UpdateDeploymentRequest, error) {
				var u UpdateDeploymentRequest
				return u, u.FromUpdateHybridDeploymentRequest(UpdateHybridDeploymentRequest{})
			},
			wantType:   "HYBRID",
			wantGoType: reflect.TypeOf(UpdateHybridDeploymentRequest{}),
		},
	}

	for _, tt := range updateTests {
		t.Run("Update"+tt.name, func(t *testing.T) {
			u, err := tt.build()
			if err != nil {
				t.Fatalf("FromUpdate%sDeploymentRequest returned error: %v", tt.name, err)
			}
			assertDiscriminator(t, u, u.Discriminator, u.ValueByDiscriminator, tt.wantType, tt.wantGoType)
		})
	}
}

func assertDiscriminator(
	t *testing.T,
	union any,
	discriminator func() (string, error),
	valueByDiscriminator func() (interface{}, error),
	wantType string,
	wantGoType reflect.Type,
) {
	t.Helper()

	// 1. The marshaled union carries the injected discriminator, even though the caller
	//    never set the (optional) Type field.
	b, err := json.Marshal(union)
	if err != nil {
		t.Fatalf("json.Marshal(union): %v", err)
	}
	var got struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if got.Type != wantType {
		t.Errorf("marshaled type = %q, want %q (json: %s)", got.Type, wantType, b)
	}

	// 2. Discriminator() reads the wire value back.
	disc, err := discriminator()
	if err != nil {
		t.Fatalf("Discriminator(): %v", err)
	}
	if disc != wantType {
		t.Errorf("Discriminator() = %q, want %q", disc, wantType)
	}

	// 3. ValueByDiscriminator() round-trips to the matching concrete variant.
	val, err := valueByDiscriminator()
	if err != nil {
		t.Fatalf("ValueByDiscriminator(): %v", err)
	}
	if reflect.TypeOf(val) != wantGoType {
		t.Errorf("ValueByDiscriminator() type = %T, want %s", val, wantGoType)
	}
}
