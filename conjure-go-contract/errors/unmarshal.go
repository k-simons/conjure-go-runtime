// Copyright (c) 2020 Palantir Technologies. All rights reserved.
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

package errors

import (
	"reflect"

	"github.com/palantir/conjure-go-runtime/conjure-go-contract/codecs"
	werror "github.com/palantir/witchcraft-go-error"
)

type errorNameAccessor struct {
	Name string `json:"errorName"`
}

// UnmarshalError attempts to deserialize the message to a known implementation of Error.
// Custom error types should be registered using RegisterErrorType.
// If the ErrorName is not recognized, a genericError is returned with all params marked unsafe.
// If we fail to unmarshal to a generic SerializableError or to the type specified by ErrorName, an error is returned.
func UnmarshalError(body []byte) (Error, error) {
	var name errorNameAccessor // TODO(bmoylan) is gson's speed worth the dependency?
	if err := codecs.JSON.Unmarshal(body, &name); err != nil {
		return nil, werror.Wrap(err, "failed to unmarshal body as conjure error")
	}
	typ, ok := registry[name.Name]
	if !ok {
		// Unrecognized error name, fall back to genericError
		typ = reflect.TypeOf(genericError{})
	}

	instance := reflect.New(typ).Interface()
	if err := codecs.JSON.Unmarshal(body, &instance); err != nil {
		// TODO(bmoylan): Do we want to be more lenient and use a genericError if this can not unmarshal?
		return nil, werror.Wrap(err, "failed to unmarshal body using registered type", werror.SafeParam("type", typ.String()))
	}

	// Cast should never panic, as we've verified in RegisterErrorType
	return instance.(Error), nil
}