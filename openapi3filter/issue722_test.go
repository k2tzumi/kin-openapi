package openapi3filter_test

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func Example_validateMultipartFormDataContainingZipFile() {
	const spec = `
openapi: 3.0.0
info:
  title: 'Validator'
  version: 0.0.1
paths:
  /test:
    post:
      requestBody:
        required: true
        content:
          multipart/form-data:
            schema:
              type: object
              required:
                - file
              allOf:
              - $ref: '#/components/schemas/Category'
              - properties:
                  file:
                    type: string
                    format: binary
                  description:
                    type: string
      responses:
        '200':
          description: Created

components:
  schemas:
    Category:
      type: object
      properties:
        name:
          type: string
      required:
      - name
`

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	if err != nil {
		panic(err)
	}
	if err = doc.Validate(loader.Context); err != nil {
		panic(err)
	}

	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		panic(err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)


	{ // Add file data
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="file"; filename="hello.zip"`)
		h.Set("Content-Type", "application/zip")

		fw, err := writer.CreatePart(h)
		if err != nil {
			panic(err)
		}
		zip := []byte{0x50, 0x4b, 0x05, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

		if _, err = io.Copy(fw, bytes.NewReader(zip)); err != nil {
			panic(err)
		}
	}

	{ // Add a single "categories" item as part data
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="categories"`)
		h.Set("Content-Type", "application/json")
		fw, err := writer.CreatePart(h)
		if err != nil {
			panic(err)
		}
		if _, err = io.Copy(fw, strings.NewReader(`{"name": "foo"}`)); err != nil {
			panic(err)
		}
	}

	{ // Add a single "categories" item as part data, again
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="categories"`)
		h.Set("Content-Type", "application/json")
		fw, err := writer.CreatePart(h)
		if err != nil {
			panic(err)
		}
		if _, err = io.Copy(fw, strings.NewReader(`{"name": "bar"}`)); err != nil {
			panic(err)
		}
	}

	{ // Add a single "discription" item as part data
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="description"`)
		fw, err := writer.CreatePart(h)
		if err != nil {
			panic(err)
		}
		if _, err = io.Copy(fw, strings.NewReader(`description note`)); err != nil {
			panic(err)
		}
	}

	writer.Close()

	req, err := http.NewRequest(http.MethodPost, "/test", bytes.NewReader(body.Bytes()))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		panic(err)
	}

	if err = openapi3filter.ValidateRequestBody(
		context.Background(),
		&openapi3filter.RequestValidationInput{
			Request:    req,
			PathParams: pathParams,
			Route:      route,
		},
		route.Operation.RequestBody.Value,
	); err != nil {
		panic(err)
	}
	// Output:
}
