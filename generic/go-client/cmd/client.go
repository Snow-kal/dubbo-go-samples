/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

import (
	"dubbo.apache.org/dubbo-go/v3"
	"dubbo.apache.org/dubbo-go/v3/client"
	"dubbo.apache.org/dubbo-go/v3/common/constant"
	_ "dubbo.apache.org/dubbo-go/v3/imports"

	hessian "github.com/apache/dubbo-go-hessian2"

	"github.com/dubbogo/gost/log/logger"
)

import (
	"github.com/apache/dubbo-go-samples/generic/go-client/pkg"
)

const (
	DirectServerURL        = "tri://127.0.0.1:50052"
	UserProvider           = "org.apache.dubbo.samples.UserProvider"
	ServiceVersion         = "1.0.0"
	ServiceGroup           = "triple"
	GenericProviderEnv     = "DUBBO_GO_GENERIC_PROVIDER"
	GenericProviderGo      = "go"
	GenericProviderJava    = "java"
	DefaultGenericProvider = GenericProviderGo
	ExpectedGetOneUserID   = "1000"
	ExpectedGetUserByID    = "A003"
)

func main() {
	hessian.RegisterPOJO(&pkg.User{})

	ins, err := dubbo.NewInstance(
		dubbo.WithName("generic-go-client"),
	)
	if err != nil {
		panic(err)
	}

	cli, err := ins.NewClient(
		client.WithClientProtocolTriple(),
		client.WithClientSerialization(constant.Hessian2Serialization),
	)
	if err != nil {
		panic(err)
	}

	genericService, err := cli.NewGenericService(
		UserProvider,
		client.WithURL(DirectServerURL),
		client.WithVersion(ServiceVersion),
		client.WithGroup(ServiceGroup),
		client.WithGeneric(),
		client.WithSerialization(constant.Hessian2Serialization),
	)
	if err != nil {
		panic(err)
	}

	logger.Infof("Direct URL: %s", DirectServerURL)
	logger.Info("Connected to server via direct URL, starting checks...")

	provider := os.Getenv(GenericProviderEnv)
	if provider == "" {
		provider = DefaultGenericProvider
	}
	logger.Infof("Generic provider phase: %s", provider)

	failed := false
	failed = runGenericChecks(genericService.Invoke) || failed
	failed = runGenericModeChecks(cli, provider) || failed

	if failed {
		logger.Errorf("Some generic call checks failed")
		os.Exit(1)
	}
	logger.Info("All generic call checks passed")
}

type genericInvokeFunc func(context.Context, string, []string, []hessian.Object) (any, error)

// beanUserDTO contains the fields that the Bean generalizer can round-trip reliably.
// time.Time is excluded because its unexported state is not represented by a JavaBean descriptor.
type beanUserDTO struct {
	ID   string
	Name string
	Age  int32
}

func runGenericChecks(invoke genericInvokeFunc) bool {
	failed := false
	ctx := context.Background()

	// GetUser1(String)
	result, err := invoke(ctx, "GetUser1", []string{"java.lang.String"}, []hessian.Object{"A003"})
	if err != nil {
		logger.Errorf("GetUser1 failed: %v", err)
		failed = true
	} else {
		logger.Infof("GetUser1(userId string) res: %+v", result)
	}

	// GetUser2(String, String)
	result, err = invoke(ctx, "GetUser2", []string{"java.lang.String", "java.lang.String"}, []hessian.Object{"A003", "lily"})
	if err != nil {
		logger.Errorf("GetUser2 failed: %v", err)
		failed = true
	} else {
		logger.Infof("GetUser2(userId string, name string) res: %+v", result)
	}

	// GetUser3(int)
	result, err = invoke(ctx, "GetUser3", []string{"int"}, []hessian.Object{int32(1)})
	if err != nil {
		logger.Errorf("GetUser3 failed: %v", err)
		failed = true
	} else {
		logger.Infof("GetUser3(userCode int) res: %+v", result)
	}

	// GetUser4(int, String)
	result, err = invoke(ctx, "GetUser4", []string{"int", "java.lang.String"}, []hessian.Object{int32(1), "zhangsan"})
	if err != nil {
		logger.Errorf("GetUser4 failed: %v", err)
		failed = true
	} else {
		logger.Infof("GetUser4(userCode int, name string) res: %+v", result)
	}

	// GetOneUser()
	result, err = invoke(ctx, "GetOneUser", []string{}, []hessian.Object{})
	if err != nil {
		logger.Errorf("GetOneUser failed: %v", err)
		failed = true
	} else {
		logger.Infof("GetOneUser() res: %+v", result)
	}

	// GetUsers(String[])
	result, err = invoke(ctx, "GetUsers", []string{"[Ljava.lang.String;"}, []hessian.Object{[]string{"001", "002", "003"}})
	if err != nil {
		logger.Errorf("GetUsers failed: %v", err)
		failed = true
	} else {
		logger.Infof("GetUsers(userIdList []string) res: %+v", result)
	}

	// GetUsersMap(String[])
	result, err = invoke(ctx, "GetUsersMap", []string{"[Ljava.lang.String;"}, []hessian.Object{[]string{"001", "002"}})
	if err != nil {
		logger.Errorf("GetUsersMap failed: %v", err)
		failed = true
	} else {
		logger.Infof("GetUsersMap(userIdList []string) res: %+v", result)
	}

	// QueryAll()
	result, err = invoke(ctx, "QueryAll", []string{}, []hessian.Object{})
	if err != nil {
		logger.Errorf("QueryAll failed: %v", err)
		failed = true
	} else {
		logger.Infof("QueryAll() res: %+v", result)
	}

	// QueryUser(User)
	testUser := &pkg.User{
		ID:   "3213",
		Name: "panty",
		Age:  25,
		Time: time.Now(),
	}
	result, err = invoke(ctx, "QueryUser", []string{"org.apache.dubbo.samples.User"}, []hessian.Object{testUser})
	if err != nil {
		logger.Errorf("QueryUser failed: %v", err)
		failed = true
	} else {
		logger.Infof("QueryUser(user *User) res: %+v", result)
	}

	// QueryUsers(User[])
	testUsers := []*pkg.User{
		{ID: "3212", Name: "XavierNiu", Age: 24, Time: time.Now()},
		{ID: "3213", Name: "zhangsan", Age: 21, Time: time.Now()},
	}
	result, err = invoke(ctx, "QueryUsers", []string{"[Lorg.apache.dubbo.samples.User;"}, []hessian.Object{testUsers})
	if err != nil {
		logger.Errorf("QueryUsers failed: %v", err)
		failed = true
	} else {
		logger.Infof("QueryUsers(users []*User) res: %+v", result)
	}

	return failed
}

func runGenericModeChecks(cli *client.Client, provider string) bool {
	failed := false
	ctx := context.Background()

	if _, err := cli.NewGenericService(
		UserProvider,
		client.WithURL(DirectServerURL),
		client.WithVersion(ServiceVersion),
		client.WithGroup(ServiceGroup),
		client.WithGenericType("bad-type"),
		client.WithSerialization(constant.Hessian2Serialization),
	); err == nil {
		logger.Error("NewGenericService accepted an unknown generic mode")
		failed = true
	} else {
		logger.Infof("NewGenericService rejected unknown generic mode: %v", err)
	}

	testCases := []struct {
		name       string
		mode       string
		method     string
		types      []string
		args       []hessian.Object
		typed      bool
		expectedID string
	}{
		{
			name:       "true",
			mode:       constant.GenericSerializationDefault,
			method:     "GetUser1",
			types:      []string{"java.lang.String"},
			args:       []hessian.Object{"A003"},
			typed:      true,
			expectedID: ExpectedGetUserByID,
		},
		{
			name:   "gson",
			mode:   constant.GenericSerializationGson,
			method: "GetOneUser",
			types:  []string{},
			args:   []hessian.Object{},
		},
		{
			name:       "bean",
			mode:       constant.GenericSerializationBean,
			method:     "GetOneUser",
			types:      []string{},
			args:       []hessian.Object{},
			typed:      true,
			expectedID: ExpectedGetOneUserID,
		},
	}

	for _, testCase := range testCases {
		service, err := cli.NewGenericService(
			UserProvider,
			client.WithURL(DirectServerURL),
			client.WithVersion(ServiceVersion),
			client.WithGroup(ServiceGroup),
			client.WithGenericType(testCase.mode),
			client.WithSerialization(constant.Hessian2Serialization),
		)
		if err != nil {
			logger.Errorf("create generic service (%s) failed: %v", testCase.name, err)
			failed = true
			continue
		}
		if !testCase.typed {
			result, invokeErr := service.Invoke(ctx, testCase.method, testCase.types, testCase.args)
			if invokeErr != nil {
				logger.Errorf("%s generic result (%s) failed: %v", testCase.method, testCase.name, invokeErr)
				failed = true
				continue
			}
			validationErr := checkGsonResult(provider, result, ExpectedGetOneUserID)
			if validationErr != nil {
				logger.Errorf("%s generic result (%s) failed validation: %v", testCase.method, testCase.name, validationErr)
				failed = true
				continue
			}
			continue
		}
		if testCase.mode == constant.GenericSerializationBean {
			var user beanUserDTO
			err = service.InvokeWithType(
				ctx,
				testCase.method,
				testCase.types,
				testCase.args,
				&user,
			)
			if err != nil {
				logger.Errorf("%s typed result (%s) failed: %v", testCase.method, testCase.name, err)
				failed = true
				continue
			}
			if user.ID != testCase.expectedID || user.Name == "" || user.Age == 0 {
				logger.Errorf("%s typed result (%s) returned incomplete DTO: %+v", testCase.method, testCase.name, user)
				failed = true
				continue
			}
			logger.Infof("%s typed result (%s) DTO res: %+v", testCase.method, testCase.name, user)
			continue
		}

		var user pkg.User
		err = service.InvokeWithType(
			ctx,
			testCase.method,
			testCase.types,
			testCase.args,
			&user,
		)
		if err != nil {
			logger.Errorf("%s typed result (%s) failed: %v", testCase.method, testCase.name, err)
			failed = true
			continue
		}
		if user.ID != testCase.expectedID || user.Name == "" || user.Age == 0 || user.Time.IsZero() {
			logger.Errorf("%s typed result (%s) returned incomplete user: %+v", testCase.method, testCase.name, user)
			failed = true
			continue
		}
		logger.Infof("%s typed result (%s) res: %+v", testCase.method, testCase.name, user)
	}

	return failed
}

func checkGsonResult(provider string, result any, expectedID string) error {
	if provider == GenericProviderJava {
		switch result.(type) {
		case map[any]any, map[string]any:
			logger.Warnf("Java provider does not support gson result encoding; observed expected Map fallback type=%T", result)
			return nil
		}
	}

	jsonResult, ok := result.(string)
	if !ok {
		return fmt.Errorf("provider %q returned %T, want JSON string", provider, result)
	}

	var user pkg.User
	if err := json.Unmarshal([]byte(jsonResult), &user); err != nil {
		return fmt.Errorf("decode JSON result: %w", err)
	}
	if user.ID != expectedID || user.Name == "" || user.Age == 0 || user.Time.IsZero() {
		return fmt.Errorf("incomplete JSON user: %+v", user)
	}

	logger.Infof("GetOneUser gson JSON result (%s provider) res: %+v", provider, user)
	return nil
}
