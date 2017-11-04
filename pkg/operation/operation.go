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

func NewResult(opId uint64, fileName string, status Result_Status) ([]byte, error) {
	resultBytes, err := proto.Marshal(&Result{
		Id: opId,
		File: fileName,
		Status: status,
	})
	
	if err != nil {
		return []byte{0}, err
	}

	return resultBytes, nil
}