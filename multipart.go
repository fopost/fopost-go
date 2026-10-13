package fopost

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"sort"
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

// streamFile is one part of a streamed form; size must be exact.
type streamFile struct {
	fileField
	size int64
}

// streamMultipart writes the form through a pipe so a large file is never
// held in memory. It returns the exact Content-Length, which it computes by
// writing the same form once with zero-filled parts of the declared sizes.
func streamMultipart(fields map[string]string, files []streamFile) (body io.Reader, contentType string, length int64, err error) {
	for _, file := range files {
		if file.reader == nil {
			return nil, "", 0, fmt.Errorf("fopost: file %q has no content", file.filename)
		}
	}
	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	sort.Strings(names)

	write := func(w io.Writer, boundary string, content func(streamFile) io.Reader) (string, error) {
		writer := multipart.NewWriter(w)
		if boundary != "" {
			if err := writer.SetBoundary(boundary); err != nil {
				return "", err
			}
		}
		for _, name := range names {
			if fields[name] == "" {
				continue
			}
			if err := writer.WriteField(name, fields[name]); err != nil {
				return "", fmt.Errorf("fopost: writing form field %q: %w", name, err)
			}
		}
		for _, file := range files {
			part, err := writer.CreateFormFile(file.field, file.filename)
			if err != nil {
				return "", fmt.Errorf("fopost: creating form file %q: %w", file.filename, err)
			}
			if _, err := io.CopyN(part, content(file), file.size); err != nil {
				return "", fmt.Errorf("fopost: reading %q: %w", file.filename, err)
			}
		}
		if err := writer.Close(); err != nil {
			return "", fmt.Errorf("fopost: closing multipart form: %w", err)
		}
		return writer.Boundary(), nil
	}

	counter := &countingWriter{}
	boundary, err := write(counter, "", func(f streamFile) io.Reader { return zeroReader{} })
	if err != nil {
		return nil, "", 0, err
	}

	pr, pw := io.Pipe()
	go func() {
		_, err := write(pw, boundary, func(f streamFile) io.Reader { return f.reader })
		pw.CloseWithError(err)
	}()
	return pr, "multipart/form-data; boundary=" + boundary, counter.n, nil
}

type countingWriter struct{ n int64 }

func (c *countingWriter) Write(p []byte) (int, error) {
	c.n += int64(len(p))
	return len(p), nil
}

type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}
