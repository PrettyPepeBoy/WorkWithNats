package cacheservice

import (
	"TestTaskNats/internal/cache"
	"TestTaskNats/internal/models"
	"encoding/json"
	"github.com/sirupsen/logrus"
)

func PutInCache(c *cache.Cache, data []byte) error {
	const op = "cacheservice/PutInCache"
	var product models.ProductIDAndBody
	err := json.Unmarshal(data, &product)
	if err != nil {
		logrus.Errorf("failed to unmarsahl request body, error: %v, location: %v", err, op)
		return err
	}
	c.PutKey(product.ID, product.ProductBody)
	return nil
}

func FindInCache(c *cache.Cache, data []byte) (bool, error) {
	const op = "cacheservice/FindInCache"
	var product models.ProductID
	err := json.Unmarshal(data, &product)
	if err != nil {
		logrus.Errorf("failed to unmarshal request body, error: %v, location: %v", err, op)
		return false, err
	}
	val := c.ShowKey(product.ID)
	if val.Name == "" {
		logrus.Infof("key %v was not found in cacheservice", product.ID)
		return false, nil
	}
	logrus.Infof("find key %v with value %v", product.ID, val)
	return true, nil
}
