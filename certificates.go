package corvonfse

import (
	"bytes"
	"context"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
)

// Certificates manages the A1 (PKCS#12) certificates of a company.
type Certificates struct {
	t *Transport
}

// Upload posts a certificate as multipart/form-data with the fields
// "file" (PKCS#12 bytes, max 50 KiB) and "password". An empty fileName
// defaults to "certificate.p12". The SDK builds the multipart body —
// callers never touch it.
//
// POST /companies/:companyId/certificates — unwrap "data".
func (c *Certificates) Upload(ctx context.Context, companyID string, file []byte, password string, fileName string) (Certificate, *CorvoError) {
	if fileName == "" {
		fileName = "certificate.p12"
	}

	var form bytes.Buffer
	writer := multipart.NewWriter(&form)

	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", mime.FormatMediaType("form-data", map[string]string{
		"name":     "file",
		"filename": fileName,
	}))
	header.Set("Content-Type", "application/x-pkcs12")

	if err := writeMultipartFile(writer, header, file, password); err != nil {
		return Certificate{}, err
	}

	return requestJSON[Certificate](ctx, c.t, requestSpec{
		method:      http.MethodPost,
		path:        "/companies/" + url.PathEscape(companyID) + "/certificates",
		body:        bytes.NewReader(form.Bytes()),
		contentType: writer.FormDataContentType(),
	})
}

// writeMultipartFile appends the "file" part and the "password" field to
// the multipart writer, mapping construction failures to *CorvoError.
func writeMultipartFile(writer *multipart.Writer, header textproto.MIMEHeader, file []byte, password string) *CorvoError {
	buildError := func(err error) *CorvoError {
		return &CorvoError{
			Status:  0,
			Code:    "encode_error",
			Message: "Failed to build multipart body: " + err.Error(),
		}
	}
	part, err := writer.CreatePart(header)
	if err != nil {
		return buildError(err)
	}
	if _, err := part.Write(file); err != nil {
		return buildError(err)
	}
	if err := writer.WriteField("password", password); err != nil {
		return buildError(err)
	}
	if err := writer.Close(); err != nil {
		return buildError(err)
	}
	return nil
}

// List returns one page of certificates for a company.
// GET /companies/:companyId/certificates — unwrap "none".
func (c *Certificates) List(ctx context.Context, companyID string, params *ListParams) ([]Certificate, Pagination, *CorvoError) {
	return requestPage[Certificate](ctx, c.t,
		"/companies/"+url.PathEscape(companyID)+"/certificates", listQuery(params), false)
}

// Remove deletes a certificate. 204 on success.
// DELETE /companies/:companyId/certificates/:certId.
func (c *Certificates) Remove(ctx context.Context, companyID string, certificateID string) *CorvoError {
	_, err := requestJSON[any](ctx, c.t, requestSpec{
		method: http.MethodDelete,
		path:   "/companies/" + url.PathEscape(companyID) + "/certificates/" + url.PathEscape(certificateID),
	})
	return err
}
