package goofx_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/rockstardevs/goofx"
)

var _ = Describe("goofx", func() {
	Describe("Escape()", func() {
		Context("when given a string with unescaped chars", func() {
			It("should return a string with escaped chars", func() {
				input := "x < > \" ' & \r \t \n \x00"
				expected := "x &lt; &gt; &#34; &#39; &amp; &#xD; &#x9; &#xA; \ufffd"
				Expect(goofx.EscapeString(input)).To(Equal(expected))
			})
		})
	})
})
