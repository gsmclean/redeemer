package main

import (
	"context"
	"encoding/base64"
	"os"

	"github.com/labstack/gommon/log"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func PushToSheet(sheetId int64, spreadsheetId string, rowData []string) error {
	ctx := context.Background()
	credBytes, err := base64.StdEncoding.DecodeString(os.Getenv("KEY_JSON_BASE64"))
	// credBytes := []byte("-----BEGIN PRIVATE KEY-----\nMIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQCpEETfRdlqX7dx\n99qaxRfAkOENREW9eWPrX9bjOhfUdXj/5FI3XeaEqufURgM8MJKLUPfTrwC0dC/d\nzimI56UNfiwwkwmlc5S7z1Dx+O2aaodA4T4fDKqDEyYRs+cWYWdYhEAohlauNG++\nYOlrSiKr3Chu3jcpOroTgv2So+6BrRtz5fs0UirjObNyWfblvGOiHsD/FAYPVc1U\nW9fJXjNWbmIAtHn3HHWmVurEVaqYGu7F9l45Q2f9ZR8Z4YAp4qYL4RVOeYAF0Dlk\nHfbL99CtxSMqk8zpvnyVcDfDyru9eMm+1JbQaeJFa6HAdtoz01Zjz78Ck9JxnphE\ntkbqonndAgMBAAECggEAGeGEXuxhz3z+Uj4agwhFDepKlC4CwcdsL8N+MU7aswrr\nY7u03QZM8bulaG5HXywjDZzEM5heHUHkH1UeL/fLi/iSvzo4hybd0AUkDnGIaNPX\nXMFJMeujIWm0XDq+15hAVTakXmUiWTdscSfY098yK5O1GjPTx0YBklcX28j+DBpk\nb9tLcHJz+K+Fd34GojLSysz+NQ79CiMe/b9L+GdpYxAofptzDcTia9nOw0nwzgzp\n4na9Y0eyZrf7TRwJ2iEIUAejlSO+h53v/xn0uDWMJbWptvRGCnIQYd9j3vjd19ON\nSK5Vl9a9ViAiZeOd2no4/2omqyhxZGQhFfCBr9BFaQKBgQDsalH8NhuLDwbmaHsH\nVgVSsuhJI0gm3L+tFpxN4OIbdjOIzSUqhmeSzHd1B+hrWBiDlHwLzTU6Z+EhS6LK\nUS25DcXnNZKD8/8up67ldXRvovoMn/5LRAV3ijECo6sQ118M7JFvUWb1WpWsmgWR\nkitQLHur2e25CFnjLYDRRpaGnwKBgQC3EZ076pXXfoJtDZea5DTAvJfdcpG5td+8\n+e9DFOHo9tUMXkaue/PW353mhpGUk/MK118BhE1zUaYHtlgeDV4W//qHVaqExSK4\n2idJFD+Hm7c3fjv3+acessLP2WPGthf5X/l448svIu9TJyO8l/36TPRIgZan+hiY\nDQvtj5VaAwKBgCIhEyba1M0VZUSb7q7XbztKEpiEXGUn1w/wxK3Fej7GqJfmLahe\n8NLTa6dcdeQROrC8HdBCVp3Q40JAPgcBAx3E7D39kOI1tjARCwGbHC0FlR1/d2F8\nN2HTdFHSON7ciJ9AA5rTYI6o/hSFw6oJNPGFCnF7q4LbvsY6Cm+rxg03AoGARIVx\nRsXtQ/V0OAFIZ49XN3TfmuGRLeOnVQJvzbn5PMt2vuRirFh00k5suaZQwz4FUF+A\njf7JRoqfDG/x133FY/J4AUPNSVjIQExXPAE6LjXYhArZw11Mci8Sv91sfSoXGx4T\nMG6C1KfM0GDr/WEejRtUq/blPwZbQj5P4qFFk6UCgYB2oGLh20feFhb5iBvtjivS\nQ6gJy5Dhp8D/7kGNN249H05pB7aw9MvuZ57JnVaJQTukEhmorGM38GreeHi+nmnd\nJQ6jIsw7P0N8UCvOwDvmD5xoylBCTe95qFkxV/1HJxKviLeL8axDiTTrO1dlgB4f\nYnL6Lens4x0e6DZfLreV+A==\n-----END PRIVATE KEY-----\n")
	if err != nil {
		log.Error(err)
		return err
	}
	config, err := google.JWTConfigFromJSON(credBytes, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Error(err)
		return err
	}

	// create client with config and context
	client := config.Client(ctx)

	// create new service using client
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Error(err)
		return err
	}

	response1, err := srv.Spreadsheets.Get(spreadsheetId).Fields("sheets(properties(sheetId,title))").Do()
	if err != nil || response1.HTTPStatusCode != 200 {
		log.Error(err)
		return err
	}

	sheetName := ""
	for _, v := range response1.Sheets {
		prop := v.Properties
		if prop.SheetId == int64(sheetId) {
			sheetName = prop.Title
			break
		}
	}
	newData := make([]interface{}, len(rowData))
	for i, v := range rowData {
		newData[i] = v
	}
	//Append value to the sheet.
	row := &sheets.ValueRange{
		Values: [][]interface{}{newData},
	}

	response2, err := srv.Spreadsheets.Values.Append(spreadsheetId, sheetName, row).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Context(ctx).Do()
	if err != nil || response2.HTTPStatusCode != 200 {
		log.Error(err)
	}
	return err
}
