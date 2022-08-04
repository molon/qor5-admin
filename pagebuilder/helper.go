package pagebuilder

import (
	"path"
	"regexp"

	"github.com/goplaid/web"
	"gorm.io/gorm"
)

var (
	pathRe             = regexp.MustCompile(`^/[0-9a-zA-Z\/]*$`)
	slugWithCategoryRe = regexp.MustCompile(`^/[0-9a-zA-Z]*$`)
)

const (
	queryPathWithSlugSQL = `
SELECT pages.id, categories.path || pages.slug AS path_with_slug
FROM page_builder_pages pages
         LEFT JOIN page_builder_categories categories ON category_id = categories.id
WHERE pages.deleted_at IS NULL
`
	missingCategoryOrSlugMsg = "Category or Slug is required"
	invalidSlugMsg           = "Invalid Slug format"
	invalidPathMsg           = "Invalid Path format"
	conflictSlugMsg          = "Conflicting Slug"
	conflictPathMsg          = "Conflicting Path"
	existingPathMsg          = "Existing Path"

	unableDeleteCategoryMsg = "this category cannot be deleted because it has used with pages"
)

func pageValidator(p *Page, db *gorm.DB) (err web.ValidationErrors) {
	if p.CategoryID == 0 && p.Slug == "" {
		err.FieldError("Page.Category", missingCategoryOrSlugMsg)
		err.FieldError("Page.Slug", missingCategoryOrSlugMsg)
		return
	}

	if p.CategoryID == 0 {
		s := path.Clean(p.Slug)
		if s != "" && !pathRe.MatchString(s) {
			err.FieldError("Page.Slug", invalidSlugMsg)
			return
		}
	} else {
		if p.Slug != "" {
			s := path.Clean(p.Slug)
			if !slugWithCategoryRe.MatchString(s) {
				err.FieldError("Page.Slug", invalidSlugMsg)
				return
			}
		}
	}

	categories := []*Category{}
	if err := db.Model(&Category{}).Find(&categories).Error; err != nil {
		panic(err)
	}
	var c Category
	for _, e := range categories {
		if e.ID == p.CategoryID {
			c = *e
			break
		}
	}

	urlPath := c.Path + p.Slug

	type result struct {
		ID           uint
		PathWithSlug string
	}

	var results []result
	if err := db.Raw(queryPathWithSlugSQL).Scan(&results).Error; err != nil {
		panic(err)
	}

	for _, r := range results {
		if r.PathWithSlug == urlPath {
			err.FieldError("Page.Slug", conflictSlugMsg)
			return
		}
	}

	for _, e := range categories {
		if e.Path == urlPath {
			err.FieldError("Page.Slug", conflictSlugMsg)
			return
		}
	}

	return
}

func categoryValidator(category *Category, db *gorm.DB) (err web.ValidationErrors) {
	p := path.Clean(category.Path)
	if p != "" && !pathRe.MatchString(p) {
		err.FieldError("Category.Category", invalidPathMsg)
		return
	}

	// Verify category does not conflict the category with slug.
	type result struct {
		ID           uint
		PathWithSlug string
	}

	var results []result
	if err := db.Raw(queryPathWithSlugSQL).Scan(&results).Error; err != nil {
		panic(err)
	}

	for _, r := range results {
		if r.PathWithSlug == p {
			err.FieldError("Category.Category", conflictPathMsg)
			return
		}
	}

	// Verify category not duplicate.
	categories := []*Category{}
	if err := db.Model(&Category{}).Find(&categories).Error; err != nil {
		panic(err)
	}

	for _, c := range categories {
		if c.Path == p {
			err.FieldError("Category.Category", existingPathMsg)
			return
		}
	}

	return
}
