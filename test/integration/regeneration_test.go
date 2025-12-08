// SPDX-FileCopyrightText: 2025 Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
//
// SPDX-License-Identifier: MIT

package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type RegenerationTestSuite struct {
	suite.Suite
	tempDir       string
	fabricaBinary string
}

func (s *RegenerationTestSuite) SetupSuite() {
	// Find fabrica binary
	wd, err := os.Getwd()
	s.Require().NoError(err)

	// Assuming test is run from test/integration
	projectRoot := filepath.Join(wd, "..", "..")
	s.fabricaBinary = filepath.Join(projectRoot, "bin", "fabrica")

	// Verify binary exists
	_, err = os.Stat(s.fabricaBinary)
	s.Require().NoError(err, "fabrica binary not found at %s. Run 'make build' first.", s.fabricaBinary)
}

func (s *RegenerationTestSuite) SetupTest() {
	var err error
	s.tempDir, err = os.MkdirTemp("", "fabrica-regeneration-test-*")
	s.Require().NoError(err)
}

func (s *RegenerationTestSuite) TearDownTest() {
	os.RemoveAll(s.tempDir)
}

func (s *RegenerationTestSuite) TestRegenerationAddsNewResources() {
	projectName := "regeneration-test"
	project := NewTestProject(&s.Suite, s.tempDir, projectName, "github.com/example/regeneration", "file")

	// 1. Initialize project
	err := project.Initialize(s.fabricaBinary)
	s.Require().NoError(err)

	// 2. Add first resource
	err = project.AddResource(s.fabricaBinary, "DeviceProfile")
	s.Require().NoError(err)

	// 3. Generate
	err = project.Generate(s.fabricaBinary)
	s.Require().NoError(err)

	// 4. Verify first resource routes exist
	routesFile := filepath.Join(project.Dir, "cmd", "server", "routes_generated.go")
	content, err := os.ReadFile(routesFile)
	s.Require().NoError(err)
	s.Contains(string(content), "DeviceProfile", "Routes should contain DeviceProfile")

	// 5. Add second resource
	err = project.AddResource(s.fabricaBinary, "UpdateProfile")
	s.Require().NoError(err)

	// 6. Generate again
	err = project.Generate(s.fabricaBinary)
	s.Require().NoError(err)

	// 7. Verify both resources exist in routes
	content, err = os.ReadFile(routesFile)
	s.Require().NoError(err)
	s.Contains(string(content), "DeviceProfile", "Routes should still contain DeviceProfile")
	s.Contains(string(content), "UpdateProfile", "Routes should now contain UpdateProfile")

	// 8. Verify registration file contains both resources
	regFile := filepath.Join(project.Dir, "pkg", "resources", "register_generated.go")
	regContent, err := os.ReadFile(regFile)
	s.Require().NoError(err)
	s.Contains(string(regContent), "DeviceProfile", "Registration should contain DeviceProfile")
	s.Contains(string(regContent), "UpdateProfile", "Registration should contain UpdateProfile")

	// 9. Verify other global files contain the new resource
	filesToCheck := []string{
		filepath.Join("cmd", "server", "models_generated.go"),
		filepath.Join("internal", "storage", "storage_generated.go"),
		filepath.Join("pkg", "client", "client_generated.go"),
		filepath.Join("cmd", "server", "openapi_generated.go"),
	}

	for _, relPath := range filesToCheck {
		fullPath := filepath.Join(project.Dir, relPath)
		content, err := os.ReadFile(fullPath)
		s.Require().NoError(err, "Failed to read %s", relPath)
		s.Contains(string(content), "UpdateProfile", "%s should contain UpdateProfile", relPath)
	}

	// 10. Verify new handler file exists
	handlerFile := filepath.Join(project.Dir, "cmd", "server", "updateprofile_handlers_generated.go")
	_, err = os.Stat(handlerFile)
	s.Require().NoError(err, "Handler file for UpdateProfile should exist")
}

func TestRegenerationSuite(t *testing.T) {
	suite.Run(t, new(RegenerationTestSuite))
}
