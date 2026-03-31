package dynamodb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// En: Client defines required DynamoDB methods for table cleanup operations.
// Es: Client define metodos DynamoDB para limpiar tablas.
type Client interface {
	DescribeTable(ctx context.Context, params *awsdynamodb.DescribeTableInput, optFns ...func(*awsdynamodb.Options)) (*awsdynamodb.DescribeTableOutput, error)
	Scan(ctx context.Context, params *awsdynamodb.ScanInput, optFns ...func(*awsdynamodb.Options)) (*awsdynamodb.ScanOutput, error)
	BatchWriteItem(ctx context.Context, params *awsdynamodb.BatchWriteItemInput, optFns ...func(*awsdynamodb.Options)) (*awsdynamodb.BatchWriteItemOutput, error)
}

// En: ClientOptions configures the DynamoDB client connection.
// Es: ClientOptions configura la conexion del cliente DynamoDB.
type ClientOptions struct {
	EndpointURL     string
	Region          string
	Profile         string
	AccessKeyID     string
	SecretAccessKey string
}

// En: NewClient creates a DynamoDB client with optional local endpoint.
// Es: NewClient crea cliente DynamoDB con endpoint local opcional.
func NewClient(ctx context.Context, opts ClientOptions) (*awsdynamodb.Client, error) {
	region := strings.TrimSpace(opts.Region)
	if region == "" {
		region = "us-east-1"
	}

	loadOpts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(region),
	}
	if strings.TrimSpace(opts.Profile) != "" {
		loadOpts = append(loadOpts, awsconfig.WithSharedConfigProfile(opts.Profile))
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, loadOpts...)
	if err != nil {
		return nil, fmt.Errorf("load aws config for dynamodb: %w", err)
	}

	if strings.TrimSpace(opts.EndpointURL) != "" && strings.TrimSpace(opts.AccessKeyID) != "" {
		cfg.Credentials = credentials.NewStaticCredentialsProvider(
			opts.AccessKeyID,
			opts.SecretAccessKey,
			"",
		)
	}

	clientOpts := []func(*awsdynamodb.Options){}
	if strings.TrimSpace(opts.EndpointURL) != "" {
		clientOpts = append(clientOpts, func(o *awsdynamodb.Options) {
			o.BaseEndpoint = aws.String(opts.EndpointURL)
		})
	}

	return awsdynamodb.NewFromConfig(cfg, clientOpts...), nil
}

// En: DeleteAllItems removes all items from the DynamoDB table.
// Es: DeleteAllItems elimina todos los items de la tabla DynamoDB.
func DeleteAllItems(ctx context.Context, client Client, tableName string) (int, error) {
	name := strings.TrimSpace(tableName)
	if name == "" {
		return 0, fmt.Errorf("dynamodb table name is required")
	}

	desc, err := client.DescribeTable(ctx, &awsdynamodb.DescribeTableInput{
		TableName: aws.String(name),
	})
	if err != nil {
		return 0, fmt.Errorf("describe table %s: %w", name, err)
	}
	if desc.Table == nil || len(desc.Table.KeySchema) == 0 {
		return 0, fmt.Errorf("table %s key schema is empty", name)
	}

	projection, exprNames, keyNames := projectionFromSchema(desc.Table.KeySchema)
	var (
		totalDeleted int
		startKey     map[string]types.AttributeValue
	)

	for {
		scanOut, scanErr := client.Scan(ctx, &awsdynamodb.ScanInput{
			TableName:                aws.String(name),
			ProjectionExpression:     aws.String(projection),
			ExpressionAttributeNames: exprNames,
			ExclusiveStartKey:        startKey,
			ConsistentRead:           aws.Bool(true),
			ReturnConsumedCapacity:   types.ReturnConsumedCapacityNone,
			Select:                   types.SelectSpecificAttributes,
		})
		if scanErr != nil {
			return totalDeleted, fmt.Errorf("scan table %s: %w", name, scanErr)
		}

		requests := make([]types.WriteRequest, 0, len(scanOut.Items))
		for _, item := range scanOut.Items {
			key := map[string]types.AttributeValue{}
			for _, keyName := range keyNames {
				value, ok := item[keyName]
				if !ok {
					return totalDeleted, fmt.Errorf("missing key attribute %s in scan item", keyName)
				}
				key[keyName] = value
			}
			requests = append(requests, types.WriteRequest{
				DeleteRequest: &types.DeleteRequest{
					Key: key,
				},
			})
		}

		for i := 0; i < len(requests); i += 25 {
			end := i + 25
			if end > len(requests) {
				end = len(requests)
			}
			if err := batchDeleteWithRetry(ctx, client, name, requests[i:end]); err != nil {
				return totalDeleted, err
			}
		}
		totalDeleted += len(requests)

		if len(scanOut.LastEvaluatedKey) == 0 {
			break
		}
		startKey = scanOut.LastEvaluatedKey
	}

	return totalDeleted, nil
}

// En: projectionFromSchema builds scan projection and key attribute list.
// Es: projectionFromSchema crea proyeccion scan y lista de llaves.
func projectionFromSchema(schema []types.KeySchemaElement) (string, map[string]string, []string) {
	exprNames := make(map[string]string, len(schema))
	parts := make([]string, 0, len(schema))
	keys := make([]string, 0, len(schema))

	for i, key := range schema {
		token := fmt.Sprintf("#k%d", i)
		exprNames[token] = aws.ToString(key.AttributeName)
		parts = append(parts, token)
		keys = append(keys, aws.ToString(key.AttributeName))
	}

	return strings.Join(parts, ", "), exprNames, keys
}

// En: batchDeleteWithRetry handles unprocessed delete requests with backoff.
// Es: batchDeleteWithRetry reintenta deletes no procesados con backoff.
func batchDeleteWithRetry(ctx context.Context, client Client, tableName string, reqs []types.WriteRequest) error {
	pending := reqs
	for attempt := 0; attempt < 6; attempt++ {
		out, err := client.BatchWriteItem(ctx, &awsdynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				tableName: pending,
			},
		})
		if err != nil {
			return fmt.Errorf("batch delete on %s: %w", tableName, err)
		}

		next := out.UnprocessedItems[tableName]
		if len(next) == 0 {
			return nil
		}
		pending = next
		time.Sleep(time.Duration(1<<attempt) * 10 * time.Millisecond)
	}

	return fmt.Errorf("batch delete on %s has unprocessed items after retries", tableName)
}
