// Copyright © 2025 tjj
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package cast

import (
	"fmt"
	"reflect"
)

var (
	stringType   = typeFor[string]()
	stringerType = typeFor[fmt.Stringer]()
	byteType     = typeFor[byte]()
	anyType      = typeFor[any]()
	errType      = typeFor[error]()
	nilErrValue  = reflect.Zero(errType)
)

const zerosSize = 1024
