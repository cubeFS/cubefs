// Code generated by smithy-go-codegen DO NOT EDIT.

package smithyrpcv2cbor

import (
	"context"
	"fmt"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// This operation has three possible return values:
//
//   - A successful response in the form of GreetingWithErrorsOutput
//   - An InvalidGreeting error.
//   - A ComplexError error.
//
// Implementations must be able to successfully take a response and properly
// deserialize successful and error responses.
func (c *Client) GreetingWithErrors(ctx context.Context, params *GreetingWithErrorsInput, optFns ...func(*Options)) (*GreetingWithErrorsOutput, error) {
	if params == nil {
		params = &GreetingWithErrorsInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "GreetingWithErrors", params, optFns, c.addOperationGreetingWithErrorsMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*GreetingWithErrorsOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type GreetingWithErrorsInput struct {
	noSmithyDocumentSerde
}

type GreetingWithErrorsOutput struct {
	Greeting *string

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationGreetingWithErrorsMiddlewares(stack *middleware.Stack, options Options) (err error) {
	if err := stack.Serialize.Add(&setOperationInputMiddleware{}, middleware.After); err != nil {
		return err
	}
	err = stack.Serialize.Add(&smithyRpcv2cbor_serializeOpGreetingWithErrors{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&smithyRpcv2cbor_deserializeOpGreetingWithErrors{}, middleware.After)
	if err != nil {
		return err
	}
	if err := addProtocolFinalizerMiddlewares(stack, options, "GreetingWithErrors"); err != nil {
		return fmt.Errorf("add protocol finalizers: %v", err)
	}

	if err = addlegacyEndpointContextSetter(stack, options); err != nil {
		return err
	}
	if err = addSetLoggerMiddleware(stack, options); err != nil {
		return err
	}
	if err = addClientRequestID(stack); err != nil {
		return err
	}
	if err = addComputeContentLength(stack); err != nil {
		return err
	}
	if err = addResolveEndpointMiddleware(stack, options); err != nil {
		return err
	}
	if err = addRetry(stack, options); err != nil {
		return err
	}
	if err = addRawResponseToMetadata(stack); err != nil {
		return err
	}
	if err = addRecordResponseTiming(stack); err != nil {
		return err
	}
	if err = addClientUserAgent(stack, options); err != nil {
		return err
	}
	if err = smithyhttp.AddErrorCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = addSetLegacyContextSigningOptionsMiddleware(stack); err != nil {
		return err
	}
	if err = addTimeOffsetBuild(stack, c); err != nil {
		return err
	}
	if err = addUserAgentRetryMode(stack, options); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opGreetingWithErrors(options.Region), middleware.Before); err != nil {
		return err
	}
	if err = addRecursionDetection(stack); err != nil {
		return err
	}
	if err = addRequestIDRetrieverMiddleware(stack); err != nil {
		return err
	}
	if err = addResponseErrorMiddleware(stack); err != nil {
		return err
	}
	if err = addRequestResponseLogging(stack, options); err != nil {
		return err
	}
	if err = addDisableHTTPSMiddleware(stack, options); err != nil {
		return err
	}
	return nil
}

func newServiceMetadataMiddleware_opGreetingWithErrors(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		OperationName: "GreetingWithErrors",
	}
}