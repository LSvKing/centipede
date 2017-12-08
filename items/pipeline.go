package items

import "net/http"

type (
	DataCache []DataRow

	Data struct {
		Field string
		Value interface{}
	}

	D []Data

	DataRow struct {
		Collection string
		Data       []Data
	}

	FileRow struct {
		Path string
		File
	}

	File struct {
		FileName string
		Response *http.Response
	}
)
