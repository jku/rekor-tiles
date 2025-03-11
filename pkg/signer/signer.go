/*
Copyright 2025 The Sigstore Authors.

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

// Copied from https://github.com/sigstore/rekor/blob/c820fcaf3afdc91f0acf6824d55c1ac7df249df1/pkg/signer/signer.go

package signer

import (
	"context"
	"crypto"
	"fmt"
	"strings"

	"github.com/sigstore/sigstore/pkg/signature"
	"github.com/sigstore/sigstore/pkg/signature/kms"
	"golang.org/x/exp/slices"

	// these are imported to load the providers via init() calls
	_ "github.com/sigstore/sigstore/pkg/signature/kms/aws"
	_ "github.com/sigstore/sigstore/pkg/signature/kms/azure"
	_ "github.com/sigstore/sigstore/pkg/signature/kms/gcp"
	_ "github.com/sigstore/sigstore/pkg/signature/kms/hashivault"
)

// New returns a Signer for the given KMS provider, Tink, or a private key file on disk.
func New(ctx context.Context, opts ...Option) (signature.Signer, error) {
	sc := &signerConfig{}
	for _, o := range opts {
		o(sc)
	}
	switch {
	case slices.ContainsFunc(kms.SupportedProviders(),
		func(s string) bool {
			return strings.HasPrefix(sc.kms, s)
		}):
		return kms.Get(ctx, sc.kms, sc.kmsHash)
	case sc.tinkKEKURI != "":
		return NewTinkSigner(ctx, sc.tinkKEKURI, sc.tinkKeysetPath)
	case sc.filePath != "":
		return NewFileSigner(sc.filePath, sc.password)
	default:
		return nil, fmt.Errorf("insufficient signing parameters provided, must configure one of file, KMS, or Tink signers")
	}
}

type signerConfig struct {
	filePath       string
	password       string
	kms            string
	kmsHash        crypto.Hash
	tinkKEKURI     string
	tinkKeysetPath string
}

type Option func(*signerConfig)

// WithFile configures a file-based signer with an optional password.
func WithFile(filePath, password string) Option {
	return func(sc *signerConfig) {
		sc.filePath = filePath
		sc.password = password
	}
}

// WithKMS configures a KMS signer.
func WithKMS(kms string, hash crypto.Hash) Option {
	return func(sc *signerConfig) {
		sc.kms = kms
		sc.kmsHash = hash
	}
}

// WithTink configures a Tink signer.
func WithTink(kekURI, keysetPath string) Option {
	return func(sc *signerConfig) {
		sc.tinkKEKURI = kekURI
		sc.tinkKeysetPath = keysetPath
	}
}
