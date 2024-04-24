package platform

import (
	"context"
	"fmt"
	"net/http"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
)

func NewPlatformClient(host, token, version string) (*ClientWithResponses, error) {
	// we append base url in request editor, so set to an empty string here
	cl, err := NewClientWithResponses(
		"",
		WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			baseUrl := fmt.Sprintf("%s/platform/v1beta1", host)
			return clients.CoreRequestEditor(ctx, req, baseUrl, token, version)
		}),
	)
	return cl, err
}
