package platform_v1

import (
	"encoding/json"
	"reflect"
	"testing"
)

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
				err := u.FromCreateStandardDeploymentRequest(CreateStandardDeploymentRequest{})
				return u, err
			},
			wantType:   "STANDARD",
			wantGoType: reflect.TypeOf(CreateStandardDeploymentRequest{}),
		},
		{
			name: "Dedicated",
			build: func() (CreateDeploymentRequest, error) {
				var u CreateDeploymentRequest
				err := u.FromCreateDedicatedDeploymentRequest(CreateDedicatedDeploymentRequest{})
				return u, err
			},
			wantType:   "DEDICATED",
			wantGoType: reflect.TypeOf(CreateDedicatedDeploymentRequest{}),
		},
		{
			name: "Hybrid",
			build: func() (CreateDeploymentRequest, error) {
				var u CreateDeploymentRequest
				err := u.FromCreateHybridDeploymentRequest(CreateHybridDeploymentRequest{})
				return u, err
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
			assertDiscriminator(t, u, u.Discriminator, u.ValueByDiscriminator, "type", tt.wantType, tt.wantGoType)
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
				err := u.FromUpdateStandardDeploymentRequest(UpdateStandardDeploymentRequest{})
				return u, err
			},
			wantType:   "STANDARD",
			wantGoType: reflect.TypeOf(UpdateStandardDeploymentRequest{}),
		},
		{
			name: "Dedicated",
			build: func() (UpdateDeploymentRequest, error) {
				var u UpdateDeploymentRequest
				err := u.FromUpdateDedicatedDeploymentRequest(UpdateDedicatedDeploymentRequest{})
				return u, err
			},
			wantType:   "DEDICATED",
			wantGoType: reflect.TypeOf(UpdateDedicatedDeploymentRequest{}),
		},
		{
			name: "Hybrid",
			build: func() (UpdateDeploymentRequest, error) {
				var u UpdateDeploymentRequest
				err := u.FromUpdateHybridDeploymentRequest(UpdateHybridDeploymentRequest{})
				return u, err
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
			assertDiscriminator(t, u, u.Discriminator, u.ValueByDiscriminator, "type", tt.wantType, tt.wantGoType)
		})
	}
}

func TestClusterRequestDiscriminatorSerialization(t *testing.T) {
	createTests := []struct {
		name       string
		build      func() (CreateClusterRequest, error)
		wantType   string
		wantGoType reflect.Type
	}{
		{
			name: "Aws",
			build: func() (CreateClusterRequest, error) {
				var u CreateClusterRequest
				err := u.FromCreateAwsClusterRequest(CreateAwsClusterRequest{})
				return u, err
			},
			wantType:   "AWS",
			wantGoType: reflect.TypeOf(CreateAwsClusterRequest{}),
		},
		{
			name: "Azure",
			build: func() (CreateClusterRequest, error) {
				var u CreateClusterRequest
				err := u.FromCreateAzureClusterRequest(CreateAzureClusterRequest{})
				return u, err
			},
			wantType:   "AZURE",
			wantGoType: reflect.TypeOf(CreateAzureClusterRequest{}),
		},
		{
			name: "Gcp",
			build: func() (CreateClusterRequest, error) {
				var u CreateClusterRequest
				err := u.FromCreateGcpClusterRequest(CreateGcpClusterRequest{})
				return u, err
			},
			wantType:   "GCP",
			wantGoType: reflect.TypeOf(CreateGcpClusterRequest{}),
		},
	}

	for _, tt := range createTests {
		t.Run("Create"+tt.name, func(t *testing.T) {
			u, err := tt.build()
			if err != nil {
				t.Fatalf("FromCreate%sClusterRequest returned error: %v", tt.name, err)
			}
			assertDiscriminator(t, u, u.Discriminator, u.ValueByDiscriminator, "cloudProvider", tt.wantType, tt.wantGoType)
		})
	}

	updateTests := []struct {
		name       string
		build      func() (UpdateClusterRequest, error)
		wantType   string
		wantGoType reflect.Type
	}{
		{
			name: "Dedicated",
			build: func() (UpdateClusterRequest, error) {
				var u UpdateClusterRequest
				err := u.FromUpdateDedicatedClusterRequest(UpdateDedicatedClusterRequest{})
				return u, err
			},
			wantType:   "DEDICATED",
			wantGoType: reflect.TypeOf(UpdateDedicatedClusterRequest{}),
		},
		{
			name: "Hybrid",
			build: func() (UpdateClusterRequest, error) {
				var u UpdateClusterRequest
				err := u.FromUpdateHybridClusterRequest(UpdateHybridClusterRequest{})
				return u, err
			},
			wantType:   "HYBRID",
			wantGoType: reflect.TypeOf(UpdateHybridClusterRequest{}),
		},
	}

	for _, tt := range updateTests {
		t.Run("Update"+tt.name, func(t *testing.T) {
			u, err := tt.build()
			if err != nil {
				t.Fatalf("FromUpdate%sClusterRequest returned error: %v", tt.name, err)
			}
			assertDiscriminator(t, u, u.Discriminator, u.ValueByDiscriminator, "clusterType", tt.wantType, tt.wantGoType)
		})
	}
}

func assertDiscriminator(
	t *testing.T,
	union any,
	discriminator func() (string, error),
	valueByDiscriminator func() (interface{}, error),
	jsonProperty string,
	wantType string,
	wantGoType reflect.Type,
) {
	t.Helper()

	b, err := json.Marshal(union)
	if err != nil {
		t.Fatalf("json.Marshal(union): %v", err)
	}
	var got map[string]json.RawMessage
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	var gotType string
	if raw, ok := got[jsonProperty]; ok {
		if err := json.Unmarshal(raw, &gotType); err != nil {
			t.Fatalf("json.Unmarshal(%s): %v", jsonProperty, err)
		}
	}
	if gotType != wantType {
		t.Errorf("marshaled %s = %q, want %q (json: %s)", jsonProperty, gotType, wantType, b)
	}

	disc, err := discriminator()
	if err != nil {
		t.Fatalf("Discriminator(): %v", err)
	}
	if disc != wantType {
		t.Errorf("Discriminator() = %q, want %q", disc, wantType)
	}

	val, err := valueByDiscriminator()
	if err != nil {
		t.Fatalf("ValueByDiscriminator(): %v", err)
	}
	if reflect.TypeOf(val) != wantGoType {
		t.Errorf("ValueByDiscriminator() type = %T, want %s", val, wantGoType)
	}
}
