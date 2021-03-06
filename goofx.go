package goofx

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"io/ioutil"
	"regexp"
	"strconv"
	"time"

	"github.com/rockstardevs/decimal"

	"github.com/golang/glog"
)

//revive:disable:exported
//go:generate mockgen -package=mocks -source=cleaner.go -mock_names Cleaner=MockOFXCleaner -destination=mocks/cleaner.go

var txnPattern = regexp.MustCompile(`<STMTTRN>`)

// TransactionType is a transaction type as per the OFX Spec 2.2 Section 11.4.4.3
// https://www.ofx.net/downloads/OFX%202.2.pdf
type TransactionType string

const (
	// Common Transaction Types
	DEBIT  TransactionType = "DEBIT"
	CREDIT TransactionType = "CREDIT"
	// Uncommon Transaction Types
	INTEREST      TransactionType = "INT"
	DIVIDENT      TransactionType = "DIV"
	FEE           TransactionType = "FEE"
	SERVICECHARGE TransactionType = "SRVCHG"
	DEPOSIT       TransactionType = "DEP"
	ATM           TransactionType = "ATM"
	POS           TransactionType = "POS"
	TRANSFER      TransactionType = "XFER"
	CHECK         TransactionType = "CHECK"
	PAYMENT       TransactionType = "PAYMENT"
	CASH          TransactionType = "CASH"
	DIRECTDEPOSIT TransactionType = "DIRECTDEP"
	DIRECTDEBIT   TransactionType = "DIRECTDEBIT"
	REPEATPAYMENT TransactionType = "REPEATPMT"
	OTHER         TransactionType = "OTHER"
)

type Transaction struct {
	ID     string          `xml:"_"`
	Type   TransactionType `xml:"TRNTYPE"`
	Posted string          `xml:"DTPOSTED"`
	Amount decimal.Decimal `xml:"TRNAMT"`
	FitID  string          `xml:"FITID"`
	Date   string          `xml:"DTUSER,omitempty"`
	Name   string          `xml:"NAME,omitempty"`
	Payee  string          `xml:"PAYEE,omitempty"`
	Memo   string          `xml:"MEMO,omitempty"`
}

type SignOnResponse struct {
	Code           int    `xml:"STATUS>CODE"`
	Severity       string `xml:"STATUS>SEVERITY"`
	Date           string `xml:"DTSERVER"`
	Language       string `xml:"LANGUAGE"`
	Organization   string `xml:"FI>ORG"`
	OrganizationID string `xml:"FI>FID"`
	IntuitID       string `xml:"INTU.BID,omitempty"`
}

type StatementTransactionResponseSet struct {
	ID       string               `xml:"TRNUID"`
	Code     int                  `xml:"STATUS>CODE"`
	Severity string               `xml:"STATUS>SEVERITY"`
	RS       StatementResponseSet `xml:"STMTRS"`
}

type Balance struct {
	Amount decimal.Decimal `xml:"BALAMT"`
	Date   string          `xml:"DTASOF"`
}

type StatementResponseSet struct {
	Currency         string        `xml:"CURDEF"`
	BankID           string        `xml:"BANKACCTFROM>BANKID"`
	AccountID        string        `xml:"BANKACCTFROM>ACCTID"`
	AccountType      string        `xml:"BANKACCTFROM>ACCTTYPE"`
	StartDate        string        `xml:"BANKTRANLIST>DTSTART"`
	EndDate          string        `xml:"BANKTRANLIST>DTEND"`
	Transactions     []Transaction `xml:"BANKTRANLIST>STMTTRN"`
	LedgerBalance    Balance       `xml:"LEDGERBAL"`
	AvailableBalance Balance       `xml:"AVAILBAL"`
}

type BankResponseMessageSet struct {
	TRS StatementTransactionResponseSet `xml:"STMTTRNRS"`
}

// Document is a parsed OFX/QFX Statement.
// This does not implement the complete rfc spec yet.
type Document struct {
	XMLName          xml.Name                 `xml:"OFX"`
	Response         SignOnResponse           `xml:"SIGNONMSGSRSV1>SONRS"`
	BRMS             []BankResponseMessageSet `xml:"BANKMSGSRSV1"`
	TransactionCount int
}

// NewDocumentFromXML parses the given file into a Document.
func NewDocumentFromXML(reader io.Reader, cleaner Cleaner) (*Document, error) {
	cleanXML, err := cleanData(reader, cleaner)
	if err != nil {
		return nil, err
	}

	glog.V(3).Infof("cleanXML: %s", cleanXML.String())
	document := &Document{}
	if err = xml.Unmarshal(cleanXML.Bytes(), document); err != nil {
		return nil, err
	}

	matches := txnPattern.FindAllIndex(cleanXML.Bytes(), -1)
	if matches != nil {
		document.TransactionCount = len(matches)
	}

	return document, nil
}

func cleanData(reader io.Reader, cleaner Cleaner) (*bytes.Buffer, error) {
	var (
		data []byte // Buffer to parse raw bytes from the input file.
		err  error
	)
	// Parse raw byte from the source file into data.
	if data, err = ioutil.ReadAll(reader); err != nil {
		return nil, err
	}

	data = preprocessOFXData(data)
	err = cleaner.Init(data)
	if err != nil {
		return nil, err
	}

	return cleaner.CleanupXML()
}

// GetTxns returns all transactions from the OFX document.
// These may belong to different accounts but we're assuming that by being placed along with an
// account metadata file, all txns are meant to be imported into the same account specified by the
// account metadata.
func (d *Document) GetTxns() *[]Transaction {
	txns := make([]Transaction, 0)
	for _, b := range d.BRMS {
		txns = append(txns, b.TRS.RS.Transactions...)
	}
	return &txns
}

// ParseDate parses the given OFX formatted date string to a time.Time object.
//
// If loc is not nil, it is used as the timezone location if the date doesn't
// contain a parsable timezone.
func ParseDate(d string, loc *time.Location) (*time.Time, error) {
	var (
		re     = regexp.MustCompile(`(?P<date>\d{8})(?P<time>\d{4}\d{2}?(?:\.\d{3})?)?(?:\.\d{3})?(?:\[(?P<offset>-?\d+):(?P<tz>\S+)])?`)
		format = "20060102"
		parts  = re.FindStringSubmatch(d)
		err    error
	)
	if len(parts) == 0 {
		return nil, errors.New("error - date string can not be parsed")
	}
	location := loc
	if location == nil {
		location = time.FixedZone("UTC", 0)
	}
	if parts[3] != "" {
		offset, err := strconv.ParseFloat(parts[3], 64)
		if err != nil {
			return nil, errors.New("error - date tz offset can not be parsed")
		}
		location = time.FixedZone(parts[3], int(offset*60*60))
	}
	glog.V(3).Infof("parts:%q format:%s", parts, format)
	t, err := time.ParseInLocation(format, parts[1], location)
	return &t, err
}
