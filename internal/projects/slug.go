package projects

import (
	"path/filepath"
	"regexp"
	"strings"
)

var (
	nonAlnum     = regexp.MustCompile(`[^a-z0-9-]`)
	multiHyphen  = regexp.MustCompile(`-+`)
	domainRegexp = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)*\.test$`)
)

const maxSlugLen = 50

// GenerateSlug derives a URL-safe slug from a directory path.
func GenerateSlug(dirPath string) string {
	name := filepath.Base(dirPath)
	slug := strings.ToLower(name)
	slug = strings.NewReplacer(" ", "-", "_", "-").Replace(slug)
	slug = nonAlnum.ReplaceAllString(slug, "")
	slug = multiHyphen.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if len(slug) > maxSlugLen {
		slug = slug[:maxSlugLen]
	}
	return slug
}

// DomainFromSlug generates the default primary domain for a slug.
func DomainFromSlug(slug, tld string) string {
	return slug + "." + tld
}

// IsValidDomain returns true if domain matches the .test domain pattern.
func IsValidDomain(domain string) bool {
	return domainRegexp.MatchString(domain)
}
