package resources

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/clients/labs"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeAstro is an in-process stand-in for the Astro API that deliberately CANONICALIZES CIDRs on
// create (e.g. "10.1.2.3/8" -> "10.0.0.0/8") and returns that canonical form from the list endpoint.
// This is the exact server behaviour that triggered the GH #244/#314 "inconsistent result after
// apply" class: if Create/Update wrote the API's returned value back into the Required
// ip_address_ranges attribute, it would diverge from the planned config value.
func fakeAstro(t *testing.T) *httptest.Server {
	t.Helper()
	store := map[string]string{} // id -> canonical cidr
	seq := 0
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "allowed-ip-address-ranges") {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var body struct {
				AllowedIpAddressRanges []string `json:"allowedIpAddressRanges"`
			}
			b, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(b, &body)
			for _, c := range body.AllowedIpAddressRanges {
				canonical := c
				if _, n, err := net.ParseCIDR(c); err == nil {
					canonical = n.String()
				}
				seq++
				store[string(rune('a'+seq))] = canonical
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"allowedIpAddressRanges": []any{}})
		case http.MethodGet:
			list := make([]map[string]any, 0, len(store))
			for id, cidr := range store {
				list = append(list, map[string]any{"id": id, "ipAddressRange": cidr, "organizationId": "org"})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"allowedIpAddressRanges": list, "limit": 1000, "offset": 0, "totalCount": len(list),
			})
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
}

// TestUnit_Create_StoresPlannedRangesNotAPIList regresses the P1 inconsistent-result fix: Create must
// persist the planned CIDRs verbatim, never the (possibly canonicalized) values returned by the API.
func TestUnit_Create_StoresPlannedRangesNotAPIList(t *testing.T) {
	srv := fakeAstro(t)
	defer srv.Close()

	iamClient, err := iam.NewIamClient(srv.URL, "token", "test")
	require.NoError(t, err)
	labsClient, err := labs.NewLabsClient(srv.URL, "token", "test")
	require.NoError(t, err)

	r := &allowedIpAddressRangesResource{iamClient: iamClient, labsClient: labsClient, organizationId: "org"}

	ctx := context.Background()
	s := rschema.Schema{Attributes: schemas.AllowedIpAddressRangesResourceSchemaAttributes()}
	objType := s.Type().TerraformType(ctx).(tftypes.Object)
	setType := tftypes.Set{ElementType: tftypes.String}

	// Plan: the user asked for a non-canonical CIDR ("10.1.2.3/8"). The fake API stores it as
	// "10.0.0.0/8" and returns that from the list endpoint.
	planRaw := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id": tftypes.NewValue(tftypes.String, nil),
		"ip_address_ranges": tftypes.NewValue(setType, []tftypes.Value{
			tftypes.NewValue(tftypes.String, "10.1.2.3/8"),
		}),
	})

	req := resource.CreateRequest{Plan: tfsdk.Plan{Schema: s, Raw: planRaw}}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: s, Raw: tftypes.NewValue(objType, nil)}}

	r.Create(ctx, req, resp)
	require.False(t, resp.Diagnostics.HasError(), "Create returned diagnostics: %v", resp.Diagnostics)

	var out models.AllowedIpAddressRangesResource
	resp.Diagnostics.Append(resp.State.Get(ctx, &out)...)
	require.False(t, resp.Diagnostics.HasError())

	var stored []string
	out.IpAddressRanges.ElementsAs(ctx, &stored, false)

	// The fix: state holds the PLANNED value, not the API's canonicalized form. Under the old code
	// (data.IpAddressRanges = setVal from listAll) this would be ["10.0.0.0/8"] and Terraform core
	// would raise "Provider produced inconsistent result after apply".
	assert.Equal(t, []string{"10.1.2.3/8"}, stored)
	assert.NotContains(t, stored, "10.0.0.0/8", "state must not adopt the API-canonicalized CIDR")
}
