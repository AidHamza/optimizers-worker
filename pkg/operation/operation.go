package operation

import (
	proto "github.com/golang/protobuf/proto"
)

func LoadOperation(operationBytes []byte) (*Operation, error) {
	operation := &Operation{}

	if err := proto.Unmarshal(operationBytes, operation); err != nil {
		return operation, err
	}

	return operation, nil
}