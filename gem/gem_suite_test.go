package gem_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGlobalEquitiesMomentum(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Global Equities Momentum Suite")
}
