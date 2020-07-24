package main

import (
	"context"
	"errors"
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/Masterminds/semver/v3"
	"github.com/artifacthub/hub/internal/hub"
	"github.com/artifacthub/hub/internal/img"
	"github.com/artifacthub/hub/internal/pkg"
	"github.com/artifacthub/hub/internal/repo"
	"github.com/artifacthub/hub/internal/tracker"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// Tracker is in charge of tracking the packages available in a OPA policies
// repository, registering and unregistering them as needed.
type Tracker struct {
	ctx    context.Context
	cfg    *viper.Viper
	r      *hub.Repository
	rc     hub.RepositoryCloner
	rm     hub.RepositoryManager
	pm     hub.PackageManager
	is     img.Store
	ec     tracker.ErrorsCollector
	Logger zerolog.Logger
}

// NewTracker creates a new Tracker instance.
func NewTracker(
	ctx context.Context,
	cfg *viper.Viper,
	r *hub.Repository,
	rm hub.RepositoryManager,
	pm hub.PackageManager,
	is img.Store,
	ec tracker.ErrorsCollector,
	opts ...func(t *Tracker),
) *Tracker {
	t := &Tracker{
		ctx:    ctx,
		cfg:    cfg,
		r:      r,
		rm:     rm,
		pm:     pm,
		is:     is,
		ec:     ec,
		Logger: log.With().Str("repo", r.Name).Logger(),
	}
	for _, o := range opts {
		o(t)
	}
	if t.rc == nil {
		t.rc = &repo.Cloner{}
	}
	return t
}

// Track registers or unregisters the OPA policies packages available as needed.
func (t *Tracker) Track(wg *sync.WaitGroup) error {
	defer wg.Done()

	// Clone repository
	t.Logger.Debug().Msg("cloning repository")
	tmpDir, packagesPath, err := t.rc.CloneRepository(t.ctx, t.r)
	if err != nil {
		return fmt.Errorf("error cloning repository: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Load packages already registered from this repository
	packagesRegistered, err := t.rm.GetPackagesDigest(t.ctx, t.r.RepositoryID)
	if err != nil {
		return fmt.Errorf("error getting registered packages: %w", err)
	}

	// Register available packages when needed
	bypassDigestCheck := t.cfg.GetBool("tracker.bypassDigestCheck")
	packagesAvailable := make(map[string]struct{})
	basePath := filepath.Join(tmpDir, packagesPath)
	packages, err := ioutil.ReadDir(basePath)
	if err != nil {
		return fmt.Errorf("error reading packages: %w", err)
	}
	for _, p := range packages {
		if !p.IsDir() {
			continue
		}
		pName := p.Name()
		pPath := filepath.Join(basePath, pName)

		// Get package versions available
		versionsUnfiltered, err := ioutil.ReadDir(pPath)
		if err != nil {
			t.Warn(fmt.Errorf("error reading package %s versions: %w", pName, err))
			continue
		}
		var versions []string
		for _, entryV := range versionsUnfiltered {
			if !entryV.IsDir() {
				continue
			}
			pVersion := entryV.Name()
			if _, err := semver.StrictNewVersion(pVersion); err != nil {
				t.Warn(fmt.Errorf("invalid package %s version (%s): %w", pName, pVersion, err))
				continue
			} else {
				versions = append(versions, pVersion)
			}
		}
		sort.Slice(versions, func(i, j int) bool {
			vi, _ := semver.NewVersion(versions[i])
			vj, _ := semver.NewVersion(versions[j])
			return vj.LessThan(vi)
		})

		// Process package versions available
		for i, pVersion := range versions {
			select {
			case <-t.ctx.Done():
				return nil
			default:
			}

			// Read and parse package version metadata
			pvPath := filepath.Join(pPath, pVersion)
			data, err := ioutil.ReadFile(filepath.Join(pvPath, "artifacthub.yaml"))
			if err != nil {
				t.Warn(fmt.Errorf("error reading package version metadata file %s: %w", pvPath, err))
				return nil
			}
			var md *hub.PackageMetadata
			if err = yaml.Unmarshal(data, &md); err != nil || md == nil {
				t.Warn(fmt.Errorf("error unmarshaling package version metadata file %s: %w", pvPath, err))
				return nil
			}
			pName := md.Name
			pVersion := md.Version

			// Check if this package version is already registered
			key := fmt.Sprintf("%s@%s", pName, pVersion)
			packagesAvailable[key] = struct{}{}
			if _, ok := packagesRegistered[key]; ok && !bypassDigestCheck {
				continue
			}

			// Register package version
			t.Logger.Debug().Str("name", pName).Str("v", pVersion).Msg("registering package version")
			var storeLogo bool
			if i == 0 {
				storeLogo = true
			}
			err = t.registerPackage(pvPath, md, storeLogo)
			if err != nil {
				t.Warn(fmt.Errorf("error registering package %s version %s: %w", pName, pVersion, err))
			}
		}
	}

	// Unregister packages not available anymore
	for key := range packagesRegistered {
		select {
		case <-t.ctx.Done():
			return nil
		default:
		}
		if _, ok := packagesAvailable[key]; !ok {
			p := strings.Split(key, "@")
			pName := p[0]
			pVersion := p[1]
			t.Logger.Debug().Str("name", pName).Str("v", pVersion).Msg("unregistering package")
			if err := t.unregisterPackage(pName, pVersion); err != nil {
				t.Warn(fmt.Errorf("error unregistering package %s version %s: %w", pName, pVersion, err))
			}
		}
	}

	return nil
}

// Warn is a helper that sends the error provided to the errors collector and
// logs it as a warning.
func (t *Tracker) Warn(err error) {
	t.ec.Append(t.r.RepositoryID, err)
	log.Warn().Err(err).Send()
}

// registerPackage registers a package version using the package metadata
// provided.
func (t *Tracker) registerPackage(pvPath string, md *hub.PackageMetadata, storeLogo bool) error {
	// Prepare package from metadata
	p, err := pkg.PreparePackageFromMetadata(md)
	if err != nil {
		return fmt.Errorf("error preparing package %s version %s from metadata: %w", md.Name, md.Version, err)
	}

	// Prepare source link
	var repoBaseURL, pkgsPath, provider string
	matches := repo.GitRepoURLRE.FindStringSubmatch(t.r.URL)
	if len(matches) >= 3 {
		repoBaseURL = matches[1]
		provider = matches[2]
	}
	if len(matches) == 4 {
		pkgsPath = strings.TrimSuffix(matches[3], "/")
	}
	var blobPath string
	switch provider {
	case "github":
		blobPath = "blob/master"
	case "gitlab":
		blobPath = "-/blob/master"
	}
	p.Links = append(p.Links, &hub.Link{
		Name: "source",
		URL:  fmt.Sprintf("%s/%s/%s%s", repoBaseURL, blobPath, pkgsPath, pvPath),
	})

	// Read policies file and add it to the package data field
	policies, err := ioutil.ReadFile(filepath.Join(pvPath, "policies.rego"))
	if err != nil {
		return fmt.Errorf("error reading package version policies file %s: %w", pvPath, err)
	}
	p.Data["policies"] = policies

	// Download logo image if needed and add its logo to the package
	if storeLogo && md.LogoURL != "" {
		data, err := img.Download(md.LogoURL)
		if err != nil {
			return fmt.Errorf("error downloading package %s version %s image: %w", md.Name, md.Version, err)
		}
		p.LogoImageID, err = t.is.SaveImage(t.ctx, data)
		if err != nil && !errors.Is(err, image.ErrFormat) {
			return fmt.Errorf("error saving package %s version %s image: %w", md.Name, md.Version, err)
		}
	}

	// Register package
	return t.pm.Register(t.ctx, p)
}

// unregisterPackage unregisters the package version provided.
func (t *Tracker) unregisterPackage(name, version string) error {
	p := &hub.Package{
		Name:       name,
		Version:    version,
		Repository: t.r,
	}
	return t.pm.Unregister(t.ctx, p)
}
