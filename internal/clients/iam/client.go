package iam

import (
	"context"
	"fmt"
	"net/http"

	"github.com/astronomer/astronomer-terraform-provider/internal/clients"
)

func NewIamClient(host, token, version string) (*ClientWithResponses, error) {
	// we append base url in request editor, so set to an empty string here
	cl, err := NewClientWithResponses(
		"",
		WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			baseUrl := fmt.Sprintf("%s/iam/v1beta1", host)
			return clients.CoreRequestEditor(ctx, req, baseUrl, token, version)
		}),
	)
	return cl, err
}
