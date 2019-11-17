# goofx

[![Build Status](https://img.shields.io/travis/rockstardevs/goofx.svg)](https://travis-ci.org/rockstardevs/goofx)
[![License](https://img.shields.io/github/license/rockstardevs/goofx)](https://github.com/rockstardevs/goofx/blob/master/LICENSE)
[![Coverage Status](https://img.shields.io/coveralls/rockstardevs/goofx.svg)](https://coveralls.io/r/rockstardevs/goofx?branch=master)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/rockstardevs/goofx)
[![Last Commit](https://img.shields.io/github/last-commit/rockstardevs/goofx)](https://github.com/rockstardevs/goofx/commits/master)

goofx (go ofx) is a Go library for parsing OFX format data files. It parses OFX data files and handles most common deviations from the spec like missing and unmatched tags.

The library uses a preprocessor that tries to infer missing tags and generates clean(er) XML before XML Unmarshalling into a Document object.

## Installation

Use the standard `go get` install:

```shell
$ go get github.com/rockstardevs/goofx
```

## Quick start

```golang
package main

import (
    "fmt"
    "os"
    "github.com/rockstardevs/goofx"
)

func main() {
    data := `
        <OFX>
        <SIGNONMSGSRSV1><SONRS>
            <STATUS><CODE>0<SEVERITY>INFO</STATUS>
            <DTSERVER>20190923042445<LANGUAGE>ENG
            <FI><ORG>Test Bank</ORG><FID>123</FID></FI>
        </SONRS></SIGNONMSGSRSV1>
		<BANKMSGSRSV1><STMTTRNRS>
			<TRNUID>0
			<STATUS><CODE>0<SEVERITY>INFO</STATUS>
			<STMTRS>
				<CURDEF>USD
				<BANKACCTFROM><BANKID>456<ACCTID>789<ACCTTYPE>CREDITLINE</BANKACCTFROM>
				<BANKTRANLIST>
					<DTSTART>20190101120000.000[0:GMT]<DTEND>20190131120000.000[0:GMT]
					<STMTTRN><TRNTYPE>DEBIT<DTPOSTED>20190119090000<TRNAMT>-20.96<FITID>20190119090001<NAME>Sample Expense</STMTTRN>
					<STMTTRN><TRNTYPE>DEBIT<DTPOSTED>20191115090000<TRNAMT>-115.26<FITID>20190122090002<NAME>Another Expense</STMTTRN>
				</BANKTRANLIST>
				<LEDGERBAL>
					<BALAMT>315.50<DTASOF>20190131120000.000[0:GMT]
				</LEDGERBAL>
				<AVAILBAL>
					<BALAMT>315.50<DTASOF>20190131120000.000[-7:GMT]
				</AVAILBAL>
			</STMTRS>
		</STMTTRNRS></BANKMSGSRSV1>
		</OFX>
    `
    reader := bytes.NewReader(data)

    // OR read from a file instead
    // f, err := os.Open("data.ofx")
    // defer f.Close()
    // reader = bufio.NewReader(f)

    document, err := ofx.NewDocumentFromXML(reader, ofx.GetCleaner())
    if err != nil {
        log.Exitf("error parsing data file - %s", err)
    }
    fmt.Printf("%#v", document)
}

// Output
//
// &goofx.Document{
//    XMLName:xml.Name{Space:"", Local:"OFX"},
//      Response:goofx.SignOnResponse{
//        Code:0,
//        Severity:"INFO",
//        Date:"20190131200000",
//        Language:"ENG",
//        Organization:"Test Bank",
//        OrganizationID:"123",
//        IntuitID:""},
//      BRMS:[]goofx.BankResponseMessageSet{
//        goofx.BankResponseMessageSet{
//          TRS:goofx.StatementTransactionResponseSet{
//            ID:"0",
//            Code:0,
//            Severity:"INFO",
//            RS:goofx.StatementResponseSet{
//              Currency:"USD",
//              BankID:"456",
//              AccountID:"789",
//              AccountType:"CREDITLINE",
//              StartDate:"20190101120000.000[0:GMT]",
//              EndDate:"20190131120000.000[0:GMT]",
//              Transactions:[]goofx.Transaction{
//                goofx.Transaction{
//                  Type:"DEBIT",
//                  Posted:"20190119090000",
//                  Amount:decimal.Decimal{value:(*big.Int)(0xc000143060), exp:-2},
//                  ID:"20190119090001",
//                  Date:"",
//                  Name:"Sample Expense",
//                  Payee:"",
//                  Memo:""},
//                goofx.Transaction{
//                  Type:"DEBIT",
//                  Posted:"20191115090000",
//                  Amount:decimal.Decimal{value:(*big.Int)(0xc000143320), exp:-2},
//                  ID:"20190122090002",
//                  Date:"",
//                  Name:"Another Expense",
//                  Payee:"",
//                  Memo:""}},
//              LedgerBalance:goofx.Balance{
//                Amount:decimal.Decimal{value:(*big.Int)(0xc000143560), exp:-1},
//                Date:"20190131120000.000[0:GMT]"},
//              AvailableBalance:goofx.Balance{
//                Amount:decimal.Decimal{value:(*big.Int)(0xc0001436e0), exp:-1},
//                Date:"20190131120000.000[-7:GMT]"}}}}},
//   TransactionCount:2}
```

## How it works

The OFX [spec](https://www.ofx.net/downloads/OFX%202.2.pdf) specifies that there are two distinct types of tags used in the message format.

### Aggregates

Aggregates are used for wrapping, nesting and hierarchical structure, they can not contain any char data. Aggregates can contain other aggegates of elements, OFX being the top level aggregate.

### Elements

Elements are used to contain data and can not nest other elements. These are nested inside aggegates.

The most common issue with OFX data files from banks is missing starting or closing tags. The library parses this data with a XML decoder and iterates through parsed tokens individually.

For elements, a reference to each starting tag is held till it is matched with a corresponding ending tag. When a character data token is parsed, it is assumed that it follows a starting element tag immediately. Based on the starting tag refenence, either a missing starting tag or ending tag is inferred and the missing tag is created and inserted.

As an example, this data is missing closing tags for elements.

```xml
<OFX>
    <STATUS>
    	<CODE>0
        <SEVERITY>INFO
    </STATUS>
    <DTSERVER>20191027065402
    <LANGUAGE>ENG
</OFX>
```

The library will generate the following cleaned up XML for this example, before unmarshalling

```xml
<OFX>
    <STATUS>
    	<CODE>0</CODE>
        <SEVERITY>INFO</CODE>
    </STATUS>
    <DTSERVER>20191027065402</DTSERVER>
    <LANGUAGE>ENG</LANGUAGE>
</OFX>
```

This works the same for missing starting tags.

For aggregates, it maintains a stack of tags, adding each aggregates start tag to the stack till the corresponding ending tag is found and inserts any missing closed tags by dequeueing from the stack. Only closing tags can be inferred for aggregates

As an example, this data is missing closing aggregate tags

```xml
<OFX>
    <SIGNONMSGSRSV1>
</OFX>
```

The library will generate the following cleaned up XML for this example.

```xml
<OFX>
    <SIGNONMSGSRSV1></SIGNONMSGSRSV1>
</OFX>
```
