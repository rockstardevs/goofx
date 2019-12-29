package goofx_test

import (
	"bytes"
	"errors"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/rockstardevs/decimal"

	"github.com/rockstardevs/goofx"
)

type FakeReader struct {
	err error
}

func (f FakeReader) Read(p []byte) (int, error) {
	return 0, f.err
}

type FakeCleaner struct {
	err  error
	data string
}

func (f FakeCleaner) Init(data []byte) error {
	return nil
}

func (f FakeCleaner) CleanupXML() (*bytes.Buffer, error) {
	if f.err != nil {
		return nil, f.err
	}
	return bytes.NewBufferString(f.data), nil
}

var _ = Describe("goofx", func() {
	Describe("ParseDate()", func() {
		Context("when given a valid date string", func() {
			DescribeTable("should parse to a time.", func(input, expected string, loc *time.Location) {
				e, _ := time.Parse(time.RFC822Z, expected)
				got, err := goofx.ParseDate(input, loc)
				Expect(*got).To(BeTemporally("==", e))
				Expect(err).To(Succeed())
			},
				Entry("YYYYMMDD", "20191001", "01 Oct 19 00:00 +0000", nil),
				Entry("YYYYMMDD", "20191001", "01 Oct 19 00:00 -1100", time.FixedZone("TTT", -11*60*60)),
				Entry("YYYYMMDDHHMMSS", "20171108090000", "08 Nov 17 00:00 +0000", nil),
				Entry("YYYYMMDDHHMMSS", "20171108090000", "08 Nov 17 00:00 +1000", time.FixedZone("TTT", 10*60*60)),
				Entry("YYYYMMDDHHMMSS.f[z:Z]", "20170226120000.000[0:GMT]", "26 Feb 17 00:00 +0000", nil),
				Entry("YYYYMMDDHHMMSS.f[z:Z]", "20180313093000.000[-10:EDT]", "13 Mar 18 00:00 -1000", nil),
			)
		})
		Context("when given a invalid date string", func() {
			DescribeTable("should return an error.", func(input string) {
				got, err := goofx.ParseDate(input, nil)
				Expect(got).To(BeNil())
				Expect(err).To(MatchError("error - date string can not be parsed"))
			},
				Entry("Empty", ""),
				Entry("Invalid text", "test"),
				Entry("Invalid format", "2019/01/02"),
				Entry("Missing month and date", "2019"),
				Entry("Missing date", "2019-01"),
			)
		})
	})
	Describe("NewDocumentFromXML()", func() {
		Context("when given invalid file", func() {
			It("should return an error", func() {
				r := FakeReader{err: errors.New("fake reader test error")}
				d, err := goofx.NewDocumentFromXML(&r, goofx.NewCleaner())
				Expect(err).To(MatchError("fake reader test error"))
				Expect(d).To(BeNil())
			})
		})
		Context("when given invalid XML", func() {
			It("should return an error", func() {
				r := strings.NewReader("")
				d, err := goofx.NewDocumentFromXML(r, &FakeCleaner{data: "><"})
				Expect(err).To(MatchError("XML syntax error on line 1: unexpected EOF"))
				Expect(d).To(BeNil())
			})
		})
		Context("when given invalid OFX data missing OFX tag", func() {
			It("should return an error", func() {
				r := strings.NewReader("<BANKMSGSRSV1></BANKMSGSRSV1>")
				d, err := goofx.NewDocumentFromXML(r, goofx.NewCleaner())
				Expect(err).To(MatchError("error - invalid file, OFX tag not found"))
				Expect(d).To(BeNil())
			})
		})
		Context("when given data that can not be cleaned", func() {
			It("should return an error", func() {
				r := strings.NewReader("")
				d, err := goofx.NewDocumentFromXML(r, &FakeCleaner{err: errors.New("test error - failed to clean data")})
				Expect(err).To(MatchError("test error - failed to clean data"))
				Expect(d).To(BeNil())
			})
		})
		Context("when given valid OFX data", func() {
			It("should return an initialized document", func() {
				r := strings.NewReader("<OFX></OFX>")
				d, err := goofx.NewDocumentFromXML(r, goofx.NewCleaner())
				Expect(err).To(BeNil())
				Expect(d).NotTo(BeNil())
			})
			It("should set txn count", func() {
				r := strings.NewReader("<OFX><STMTTRN><FITID>1</STMTTRN><STMTTRN>2</FITID></STMTTRN></OFX>")
				d, err := goofx.NewDocumentFromXML(r, goofx.NewCleaner())
				Expect(err).To(BeNil())
				Expect(d).NotTo(BeNil())
				Expect(d.TransactionCount).To(Equal(2))
			})
		})
	})
	Describe("Document", func() {
		Describe("GetTxns()", func() {
			Context("when document has no txns", func() {
				It("should return an empty txn set", func() {
					d := &goofx.Document{}
					t := make([]goofx.Transaction, 0)
					Expect(d.GetTxns()).To(Equal(&t))
				})
			})
			Context("when document has a single txn set", func() {
				It("should return the single txn set", func() {
					t := []goofx.Transaction{{Type: "DEBIT", Amount: decimal.New(-15, 0)}}
					d := &goofx.Document{
						BRMS: []goofx.BankResponseMessageSet{
							{
								TRS: goofx.StatementTransactionResponseSet{
									RS: goofx.StatementResponseSet{Transactions: t},
								},
							},
						},
					}
					Expect(d.GetTxns()).To(Equal(&t))
				})
			})
			Context("when document has multiple txn sets", func() {
				It("should return all txn sets", func() {
					t1 := []goofx.Transaction{{Type: "CREDIT", Amount: decimal.New(45, 0)}}
					t2 := []goofx.Transaction{{Type: "DEBIT", Amount: decimal.New(-30, 0)}}
					expected := make([]goofx.Transaction, 0, len(t1)+len(t2))
					expected = append(expected, t1...)
					expected = append(expected, t2...)

					d := &goofx.Document{
						BRMS: []goofx.BankResponseMessageSet{
							{
								TRS: goofx.StatementTransactionResponseSet{
									RS: goofx.StatementResponseSet{Transactions: t1},
								},
							},
							{
								TRS: goofx.StatementTransactionResponseSet{
									RS: goofx.StatementResponseSet{Transactions: t2},
								},
							},
						},
					}
					Expect(d.GetTxns()).To(Equal(&expected))
				})
			})
		})
	})
})
