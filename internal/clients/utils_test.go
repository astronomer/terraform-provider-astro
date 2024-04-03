package clients_test

import (
	"context"
	"net/http"
	"net/url"

	"github.com/astronomer/astronomer-terraform-provider/internal/clients"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client Utils Test", func() {
	var ctx context.Context
	BeforeEach(func() {
		ctx = context.Background()
	})

	Context("CoreRequestEditor", func() {
		var req http.Request
		BeforeEach(func() {
			req = http.Request{URL: &url.URL{Path: "/path"}, Header: make(http.Header)}
		})

		It("should add correct headers to request", func() {
			err := clients.CoreRequestEditor(ctx, &req, "http://localhost", "token", "v1")
			Expect(err).ToNot(HaveOccurred())
			Expect(req.URL.String()).To(Equal("http://localhost/path"))
			Expect(req.Header.Get("authorization")).To(Equal("Bearer token"))
			Expect(
				req.Header.Get("x-astro-client-identifier"),
			).To(Equal("astronomer-terraform-provider"))
			Expect(req.Header.Get("x-astro-client-version")).ToNot(BeEmpty())
			Expect(req.Header.Get("x-client-os-identifier")).ToNot(BeEmpty())
			Expect(req.Header.Get("User-Agent")).To(Equal("astronomer-terraform-provider/v1"))
		})
	})

	Context("NormalizeAPIError", func() {
		It("should return status code and no error if successful request", func() {
			resp := http.Response{StatusCode: http.StatusOK}
			status, diag := clients.NormalizeAPIError(ctx, &resp, nil)
			Expect(status).To(Equal(http.StatusOK))
			Expect(diag).To(BeNil())
		})

		It("should return 500 and add to daigs if http response is nil", func() {
			status, diag := clients.NormalizeAPIError(ctx, nil, nil)
			Expect(status).To(Equal(http.StatusInternalServerError))
			Expect(diag).ToNot(BeNil())
			Expect(diag.Detail()).To(ContainSubstring("failed to perform request"))
		})

		It("should return status code and error if response is not 200", func() {
			resp := http.Response{StatusCode: http.StatusNotFound}
			status, diag := clients.NormalizeAPIError(
				ctx,
				&resp,
				[]byte(`{"message": "error", "requestId": "123"}`),
			)
			Expect(status).To(Equal(http.StatusNotFound))
			Expect(diag).ToNot(BeNil())
			Expect(diag.Detail()).To(ContainSubstring("requestId: 123"))
			Expect(diag.Detail()).To(ContainSubstring("status: 404"))
		})
	})
})
