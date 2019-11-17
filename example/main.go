package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/rockstardevs/goofx"
)

func main() {
	data := []byte(`
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
	`)
	reader := bytes.NewReader(data)

	document, err := goofx.NewDocumentFromXML(reader, goofx.NewCleaner())
	if err != nil {
		log.Fatalf("error parsing data file - %s", err)
	}
	fmt.Printf("%#+v", document)
}
