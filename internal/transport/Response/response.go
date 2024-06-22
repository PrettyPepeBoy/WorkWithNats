package Response

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
)

type okResponse struct {
	StatusCode int    `json:"statusCode"`
	Data       []byte `json:"data"`
}

type errorResponse struct {
	StatusCode int   `json:"statusCode"`
	Error      error `json:"error"`
}

func OkResponse(status int, data []byte) []byte {
	resp := okResponse{
		StatusCode: status,
		Data:       data,
	}
	rawbyte, err := json.Marshal(resp)
	if err != nil {
		logrus.Error("[Response.OkResponse] failed to marshal json, error ", err)
		return nil
	}
	return rawbyte
}

func ErrorResponse(status int, err error) []byte {
	resp := errorResponse{
		StatusCode: status,
		Error:      err,
	}
	rawbyte, err := json.Marshal(resp)
	if err != nil {
		logrus.Error("[Response.ErrorResponse] failed to marshal json, error ", err)
		return nil
	}
	return rawbyte
}
