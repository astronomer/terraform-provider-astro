package import_script_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestImportScript(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Import Script Suite")
}
