package postgresservice

import (
	"TestTaskNats/internal/database/postgres"
	"TestTaskNats/internal/models"
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
)

func PutDataInTable(strg *postgres.Storage, data []byte) {
	const op = "postgresservice/PutDataInTable"
	//todo validate request body
	id, err := strg.PutInTable(data)
	if err != nil {
		logrus.Errorf("failed to put request body in database, error: %v, location: %v", err, op)
		return
	}
	logrus.Infof("added new product with id: %v", id)
}

func GetDataFromTable(strg *postgres.Storage, data []byte) ([]byte, error) {
	const op = "handlers/GetDataFromTable"
	var id models.ProductID
	err := json.Unmarshal(data, &id)
	if err != nil {
		logrus.Errorf("failed to unmarshal request body, error: %v, location: %v", err, op)
		return nil, err
	}
	resp, err := strg.GetFromTable(id.ID)
	if err != nil {
		if errors.Is(err, postgres.ErrRowNotExist) {
			return nil, nil
		}
		logrus.Errorf("failed to get resp from database, error: %v, location: %v", err, op)
		return nil, err
	}
	var product models.ProductIDAndBody
	product.ID = id.ID
	err = json.Unmarshal(resp, &product.ProductBody)
	if err != nil {
		logrus.Errorf("failed to unmarshal data, error: %v, location: %v", err, op)
		return nil, err
	}
	rawbyte, err := json.Marshal(product)
	if err != nil {
		logrus.Errorf("failed to marshal rawbyte, error: %v, location: %v", err, op)
		return nil, err
	}
	return rawbyte, nil
}
