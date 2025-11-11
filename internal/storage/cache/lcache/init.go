// Copyright © 2024 OpenIM. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lcache

import (
	"strings"
	"sync"

	"github.com/ZampoRen/go-server-comon/internal/common/config"
	"github.com/ZampoRen/go-server-comon/internal/common/storage/cache/cachekey"
)

var (
	once      sync.Once
	subscribe map[string][]string
)

func InitLocalCache(localCache *config.LocalCache) {
	once.Do(func() {
		list := []struct {
			Local config.CacheConfig
			Keys  []string
		}{
			{
				Local: localCache.User,
				Keys:  []string{cachekey.UserInfoKey},
			},
		}
		subscribe = make(map[string][]string)
		for _, v := range list {
			if v.Local.Enable() {
				subscribe[v.Local.Topic] = v.Keys
			}
		}
	})
}

func GetPublishKeysByTopic(topics []string, keys []string) map[string][]string {
	keysByTopic := make(map[string][]string)
	for _, topic := range topics {
		keysByTopic[topic] = []string{}
	}

	for _, key := range keys {
		for _, topic := range topics {
			prefixes, ok := subscribe[topic]
			if !ok {
				continue
			}
			for _, prefix := range prefixes {
				if strings.HasPrefix(key, prefix) {
					keysByTopic[topic] = append(keysByTopic[topic], key)
					break
				}
			}
		}
	}

	return keysByTopic
}
