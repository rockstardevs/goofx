package goofx_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/rockstardevs/goofx"
)

var _ = Describe("goofx", func() {
	Describe("Init()", func() {
		Context("when given an unparsable OFX document", func() {
			It("should return an error", func() {
				c := goofx.NewCleaner()
				err := c.Init([]byte(`<STATUS><CODE>0</CODE></STATUS>`))
				Expect(err).To(MatchError("error - invalid file, OFX tag not found"))
			})
		})
		Context("when given an valid OFX document", func() {
			It("should initialize successfully", func() {
				c := goofx.NewCleaner()
				err := c.Init([]byte(`<OFX></OFX>`))
				Expect(err).To(BeNil())
			})
		})
	})
	Describe("CleanupXML()", func() {
		Context("when given an unparsable OFX document", func() {
			DescribeTable("should return an error", func(data []byte, errMessage string) {
				cleaner := goofx.NewCleaner()
				err := cleaner.Init(data)
				Expect(err).To(BeNil())
				cleanData, err := cleaner.CleanupXML()
				Expect(cleanData).To(BeNil())
				if errMessage != "" {
					Expect(err).To(MatchError(errMessage))
				} else {
					Expect(err).To(HaveOccurred())
				}
			},
				Entry("when containing malformed tokens",
					[]byte(`<OFX>>CODE<</OFX>`),
					""),
				Entry("when elements are missing start and end tag",
					[]byte(`<OFX><STMTTRN>foo</STMTTRN></STATUS>`),
					"error: charData(foo) missing start and end tags"),
				Entry("when elements have mismatched start and end tag",
					[]byte(`<OFX><CODE>bar</SEVERITY></STATUS>`),
					"error: charData(bar) has ambigious closing tags"),
				Entry("when elements have mismatched start and end tag",
					[]byte(`<OFX><STATUS>baz<SEVERITY>INFO</STATUS>`),
					"error: charData(baz) missing start and end tags"),
			)
		})
		Context("when given a parsable OFX document", func() {
			DescribeTable("should parse to clean XML", func(data []byte, expected []byte) {
				cleaner := goofx.NewCleaner()
				err := cleaner.Init(data)
				Expect(err).To(BeNil())
				cleanData, err := cleaner.CleanupXML()
				Expect(err).To(BeNil())
				Expect(cleanData).ToNot(BeNil())
				Expect(cleanData.Bytes()).To(Equal(expected))
			},
				Entry("when aggregate is well formed",
					[]byte(`<OFX><SIGNONMSGSRSV1>	</SIGNONMSGSRSV1></OFX>`),
					[]byte(`<OFX><SIGNONMSGSRSV1></SIGNONMSGSRSV1></OFX>`)),
				Entry("when aggregate is missing end tags",
					[]byte(`<OFX><SIGNONMSGSRSV1></OFX>`),
					[]byte(`<OFX><SIGNONMSGSRSV1></SIGNONMSGSRSV1></OFX>`)),
				Entry("when aggregate is missing start tags",
					[]byte(`<OFX></SIGNONMSGSRSV1></OFX>`),
					[]byte(`<OFX></OFX>`)),
				Entry("when element is missing end tags",
					[]byte(`<OFX>
							<STATUS>
							<CODE>0
							<SEVERITY>INFO
							</STATUS>
							<DTSERVER>20191027065402
							<LANGUAGE>ENG
							</OFX>`),
					[]byte(`<OFX><STATUS><CODE>0</CODE><SEVERITY>INFO</SEVERITY></STATUS><DTSERVER>20191027065402</DTSERVER><LANGUAGE>ENG</LANGUAGE></OFX>`)),
				Entry("when element is missing starting tags",
					[]byte(`<OFX>
							<BANKTRANLIST>
							2018-01-01</DTSTART>
							2018-06-30</DTEND>
							</BANKTRANLIST>
							</OFX>`),
					[]byte(`<OFX><BANKTRANLIST><DTSTART>2018-01-01</DTSTART><DTEND>2018-06-30</DTEND></BANKTRANLIST></OFX>`)),
				Entry("when aggregates have no nested elements",
					[]byte(`<OFX><BANKMSGSRSV1></STMTTRNRS></BANKMSGSRSV1></OFX>`),
					[]byte(`<OFX><BANKMSGSRSV1></BANKMSGSRSV1></OFX>`)),
			)
		})
	})
})
