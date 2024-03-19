package ginx

import (
	"encoding/csv"
	"fmt"
	"net/http"
)

// CSV struct.
type CSV struct {
	Data  [][]string
	Title string
}

var csvContentType = []string{"text/csv; charset=utf-8"}

// Render (CSV) writes data with CSV ContentType.
func (c CSV) Render(w http.ResponseWriter) (err error) {
	c.WriteContentType(w)

	writer := csv.NewWriter(w)

	// add utf bom
	w.Write([]byte{0xEF, 0xBB, 0xBF})

	if err = writer.WriteAll(c.Data); err != nil {
		return
	}
	return
}

// WriteContentType write CSV ContentType.
func (c CSV) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = csvContentType
	}

	header["Content-Disposition"] = append(header["Content-Disposition"],
		fmt.Sprintf("attachment; filename=%s.csv", c.Title))
}
