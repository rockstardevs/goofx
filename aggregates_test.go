package goofx_test

import (
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/rockstardevs/goofx"
)

var _ = Describe("goofx", func() {
	Describe("GetAggregates()", func() {
		It("should return the singleton instance.", func() {
			i1 := goofx.GetAggregates()
			i2 := goofx.GetAggregates()
			Expect(i1).NotTo(BeNil())
			Expect(i2).NotTo(BeNil())
			Expect(reflect.ValueOf(i1).Pointer()).To(Equal(reflect.ValueOf(i2).Pointer()))
		})
	})
	Describe("IsAggregate()", func() {
		Context("when given an element name", func() {
			DescribeTable("should return true if the element is aggregate", func(name string, expected bool) {
				Expect(goofx.IsAggregate(name)).To(Equal(expected))
			},
				Entry("OFX", "OFX", true),
				Entry("SIGNONMSGSRSV1", "SIGNONMSGSRSV1", true),
				Entry("SONRS", "SONRS", true),
				Entry("STATUS", "STATUS", true),
				Entry("FI", "FI", true),
				Entry("BANKMSGSRSV1", "BANKMSGSRSV1", true),
				Entry("STMTTRNRS", "STMTTRNRS", true),
				Entry("STMTRS", "STMTRS", true),
				Entry("BANKACCTFROM", "BANKACCTFROM", true),
				Entry("BANKTRANLIST", "BANKTRANLIST", true),
				Entry("STMTTRN", "STMTTRN", true),
				Entry("LEDGERBAL", "LEDGERBAL", true),
				Entry("AVAILBAL", "AVAILBAL", true),

				Entry("CODE", "CODE", false),
				Entry("SEVERITY", "SEVERITY", false),
				Entry("DEFAULT", "DEFAULT", false),
			)
		})
	})
})
