// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package loader

import "errors"

var (
	ErrHTTP     = errors.New("HTTP Error")
	ErrRtNotDir = errors.New("must build at directory")
)
