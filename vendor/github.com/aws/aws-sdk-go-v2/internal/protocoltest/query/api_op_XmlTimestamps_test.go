// Code generated by smithy-go-codegen DO NOT EDIT.

package query

import (
	"bytes"
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/smithy-go/middleware"
	"github.com/aws/smithy-go/ptr"
	smithyrand "github.com/aws/smithy-go/rand"
	smithytesting "github.com/aws/smithy-go/testing"
	smithytime "github.com/aws/smithy-go/time"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestClient_XmlTimestamps_awsAwsqueryDeserialize(t *testing.T) {
	cases := map[string]struct {
		StatusCode    int
		Header        http.Header
		BodyMediaType string
		Body          []byte
		ExpectResult  *XmlTimestampsOutput
	}{
		// Tests how normal timestamps are serialized
		"QueryXmlTimestamps": {
			StatusCode: 200,
			Header: http.Header{
				"Content-Type": []string{"text/xml"},
			},
			BodyMediaType: "application/xml",
			Body: []byte(`<XmlTimestampsResponse xmlns="https://example.com/">
			    <XmlTimestampsResult>
			        <normal>2014-04-29T18:30:38Z</normal>
			    </XmlTimestampsResult>
			</XmlTimestampsResponse>
			`),
			ExpectResult: &XmlTimestampsOutput{
				Normal: ptr.Time(smithytime.ParseEpochSeconds(1398796238)),
			},
		},
		// Ensures that the timestampFormat of date-time works like normal timestamps
		"QueryXmlTimestampsWithDateTimeFormat": {
			StatusCode: 200,
			Header: http.Header{
				"Content-Type": []string{"text/xml"},
			},
			BodyMediaType: "application/xml",
			Body: []byte(`<XmlTimestampsResponse xmlns="https://example.com/">
			    <XmlTimestampsResult>
			        <dateTime>2014-04-29T18:30:38Z</dateTime>
			    </XmlTimestampsResult>
			</XmlTimestampsResponse>
			`),
			ExpectResult: &XmlTimestampsOutput{
				DateTime: ptr.Time(smithytime.ParseEpochSeconds(1398796238)),
			},
		},
		// Ensures that the timestampFormat of date-time on the target shape works like
		// normal timestamps
		"QueryXmlTimestampsWithDateTimeOnTargetFormat": {
			StatusCode: 200,
			Header: http.Header{
				"Content-Type": []string{"text/xml"},
			},
			BodyMediaType: "application/xml",
			Body: []byte(`<XmlTimestampsResponse xmlns="https://example.com/">
			    <XmlTimestampsResult>
			        <dateTimeOnTarget>2014-04-29T18:30:38Z</dateTimeOnTarget>
			    </XmlTimestampsResult>
			</XmlTimestampsResponse>
			`),
			ExpectResult: &XmlTimestampsOutput{
				DateTimeOnTarget: ptr.Time(smithytime.ParseEpochSeconds(1398796238)),
			},
		},
		// Ensures that the timestampFormat of epoch-seconds works
		"QueryXmlTimestampsWithEpochSecondsFormat": {
			StatusCode: 200,
			Header: http.Header{
				"Content-Type": []string{"text/xml"},
			},
			BodyMediaType: "application/xml",
			Body: []byte(`<XmlTimestampsResponse xmlns="https://example.com/">
			    <XmlTimestampsResult>
			        <epochSeconds>1398796238</epochSeconds>
			    </XmlTimestampsResult>
			</XmlTimestampsResponse>
			`),
			ExpectResult: &XmlTimestampsOutput{
				EpochSeconds: ptr.Time(smithytime.ParseEpochSeconds(1398796238)),
			},
		},
		// Ensures that the timestampFormat of epoch-seconds on the target shape works
		"QueryXmlTimestampsWithEpochSecondsOnTargetFormat": {
			StatusCode: 200,
			Header: http.Header{
				"Content-Type": []string{"text/xml"},
			},
			BodyMediaType: "application/xml",
			Body: []byte(`<XmlTimestampsResponse xmlns="https://example.com/">
			    <XmlTimestampsResult>
			        <epochSecondsOnTarget>1398796238</epochSecondsOnTarget>
			    </XmlTimestampsResult>
			</XmlTimestampsResponse>
			`),
			ExpectResult: &XmlTimestampsOutput{
				EpochSecondsOnTarget: ptr.Time(smithytime.ParseEpochSeconds(1398796238)),
			},
		},
		// Ensures that the timestampFormat of http-date works
		"QueryXmlTimestampsWithHttpDateFormat": {
			StatusCode: 200,
			Header: http.Header{
				"Content-Type": []string{"text/xml"},
			},
			BodyMediaType: "application/xml",
			Body: []byte(`<XmlTimestampsResponse xmlns="https://example.com/">
			    <XmlTimestampsResult>
			        <httpDate>Tue, 29 Apr 2014 18:30:38 GMT</httpDate>
			    </XmlTimestampsResult>
			</XmlTimestampsResponse>
			`),
			ExpectResult: &XmlTimestampsOutput{
				HttpDate: ptr.Time(smithytime.ParseEpochSeconds(1398796238)),
			},
		},
		// Ensures that the timestampFormat of http-date on the target shape works
		"QueryXmlTimestampsWithHttpDateOnTargetFormat": {
			StatusCode: 200,
			Header: http.Header{
				"Content-Type": []string{"text/xml"},
			},
			BodyMediaType: "application/xml",
			Body: []byte(`<XmlTimestampsResponse xmlns="https://example.com/">
			    <XmlTimestampsResult>
			        <httpDateOnTarget>Tue, 29 Apr 2014 18:30:38 GMT</httpDateOnTarget>
			    </XmlTimestampsResult>
			</XmlTimestampsResponse>
			`),
			ExpectResult: &XmlTimestampsOutput{
				HttpDateOnTarget: ptr.Time(smithytime.ParseEpochSeconds(1398796238)),
			},
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			serverURL := "http://localhost:8888/"
			client := New(Options{
				HTTPClient: smithyhttp.ClientDoFunc(func(r *http.Request) (*http.Response, error) {
					headers := http.Header{}
					for k, vs := range c.Header {
						for _, v := range vs {
							headers.Add(k, v)
						}
					}
					if len(c.BodyMediaType) != 0 && len(headers.Values("Content-Type")) == 0 {
						headers.Set("Content-Type", c.BodyMediaType)
					}
					response := &http.Response{
						StatusCode: c.StatusCode,
						Header:     headers,
						Request:    r,
					}
					if len(c.Body) != 0 {
						response.ContentLength = int64(len(c.Body))
						response.Body = ioutil.NopCloser(bytes.NewReader(c.Body))
					} else {

						response.Body = http.NoBody
					}
					return response, nil
				}),
				APIOptions: []func(*middleware.Stack) error{
					func(s *middleware.Stack) error {
						s.Finalize.Clear()
						s.Initialize.Remove(`OperationInputValidation`)
						return nil
					},
				},
				EndpointResolver: EndpointResolverFunc(func(region string, options EndpointResolverOptions) (e aws.Endpoint, err error) {
					e.URL = serverURL
					e.SigningRegion = "us-west-2"
					return e, err
				}),
				IdempotencyTokenProvider: smithyrand.NewUUIDIdempotencyToken(&smithytesting.ByteLoop{}),
				Region:                   "us-west-2",
			})
			var params XmlTimestampsInput
			result, err := client.XmlTimestamps(context.Background(), &params)
			if err != nil {
				t.Fatalf("expect nil err, got %v", err)
			}
			if result == nil {
				t.Fatalf("expect not nil result")
			}
			if err := smithytesting.CompareValues(c.ExpectResult, result); err != nil {
				t.Errorf("expect c.ExpectResult value match:\n%v", err)
			}
		})
	}
}