package missinggo

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
	haveWritten bool
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if w.haveWritten {
		goto write
	}
	w.haveWritten = true
	if w.Header().Get("Content-Type") != "" {
		goto write
	}
	if type_ := http.DetectContentType(b); type_ != "application/octet-stream" {
		w.Header().Set("Content-Type", type_)
	}
write:
	return w.Writer.Write(b)
}

// Gzips response body if the request says it'll allow it.
func GzipHTTPHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		h.ServeHTTP(&gzipResponseWriter{
			Writer:         gz,
			ResponseWriter: w,
		}, r)
	})
}
