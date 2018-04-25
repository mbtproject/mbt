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

package intercept

import (
	"fmt"
	"reflect"
)

// Config is the configuration of an intercepted functions behavior.
type Config interface {
	// Return configures specific return value(s).
	Return(...interface{}) Config

	// Do configure a specific behavior for a function.
	Do(func(...interface{}) []interface{}) Config
}

type stdConfig struct {
	impl func(...interface{}) []interface{}
}

func (c *stdConfig) Return(values ...interface{}) Config {
	c.impl = func(...interface{}) []interface{} {
		return values
	}
	return c
}

func (c *stdConfig) Do(impl func(...interface{}) []interface{}) Config {
	c.impl = impl
	return c
}

// Interceptor provides the means for intercepting.
type Interceptor struct {
	config map[string]*stdConfig
	target interface{}
}

// NewInterceptor creates an interceptor for the specified target.
func NewInterceptor(target interface{}) *Interceptor {
	return &Interceptor{
		config: make(map[string]*stdConfig),
		target: target,
	}
}

// Config returns a configuration object for the specified function.
// Config can be used to manipulate the behavior of the target function.
func (i *Interceptor) Config(name string) Config {
	m := resolveMethod(name, i.target)
	if config, ok := i.config[name]; ok {
		return config
	}

	config := &stdConfig{
		impl: func(args ...interface{}) []interface{} {
			return invokeTarget(i.target, m, name, args...)
		},
	}

	i.config[name] = config
	return config
}

// Call invokes the configured target and returns the returns the results.
func (i *Interceptor) Call(name string, args ...interface{}) []interface{} {
	m := resolveMethod(name, i.target)
	if config, ok := i.config[name]; ok {
		return config.impl(args...)
	}

	return invokeTarget(i.target, m, name, args...)
}

func resolveMethod(name string, target interface{}) reflect.Value {
	tv := reflect.ValueOf(target)
	m := tv.MethodByName(name)
	if !m.IsValid() {
		panic(fmt.Errorf("method not found %s", name))
	}
	return m
}

func invokeTarget(target interface{}, method reflect.Value, name string, args ...interface{}) []interface{} {
	in := make([]reflect.Value, 0, len(args))
	for _, v := range args {
		in = append(in, reflect.ValueOf(v))
	}

	ov := method.Call(in)

	out := make([]interface{}, 0, len(ov))
	for _, v := range ov {
		out = append(out, v.Interface())
	}

	return out
}
