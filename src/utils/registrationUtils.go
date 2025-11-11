package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"perform-cli-framework-go/src/logger"

	"net/http"
	"perform-cli-framework-go/src/conf"
	"time"
)

type RegistrationUtils struct {
	client   *http.Client
	name     string
	endpoint string
}

func NewRegistrationUtils() *RegistrationUtils {
	return &RegistrationUtils{
		client: &http.Client{
			Timeout: 10 * time.Second, // 设置超时时间
		},
	}
}

func (ru *RegistrationUtils) Register(line conf.BenchConfig) error {
	if line.GrpcCfg.RegistrationCtEndpoint == "" {
		return nil
	}
	logger.Info("Register to remote controller: %s", line.GrpcCfg.RegistrationCtEndpoint)
	jsonString, err := ru.getString(line)
	if err != nil {
		return err
	}
	ru.endpoint = line.GrpcCfg.RegistrationCtEndpoint
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/executor/add", ru.endpoint), bytes.NewBuffer([]byte(jsonString)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := ru.client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	var responseMap map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseMap); err != nil {
		return err
	}
	logger.Info("R body: %v", responseMap)
	// 检查 code 是否为 0
	if code, ok := responseMap["code"]; ok {
		if v, ok := code.(float64); ok && int(v) == 0 {
			return nil
		}
	}
	return fmt.Errorf("unexpected response: %v", responseMap)
}

func (ru *RegistrationUtils) getString(line conf.BenchConfig) (string, error) {
	ru.name = line.GrpcCfg.Name
	js, err := json.Marshal(line)
	if err != nil {
		return "", err
	}
	data := map[string]interface{}{
		"name":            ru.name,
		"host":            fmt.Sprintf("%s:%d", line.GrpcCfg.LocalIp, line.GrpcCfg.Port),
		"group_name":      line.GrpcCfg.GroupName,
		"executor_config": string(js),
		"description":     "Executor auto add",
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

func (ru *RegistrationUtils) Unregister() error {
	if ru.name == "" {
		return nil
	}
	logger.Info("unRegister from remote controller")
	jsonString := fmt.Sprintf("{\"name\": \"%s\"}", ru.name)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/executor/del", ru.endpoint), bytes.NewBuffer([]byte(jsonString)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := ru.client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	var responseMap map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseMap); err != nil {
		return err
	}

	logger.Info("UR body: %v", responseMap)

	return nil
}
