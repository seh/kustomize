// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types

// Restrictions on what things can be referred to in a kustomization file.
//
//go:generate stringer -type=LoadRestrictions
type LoadRestrictions int

const (
	// LoadRestrictionsUnknown represents the lack of a specified restriction.
	LoadRestrictionsUnknown LoadRestrictions = iota

	// LoadRestrictionsRootOnly requires that files referenced by a kustomization file must be in or
	// under the directory holding the kustomization file itself.
	LoadRestrictionsRootOnly

	// LoadRestrictionsDominatedShallowly requires that regular files and symbolic links referenced
	// by a kustomization file must be in or under the directory holding the kustomization file
	// itself, but symbolic links located therein that may point outside of that directory tree.
	LoadRestrictionsDominatedShallowly

	// LoadRestrictionsNone allows the kustomization file to specify absolute or relative paths to
	// patch or resources files outside its own tree.
	LoadRestrictionsNone
)
