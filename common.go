/*
 * Copyright 2021 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package polaris

import (
	"fmt"
	"net"
	"strconv"

	perrors "github.com/pkg/errors"
	"github.com/polarismesh/polaris-go/api"
	"github.com/polarismesh/polaris-go/pkg/config"
	"github.com/polarismesh/polaris-go/pkg/model"
)

// instanceInfo used to stored service basic info in polaris.
type instanceInfo struct {
	Network string            `json:"network"`
	Address string            `json:"address"`
	Weight  int               `json:"weight"`
	Tags    map[string]string `json:"tags"`
}

// GetPolarisConfig get polaris config from endpoints
func GetPolarisConfig(endpoints []string) (api.SDKContext, error) {
	if len(endpoints) == 0 {
		return nil, perrors.New("endpoints is empty!")
	}

	serverConfigs := make([]string, 0, len(endpoints))
	for _, addr := range endpoints {
		ip, portStr, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, perrors.WithMessagef(err, "split [%s] ", addr)
		}
		port, _ := strconv.Atoi(portStr)
		serverConfigs = append(serverConfigs, fmt.Sprintf("%s:%d", ip, uint64(port)))
	}

	polarisConf := config.NewDefaultConfiguration(serverConfigs)

	confPath := "polaris.yaml"
	if confPath != "" && model.IsFile(confPath) {
		complexConf, err := config.LoadConfigurationByFile(confPath)
		if err != nil {
			return nil, err
		}
		mergePolarisConfiguration(polarisConf, complexConf)
		polarisConf = complexConf
	}
	sdkCtx, err := api.InitContextByConfig(polarisConf)
	if err != nil {
		return nil, err
	}
	return sdkCtx, nil
}

func mergePolarisConfiguration(easy, complexConf config.Configuration) {
	easySvrList := easy.GetGlobal().GetServerConnector().GetAddresses()

	complexSvrList := complexConf.GetGlobal().GetServerConnector().GetAddresses()

	result := make(map[string]bool)

	for i := range complexSvrList {
		result[complexSvrList[i]] = true
	}

	for i := range easySvrList {
		if _, exist := result[easySvrList[i]]; !exist {
			result[easySvrList[i]] = true
		}
	}

	finalSvrList := make([]string, 0)
	for k := range result {
		finalSvrList = append(finalSvrList, k)
	}

	complexConf.GetGlobal().GetServerConnector().SetAddresses(finalSvrList)
}
