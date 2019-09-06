package graphql

import (
	"context"
	"errors"

	"github.com/graphql-go/graphql/language/ast"

	"github.com/spaceuptech/space-cloud/model"
	"github.com/spaceuptech/space-cloud/utils"
)

func (graph *Module) execReadRequest(field *ast.Field, token string, store utils.M) (interface{}, error) {
	dbType := getDBType(field)
	col, err := getCollection(field)
	if err != nil {
		return nil, err
	}

	req, err := generateReadRequest(field, store)
	if err != nil {
		return nil, err
	}

	if _, err := graph.auth.IsReadOpAuthorised(graph.project, dbType, col, token, req); err != nil {
		return nil, err
	}

	return graph.crud.Read(context.TODO(), dbType, graph.project, col, req)
}

func generateReadRequest(field *ast.Field, store utils.M) (*model.ReadRequest, error) {
	var err error

	// Create a read request object
	readRequest := model.ReadRequest{Operation: utils.All, Options: new(model.ReadOptions)}

	readRequest.Find, err = extractWhereClause(field.Arguments, store)
	if err != nil {
		return nil, err
	}

	readRequest.Options, err = generateOptions(field.Arguments, store)
	if err != nil {
		return nil, err
	}

	return &readRequest, nil
}

func extractWhereClause(args []*ast.Argument, store utils.M) (map[string]interface{}, error) {
	for _, v := range args {
		switch v.Name.Value {
		case "where":
			temp, err := ParseValue(v.Value, store)
			if err != nil {
				return nil, err
			}
			if obj, ok := temp.(utils.M); ok {
				return obj, nil
			}
			if obj, ok := temp.(map[string]interface{}); ok {
				return obj, nil
			}
			return nil, errors.New("Invalid where clause provided")
		}
	}

	return utils.M{}, nil
}

func generateOptions(args []*ast.Argument, store utils.M) (*model.ReadOptions, error) {
	options := model.ReadOptions{}
	for _, v := range args {
		switch v.Name.Value {
		case "skip":
			temp, err := ParseValue(v.Value, store)
			if err != nil {
				return nil, err
			}

			tempInt, ok := temp.(int)
			if !ok {
				return nil, errors.New("Invalid type for skip")
			}

			tempInt64 := int64(tempInt)
			options.Skip = &tempInt64

		case "limit":
			temp, err := ParseValue(v.Value, store)
			if err != nil {
				return nil, err
			}

			tempInt, ok := temp.(int)
			if !ok {
				return nil, errors.New("Invalid type for skip")
			}

			tempInt64 := int64(tempInt)
			options.Limit = &tempInt64

			// TODO: implement sort, distinct, etc.
		}
	}
	return &options, nil
}
