package items

import (
	"io"
	"net/http"
)

type (
	DataCache []DataRow

	Data []struct {
		Field string
		Value interface{}
	}

	DataRow struct {
		Collection string
		Data
	}

	FileRow struct {
		Path string
		File
	}

	File struct {
		Body     io.ReadCloser
		FileName string
		Response *http.Response
	}
)
