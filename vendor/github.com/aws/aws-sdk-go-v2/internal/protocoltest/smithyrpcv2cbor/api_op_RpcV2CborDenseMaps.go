// Code generated by smithy-go-codegen DO NOT EDIT.

package smithyrpcv2cbor

import (
	"context"
	"fmt"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/internal/protocoltest/smithyrpcv2cbor/types"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// The example tests basic map serialization.
func (c *Client) RpcV2CborDenseMaps(ctx context.Context, params *RpcV2CborDenseMapsInput, optFns ...func(*Options)) (*RpcV2CborDenseMapsOutput, error) {
	if params == nil {
		params = &RpcV2CborDenseMapsInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "RpcV2CborDenseMaps", params, optFns, c.addOperationRpcV2CborDenseMapsMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*RpcV2CborDenseMapsOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type RpcV2CborDenseMapsInput struct {
	DenseBooleanMap map[string]bool

	DenseNumberMap map[string]int32

	DenseSetMap map[string][]string

	DenseStringMap map[string]string

	DenseStructMap map[string]types.GreetingStruct

	noSmithyDocumentSerde
}

type RpcV2CborDenseMapsOutput struct {
	DenseBooleanMap map[string]bool

	DenseNumberMap map[string]int32

	DenseSetMap map[string][]string

	DenseStringMap map[string]string

	DenseStructMap map[string]types.GreetingStruct

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationRpcV2CborDenseMapsMiddlewares(stack *middleware.Stack, options Options) (err error) {
	if err := stack.Serialize.Add(&setOperationInputMiddleware{}, middleware.After); err != nil {
		return err
	}
	err = stack.Serialize.Add(&smithyRpcv2cbor_serializeOpRpcV2CborDenseMaps{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&smithyRpcv2cbor_deserializeOpRpcV2CborDenseMaps{}, middleware.After)
	if err != nil {
		return err
	}
	if err := addProtocolFinalizerMiddlewares(stack, options, "RpcV2CborDenseMaps"); err != nil {
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
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opRpcV2CborDenseMaps(options.Region), middleware.Before); err != nil {
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

func newServiceMetadataMiddleware_opRpcV2CborDenseMaps(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		OperationName: "RpcV2CborDenseMaps",
	}
}