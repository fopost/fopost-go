package fopost

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
)

type fileField struct {
	field    string
	filename string
	reader   io.Reader
}

type multipartBody struct {
	body        []byte
	contentType string
}

// buildMultipart buffers the whole form, so a retry can replay it.
func buildMultipart(fields map[string]string, files []fileField) (*multipartBody, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	for name, value := range fields {
		if value == "" {
			continue
		}
		if err := writer.WriteField(name, value); err != nil {
			return nil, fmt.Errorf("fopost: writing form field %q: %w", name, err)
		}
	}

	for _, file := range files {
		if file.reader == nil {
			return nil, fmt.Errorf("fopost: file %q has no content", file.filename)
		}
		part, err := writer.CreateFormFile(file.field, file.filename)
		if err != nil {
			return nil, fmt.Errorf("fopost: creating form file %q: %w", file.filename, err)
		}
		if _, err := io.Copy(part, file.reader); err != nil {
			return nil, fmt.Errorf("fopost: reading %q: %w", file.filename, err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("fopost: closing multipart form: %w", err)
	}
	return &multipartBody{body: buf.Bytes(), contentType: writer.FormDataContentType()}, nil
}
