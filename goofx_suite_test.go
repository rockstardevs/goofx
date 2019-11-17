package goofx_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestOfx(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "goofx Suite")
}
