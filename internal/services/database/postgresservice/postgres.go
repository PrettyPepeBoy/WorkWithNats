package postgresservice

import (
	"TestTaskNats/internal/database/postgres"
	"TestTaskNats/internal/models"
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
)

func PutDataInTable(strg *postgres.Storage, product models.ProductBody) {
	data, err := json.Marshal(product)
	if err != nil {
		logrus.Error("[postgresservice.PutDataInTable] failed to unmarshal json, error ", err)
		return
	}
	id, err := strg.PutInTable(product.Name, data)
	if err != nil {
		logrus.Error("failed to put request body in database, error ", err)
		return
	}
	logrus.Infof("added new product with id: %v", id)
}

func GetDataFromTable(strg *postgres.Storage, id int) ([]byte, error) {
	data, err := strg.GetFromTable(id)
	if err != nil {
		if errors.Is(err, postgres.ErrRowNotExist) {
			return nil, nil
		}
		logrus.Error(" [postgresservice.GetDataFromTable] failed to get resp from database, error ", err)
		return nil, err
	}
	return data, nil
}

func CheckInTable(strg *postgres.Storage) ([]interface{}, error) {
	//todo Creat CheckInTableHandler
	resp, err := strg.CheckInTable()
	if err != nil {
		return nil, err
	}
	return resp, err
}
