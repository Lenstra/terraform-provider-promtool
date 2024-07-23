// Lenstra - This file contains methods used by the promtool package. - https://github.com/prometheus/prometheus/blob/5a6c8f9c152dfab5f96f5b6f14703b801b014255/cmd/promtool/main.go

// Copyright 2015 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package promtool

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-kit/log"
	config_util "github.com/prometheus/common/config"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/discovery/file"
	"github.com/prometheus/prometheus/discovery/kubernetes"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/notifier"
	"github.com/prometheus/prometheus/scrape"
	"gopkg.in/yaml.v2"
)

func CheckConfig(content string, checkSyntaxOnly bool) ([]string, error) {
	cfg, err := config.Load(content, false, log.NewNopLogger())
	if err != nil {
		return nil, err
	}

	var ruleFiles []string
	if !checkSyntaxOnly {
		for _, rf := range cfg.RuleFiles {
			rfs, err := filepath.Glob(rf)
			if err != nil {
				return nil, err
			}
			// If an explicit file was given, error if it is not accessible.
			if !strings.Contains(rf, "*") {
				if len(rfs) == 0 {
					return nil, fmt.Errorf("%q does not point to an existing file", rf)
				}
				if err := checkFileExists(rfs[0]); err != nil {
					return nil, fmt.Errorf("error checking rule file %q: %w", rfs[0], err)
				}
			}
			ruleFiles = append(ruleFiles, rfs...)
		}
	}

	var scfgs []*config.ScrapeConfig
	if checkSyntaxOnly {
		scfgs = cfg.ScrapeConfigs
	} else {
		var err error
		scfgs, err = cfg.GetScrapeConfigs()
		if err != nil {
			return nil, fmt.Errorf("error loading scrape configs: %w", err)
		}
	}

	for _, scfg := range scfgs {
		if !checkSyntaxOnly && scfg.HTTPClientConfig.Authorization != nil {
			if err := checkFileExists(scfg.HTTPClientConfig.Authorization.CredentialsFile); err != nil {
				return nil, fmt.Errorf("error checking authorization credentials or bearer token file %q: %w", scfg.HTTPClientConfig.Authorization.CredentialsFile, err)
			}
		}

		if err := checkTLSConfig(scfg.HTTPClientConfig.TLSConfig, checkSyntaxOnly); err != nil {
			return nil, err
		}

		for _, c := range scfg.ServiceDiscoveryConfigs {
			switch c := c.(type) {
			case *kubernetes.SDConfig:
				if err := checkTLSConfig(c.HTTPClientConfig.TLSConfig, checkSyntaxOnly); err != nil {
					return nil, err
				}
			case *file.SDConfig:
				if checkSyntaxOnly {
					break
				}
				for _, file := range c.Files {
					files, err := filepath.Glob(file)
					if err != nil {
						return nil, err
					}
					if len(files) != 0 {
						for _, f := range files {
							var targetGroups []*targetgroup.Group
							targetGroups, err = checkSDFile(f)
							if err != nil {
								return nil, fmt.Errorf("checking SD file %q: %w", file, err)
							}
							if err := checkTargetGroupsForScrapeConfig(targetGroups, scfg); err != nil {
								return nil, err
							}
						}
						continue
					}
					fmt.Printf("  WARNING: file %q for file_sd in scrape job %q does not exist\n", file, scfg.JobName)
				}
			case discovery.StaticConfig:
				if err := checkTargetGroupsForScrapeConfig(c, scfg); err != nil {
					return nil, err
				}
			}
		}
	}

	alertConfig := cfg.AlertingConfig
	for _, amcfg := range alertConfig.AlertmanagerConfigs {
		for _, c := range amcfg.ServiceDiscoveryConfigs {
			switch c := c.(type) {
			case *file.SDConfig:
				if checkSyntaxOnly {
					break
				}
				for _, file := range c.Files {
					files, err := filepath.Glob(file)
					if err != nil {
						return nil, err
					}
					if len(files) != 0 {
						for _, f := range files {
							var targetGroups []*targetgroup.Group
							targetGroups, err = checkSDFile(f)
							if err != nil {
								return nil, fmt.Errorf("checking SD file %q: %w", file, err)
							}

							if err := checkTargetGroupsForAlertmanager(targetGroups, amcfg); err != nil {
								return nil, err
							}
						}
						continue
					}
					fmt.Printf("  WARNING: file %q for file_sd in alertmanager config does not exist\n", file)
				}
			case discovery.StaticConfig:
				if err := checkTargetGroupsForAlertmanager(c, amcfg); err != nil {
					return nil, err
				}
			}
		}
	}
	return ruleFiles, nil
}

func checkTargetGroupsForScrapeConfig(targetGroups []*targetgroup.Group, scfg *config.ScrapeConfig) error {
	var targets []*scrape.Target
	lb := labels.NewBuilder(labels.EmptyLabels())
	for _, tg := range targetGroups {
		var failures []error
		targets, failures = scrape.TargetsFromGroup(tg, scfg, false, targets, lb)
		if len(failures) > 0 {
			first := failures[0]
			return first
		}
	}

	return nil
}

func checkTargetGroupsForAlertmanager(targetGroups []*targetgroup.Group, amcfg *config.AlertmanagerConfig) error {
	for _, tg := range targetGroups {
		if _, _, err := notifier.AlertmanagerFromGroup(tg, amcfg); err != nil {
			return err
		}
	}

	return nil
}

func checkSDFile(filename string) ([]*targetgroup.Group, error) {
	fd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	content, err := io.ReadAll(fd)
	if err != nil {
		return nil, err
	}

	var targetGroups []*targetgroup.Group

	switch ext := filepath.Ext(filename); strings.ToLower(ext) {
	case ".json":
		if err := json.Unmarshal(content, &targetGroups); err != nil {
			return nil, err
		}
	case ".yml", ".yaml":
		if err := yaml.UnmarshalStrict(content, &targetGroups); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid file extension: %q", ext)
	}

	for i, tg := range targetGroups {
		if tg == nil {
			return nil, fmt.Errorf("nil target group item found (index %d)", i)
		}
	}

	return targetGroups, nil
}

func checkTLSConfig(tlsConfig config_util.TLSConfig, checkSyntaxOnly bool) error {
	if len(tlsConfig.CertFile) > 0 && len(tlsConfig.KeyFile) == 0 {
		return fmt.Errorf("client cert file %q specified without client key file", tlsConfig.CertFile)
	}
	if len(tlsConfig.KeyFile) > 0 && len(tlsConfig.CertFile) == 0 {
		return fmt.Errorf("client key file %q specified without client cert file", tlsConfig.KeyFile)
	}

	if checkSyntaxOnly {
		return nil
	}

	if err := checkFileExists(tlsConfig.CertFile); err != nil {
		return fmt.Errorf("error checking client cert file %q: %w", tlsConfig.CertFile, err)
	}
	if err := checkFileExists(tlsConfig.KeyFile); err != nil {
		return fmt.Errorf("error checking client key file %q: %w", tlsConfig.KeyFile, err)
	}

	return nil
}

func checkFileExists(fn string) error {
	// Nothing set, nothing to error on.
	if fn == "" {
		return nil
	}
	_, err := os.Stat(fn)
	return err
}
