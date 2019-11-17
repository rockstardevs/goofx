package goofx_test

import (
	"bytes"
	"errors"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/shopspring/decimal"

	"github.com/rockstardevs/goofx"
)

const (
	layoutISO = "2006-01-02"
)

type FakeReader struct {
	err error
}

func (f FakeReader) Read(p []byte) (int, error) {
	return 0, f.err
}

type FakeCleaner struct {
	data string
}

func (f FakeCleaner) CleanupXML(data []byte) (*bytes.Buffer, error) {
	return bytes.NewBufferString(f.data), nil
}

var _ = Describe("goofx", func() {
	Describe("ParseDate()", func() {
		Context("when given a valid date string", func() {
			DescribeTable("should parse to a time.", func(input, expected string) {
				e, _ := time.Parse(layoutISO, expected)
				got, err := goofx.ParseDate(input)
				Expect(*got).To(BeTemporally("==", e))
				Expect(err).To(Succeed())
			},
				Entry("YYYYMMDD", "20191001", "2019-10-01"),
				Entry("YYYYMMDDHHMMSS", "20171108000000", "2017-11-08"),
				Entry("YYYYMMDDHHMMSS.f[z:Z]", "20170226120000.000[0:GMT]", "2017-02-26"),
			)
		})
		Context("when given a invalid date string", func() {
			DescribeTable("should return an error.", func(input string) {
				got, err := goofx.ParseDate(input)
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
		Context("when given a invalid timezone string", func() {
			It("should return an error", func() {
				got, err := goofx.ParseDate("20170226120000.000[0:TTT]")
				Expect(got).To(BeNil())
				Expect(err).To(MatchError("unknown time zone TTT"))
			})
		})
	})
	Describe("GetTxns", func() {
		Context("when given a document with no txns", func() {
			It("should return an empty txn set", func() {
				d := &goofx.Document{}
				t := make([]goofx.Transaction, 0)
				Expect(d.GetTxns()).To(Equal(&t))
			})
		})
		Context("when given a document a single txn set", func() {
			It("should return the single txn set", func() {
				t := []goofx.Transaction{goofx.Transaction{Type: "DEBIT", Amount: decimal.New(-15, 0)}}
				d := &goofx.Document{
					BRMS: []goofx.BankResponseMessageSet{
						goofx.BankResponseMessageSet{
							TRS: goofx.StatementTransactionResponseSet{
								RS: goofx.StatementResponseSet{Transactions: t},
							},
						},
					},
				}
				Expect(d.GetTxns()).To(Equal(&t))
			})
		})
		Context("when given a document with multiple txn sets", func() {
			It("should return an all txn sets", func() {
				t1 := []goofx.Transaction{goofx.Transaction{Type: "CREDIT", Amount: decimal.New(45, 0)}}
				t2 := []goofx.Transaction{goofx.Transaction{Type: "DEBIT", Amount: decimal.New(-30, 0)}}
				expected := make([]goofx.Transaction, 0, len(t1)+len(t2))
				for _, t := range t1 {
					expected = append(expected, t)
				}
				for _, t := range t2 {
					expected = append(expected, t)
				}

				d := &goofx.Document{
					BRMS: []goofx.BankResponseMessageSet{
						goofx.BankResponseMessageSet{
							TRS: goofx.StatementTransactionResponseSet{
								RS: goofx.StatementResponseSet{Transactions: t1},
							},
						},
						goofx.BankResponseMessageSet{
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
	Describe("NewDocumentFromXML()", func() {
		Context("when given invalid file", func() {
			It("should return an error", func() {
				r := FakeReader{err: errors.New("fake reader test error")}
				d, err := goofx.NewDocumentFromXML(&r, goofx.GetCleaner())
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
				d, err := goofx.NewDocumentFromXML(r, goofx.GetCleaner())
				Expect(err).To(MatchError("error - invalid file, OFX tag not found"))
				Expect(d).To(BeNil())
			})
		})
		Context("when given valid OFX data", func() {
			It("should return an initialized document", func() {
				r := strings.NewReader("<OFX></OFX>")
				d, err := goofx.NewDocumentFromXML(r, goofx.GetCleaner())
				Expect(err).To(BeNil())
				Expect(d).NotTo(BeNil())
			})
			It("should set txn count", func() {
				r := strings.NewReader("<OFX><STMTTRN><FITID>1</STMTTRN><STMTTRN>2</FITID></STMTTRN></OFX>")
				d, err := goofx.NewDocumentFromXML(r, goofx.GetCleaner())
				Expect(err).To(BeNil())
				Expect(d).NotTo(BeNil())
				Expect(d.TransactionCount).To(Equal(2))
			})
		})
	})
})
