// Code generated by smithy-go-codegen DO NOT EDIT.

package secretsmanager

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	internalauth "github.com/aws/aws-sdk-go-v2/internal/auth"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// Creates a new version with a new encrypted secret value and attaches it to the
// secret. The version can contain a new SecretString value or a new SecretBinary
// value. We recommend you avoid calling PutSecretValue at a sustained rate of
// more than once every 10 minutes. When you update the secret value, Secrets
// Manager creates a new version of the secret. Secrets Manager removes outdated
// versions when there are more than 100, but it does not remove versions created
// less than 24 hours ago. If you call PutSecretValue more than once every 10
// minutes, you create more versions than Secrets Manager removes, and you will
// reach the quota for secret versions. You can specify the staging labels to
// attach to the new version in VersionStages . If you don't include VersionStages
// , then Secrets Manager automatically moves the staging label AWSCURRENT to this
// version. If this operation creates the first version for the secret, then
// Secrets Manager automatically attaches the staging label AWSCURRENT to it. If
// this operation moves the staging label AWSCURRENT from another version to this
// version, then Secrets Manager also automatically moves the staging label
// AWSPREVIOUS to the version that AWSCURRENT was removed from. This operation is
// idempotent. If you call this operation with a ClientRequestToken that matches
// an existing version's VersionId, and you specify the same secret data, the
// operation succeeds but does nothing. However, if the secret data is different,
// then the operation fails because you can't modify an existing version; you can
// only create new ones. Secrets Manager generates a CloudTrail log entry when you
// call this action. Do not include sensitive information in request parameters
// except SecretBinary or SecretString because it might be logged. For more
// information, see Logging Secrets Manager events with CloudTrail (https://docs.aws.amazon.com/secretsmanager/latest/userguide/retrieve-ct-entries.html)
// . Required permissions: secretsmanager:PutSecretValue . For more information,
// see IAM policy actions for Secrets Manager (https://docs.aws.amazon.com/secretsmanager/latest/userguide/reference_iam-permissions.html#reference_iam-permissions_actions)
// and Authentication and access control in Secrets Manager (https://docs.aws.amazon.com/secretsmanager/latest/userguide/auth-and-access.html)
// .
func (c *Client) PutSecretValue(ctx context.Context, params *PutSecretValueInput, optFns ...func(*Options)) (*PutSecretValueOutput, error) {
	if params == nil {
		params = &PutSecretValueInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "PutSecretValue", params, optFns, c.addOperationPutSecretValueMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*PutSecretValueOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type PutSecretValueInput struct {

	// The ARN or name of the secret to add a new version to. For an ARN, we recommend
	// that you specify a complete ARN rather than a partial ARN. See Finding a secret
	// from a partial ARN (https://docs.aws.amazon.com/secretsmanager/latest/userguide/troubleshoot.html#ARN_secretnamehyphen)
	// . If the secret doesn't already exist, use CreateSecret instead.
	//
	// This member is required.
	SecretId *string

	// A unique identifier for the new version of the secret. If you use the Amazon
	// Web Services CLI or one of the Amazon Web Services SDKs to call this operation,
	// then you can leave this parameter empty. The CLI or SDK generates a random UUID
	// for you and includes it as the value for this parameter in the request. If you
	// generate a raw HTTP request to the Secrets Manager service endpoint, then you
	// must generate a ClientRequestToken and include it in the request. This value
	// helps ensure idempotency. Secrets Manager uses this value to prevent the
	// accidental creation of duplicate versions if there are failures and retries
	// during a rotation. We recommend that you generate a UUID-type (https://wikipedia.org/wiki/Universally_unique_identifier)
	// value to ensure uniqueness of your versions within the specified secret.
	//   - If the ClientRequestToken value isn't already associated with a version of
	//   the secret then a new version of the secret is created.
	//   - If a version with this value already exists and that version's SecretString
	//   or SecretBinary values are the same as those in the request then the request
	//   is ignored. The operation is idempotent.
	//   - If a version with this value already exists and the version of the
	//   SecretString and SecretBinary values are different from those in the request,
	//   then the request fails because you can't modify a secret version. You can only
	//   create new versions to store new secret values.
	// This value becomes the VersionId of the new version.
	ClientRequestToken *string

	// The binary data to encrypt and store in the new version of the secret. To use
	// this parameter in the command-line tools, we recommend that you store your
	// binary data in a file and then pass the contents of the file as a parameter. You
	// must include SecretBinary or SecretString , but not both. You can't access this
	// value from the Secrets Manager console.
	SecretBinary []byte

	// The text to encrypt and store in the new version of the secret. You must
	// include SecretBinary or SecretString , but not both. We recommend you create the
	// secret string as JSON key/value pairs, as shown in the example.
	SecretString *string

	// A list of staging labels to attach to this version of the secret. Secrets
	// Manager uses staging labels to track versions of a secret through the rotation
	// process. If you specify a staging label that's already associated with a
	// different version of the same secret, then Secrets Manager removes the label
	// from the other version and attaches it to this version. If you specify
	// AWSCURRENT , and it is already attached to another version, then Secrets Manager
	// also moves the staging label AWSPREVIOUS to the version that AWSCURRENT was
	// removed from. If you don't include VersionStages , then Secrets Manager
	// automatically moves the staging label AWSCURRENT to this version.
	VersionStages []string

	noSmithyDocumentSerde
}

type PutSecretValueOutput struct {

	// The ARN of the secret.
	ARN *string

	// The name of the secret.
	Name *string

	// The unique identifier of the version of the secret.
	VersionId *string

	// The list of staging labels that are currently attached to this version of the
	// secret. Secrets Manager uses staging labels to track a version as it progresses
	// through the secret rotation process.
	VersionStages []string

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationPutSecretValueMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsAwsjson11_serializeOpPutSecretValue{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsAwsjson11_deserializeOpPutSecretValue{}, middleware.After)
	if err != nil {
		return err
	}
	if err = addlegacyEndpointContextSetter(stack, options); err != nil {
		return err
	}
	if err = addSetLoggerMiddleware(stack, options); err != nil {
		return err
	}
	if err = awsmiddleware.AddClientRequestIDMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddComputeContentLengthMiddleware(stack); err != nil {
		return err
	}
	if err = addResolveEndpointMiddleware(stack, options); err != nil {
		return err
	}
	if err = v4.AddComputePayloadSHA256Middleware(stack); err != nil {
		return err
	}
	if err = addRetryMiddlewares(stack, options); err != nil {
		return err
	}
	if err = addHTTPSignerV4Middleware(stack, options); err != nil {
		return err
	}
	if err = awsmiddleware.AddRawResponseToMetadata(stack); err != nil {
		return err
	}
	if err = awsmiddleware.AddRecordResponseTiming(stack); err != nil {
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
	if err = addPutSecretValueResolveEndpointMiddleware(stack, options); err != nil {
		return err
	}
	if err = addIdempotencyToken_opPutSecretValueMiddleware(stack, options); err != nil {
		return err
	}
	if err = addOpPutSecretValueValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opPutSecretValue(options.Region), middleware.Before); err != nil {
		return err
	}
	if err = awsmiddleware.AddRecursionDetection(stack); err != nil {
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
	if err = addendpointDisableHTTPSMiddleware(stack, options); err != nil {
		return err
	}
	return nil
}

type idempotencyToken_initializeOpPutSecretValue struct {
	tokenProvider IdempotencyTokenProvider
}

func (*idempotencyToken_initializeOpPutSecretValue) ID() string {
	return "OperationIdempotencyTokenAutoFill"
}

func (m *idempotencyToken_initializeOpPutSecretValue) HandleInitialize(ctx context.Context, in middleware.InitializeInput, next middleware.InitializeHandler) (
	out middleware.InitializeOutput, metadata middleware.Metadata, err error,
) {
	if m.tokenProvider == nil {
		return next.HandleInitialize(ctx, in)
	}

	input, ok := in.Parameters.(*PutSecretValueInput)
	if !ok {
		return out, metadata, fmt.Errorf("expected middleware input to be of type *PutSecretValueInput ")
	}

	if input.ClientRequestToken == nil {
		t, err := m.tokenProvider.GetIdempotencyToken()
		if err != nil {
			return out, metadata, err
		}
		input.ClientRequestToken = &t
	}
	return next.HandleInitialize(ctx, in)
}
func addIdempotencyToken_opPutSecretValueMiddleware(stack *middleware.Stack, cfg Options) error {
	return stack.Initialize.Add(&idempotencyToken_initializeOpPutSecretValue{tokenProvider: cfg.IdempotencyTokenProvider}, middleware.Before)
}

func newServiceMetadataMiddleware_opPutSecretValue(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "secretsmanager",
		OperationName: "PutSecretValue",
	}
}

type opPutSecretValueResolveEndpointMiddleware struct {
	EndpointResolver EndpointResolverV2
	BuiltInResolver  builtInParameterResolver
}

func (*opPutSecretValueResolveEndpointMiddleware) ID() string {
	return "ResolveEndpointV2"
}

func (m *opPutSecretValueResolveEndpointMiddleware) HandleSerialize(ctx context.Context, in middleware.SerializeInput, next middleware.SerializeHandler) (
	out middleware.SerializeOutput, metadata middleware.Metadata, err error,
) {
	if awsmiddleware.GetRequiresLegacyEndpoints(ctx) {
		return next.HandleSerialize(ctx, in)
	}

	req, ok := in.Request.(*smithyhttp.Request)
	if !ok {
		return out, metadata, fmt.Errorf("unknown transport type %T", in.Request)
	}

	if m.EndpointResolver == nil {
		return out, metadata, fmt.Errorf("expected endpoint resolver to not be nil")
	}

	params := EndpointParameters{}

	m.BuiltInResolver.ResolveBuiltIns(&params)

	var resolvedEndpoint smithyendpoints.Endpoint
	resolvedEndpoint, err = m.EndpointResolver.ResolveEndpoint(ctx, params)
	if err != nil {
		return out, metadata, fmt.Errorf("failed to resolve service endpoint, %w", err)
	}

	req.URL = &resolvedEndpoint.URI

	for k := range resolvedEndpoint.Headers {
		req.Header.Set(
			k,
			resolvedEndpoint.Headers.Get(k),
		)
	}

	authSchemes, err := internalauth.GetAuthenticationSchemes(&resolvedEndpoint.Properties)
	if err != nil {
		var nfe *internalauth.NoAuthenticationSchemesFoundError
		if errors.As(err, &nfe) {
			// if no auth scheme is found, default to sigv4
			signingName := "secretsmanager"
			signingRegion := m.BuiltInResolver.(*builtInResolver).Region
			ctx = awsmiddleware.SetSigningName(ctx, signingName)
			ctx = awsmiddleware.SetSigningRegion(ctx, signingRegion)

		}
		var ue *internalauth.UnSupportedAuthenticationSchemeSpecifiedError
		if errors.As(err, &ue) {
			return out, metadata, fmt.Errorf(
				"This operation requests signer version(s) %v but the client only supports %v",
				ue.UnsupportedSchemes,
				internalauth.SupportedSchemes,
			)
		}
	}

	for _, authScheme := range authSchemes {
		switch authScheme.(type) {
		case *internalauth.AuthenticationSchemeV4:
			v4Scheme, _ := authScheme.(*internalauth.AuthenticationSchemeV4)
			var signingName, signingRegion string
			if v4Scheme.SigningName == nil {
				signingName = "secretsmanager"
			} else {
				signingName = *v4Scheme.SigningName
			}
			if v4Scheme.SigningRegion == nil {
				signingRegion = m.BuiltInResolver.(*builtInResolver).Region
			} else {
				signingRegion = *v4Scheme.SigningRegion
			}
			if v4Scheme.DisableDoubleEncoding != nil {
				// The signer sets an equivalent value at client initialization time.
				// Setting this context value will cause the signer to extract it
				// and override the value set at client initialization time.
				ctx = internalauth.SetDisableDoubleEncoding(ctx, *v4Scheme.DisableDoubleEncoding)
			}
			ctx = awsmiddleware.SetSigningName(ctx, signingName)
			ctx = awsmiddleware.SetSigningRegion(ctx, signingRegion)
			break
		case *internalauth.AuthenticationSchemeV4A:
			v4aScheme, _ := authScheme.(*internalauth.AuthenticationSchemeV4A)
			if v4aScheme.SigningName == nil {
				v4aScheme.SigningName = aws.String("secretsmanager")
			}
			if v4aScheme.DisableDoubleEncoding != nil {
				// The signer sets an equivalent value at client initialization time.
				// Setting this context value will cause the signer to extract it
				// and override the value set at client initialization time.
				ctx = internalauth.SetDisableDoubleEncoding(ctx, *v4aScheme.DisableDoubleEncoding)
			}
			ctx = awsmiddleware.SetSigningName(ctx, *v4aScheme.SigningName)
			ctx = awsmiddleware.SetSigningRegion(ctx, v4aScheme.SigningRegionSet[0])
			break
		case *internalauth.AuthenticationSchemeNone:
			break
		}
	}

	return next.HandleSerialize(ctx, in)
}

func addPutSecretValueResolveEndpointMiddleware(stack *middleware.Stack, options Options) error {
	return stack.Serialize.Insert(&opPutSecretValueResolveEndpointMiddleware{
		EndpointResolver: options.EndpointResolverV2,
		BuiltInResolver: &builtInResolver{
			Region:       options.Region,
			UseDualStack: options.EndpointOptions.UseDualStackEndpoint,
			UseFIPS:      options.EndpointOptions.UseFIPSEndpoint,
			Endpoint:     options.BaseEndpoint,
		},
	}, "ResolveEndpoint", middleware.After)
}
