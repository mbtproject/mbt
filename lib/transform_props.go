/*
Copyright 2018 MBT Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package lib

import "github.com/mbtproject/mbt/e"

// transformProps is a helper function to convert all map[interface{}]interface{}
// to map[string]interface{}.
// Background:
// This operation is required due to what seems to be an anomaly in yaml
// library (https://github.com/go-yaml/yaml/issues/286)
// Due to this behavior, if we serialise the output of yaml.Unmarshal to
// with json.Marshal, it blows-up for scenarios with nested maps.
// This function is used to normalise the the whole tree before
// we use it.
func transformProps(p map[string]interface{}) (map[string]interface{}, error) {
	for k, v := range p {
		nv, err := transformIfRequired(v)
		if err != nil {
			return nil, err
		}
		p[k] = nv
	}
	return p, nil
}

func transformIfRequired(v interface{}) (interface{}, error) {
	if c, ok := v.(map[interface{}]interface{}); ok {
		return transformMaps(c)
	} else if c, ok := v.([]interface{}); ok {
		a := make([]interface{}, len(c))
		for i, v := range c {
			e, err := transformIfRequired(v)
			if err != nil {
				return nil, err
			}
			a[i] = e
		}
		return a, nil
	}

	return v, nil
}

func transformMaps(m map[interface{}]interface{}) (map[string]interface{}, error) {
	newMap := make(map[string]interface{})
	for k, v := range m {
		sk, ok := k.(string)
		if !ok {
			return nil, e.NewErrorf(ErrClassInternal, "Key is not a string %v", k)
		}

		nv, err := transformIfRequired(v)
		if err != nil {
			return nil, err
		}
		newMap[sk] = nv
	}
	return newMap, nil
}
