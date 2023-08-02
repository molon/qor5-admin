package pagebuilder

import (
	"path"
	"regexp"

	"github.com/qor5/admin/l10n"
	"github.com/qor5/web"
	"gorm.io/gorm"
)

var (
	directoryRe = regexp.MustCompile(`^([\/]{1}[a-z0-9.]+)+(\/?){1}$|^([\/]{1})$`)
	slugRe      = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
)

const (
	queryLocaleCodeCategoryPathSlugSQL = `
	SELECT pages.id AS id,
	       pages.version AS version,
	       pages.locale_code AS locale_code,
	       categories.path AS category_path,
	       pages.slug AS slug
FROM page_builder_pages pages
LEFT JOIN page_builder_categories categories ON category_id = categories.id AND pages.locale_code = categories.locale_code
WHERE pages.deleted_at IS NULL AND categories.deleted_at IS NULL
`
	missingCategoryOrSlugMsg = "Category or Slug is required"
	invalidPathMsg           = "Invalid Path"
	invalidSlugMsg           = "Invalid Slug"
	conflictSlugMsg          = "Conflicting Slug"
	conflictPathMsg          = "Conflicting Path"
	existingPathMsg          = "Existing Path"

	unableDeleteCategoryMsg = "this category cannot be deleted because it has used with pages"
)

type pagePathInfo struct {
	ID           uint
	Version      string
	LocaleCode   string
	CategoryPath string
	Slug         string
}

func pageValidator(p *Page, db *gorm.DB, l10nB *l10n.Builder) (err web.ValidationErrors) {
	if p.CategoryID == 0 && p.Slug == "" {
		// Page category can be empty when slug is not empty.
		err.FieldError("Page.Category", missingCategoryOrSlugMsg)
		err.FieldError("Page.Slug", missingCategoryOrSlugMsg)
		return
	}

	if p.Slug != "" {
		if !slugRe.MatchString(p.Slug) {
			err.FieldError("Page.Slug", invalidSlugMsg)
			return
		}
	}

	categories := []*Category{}
	if err := db.Model(&Category{}).Find(&categories).Error; err != nil {
		panic(err)
	}
	var currentPageCategory Category
	for _, category := range categories {
		if category.ID == p.CategoryID && category.LocaleCode == p.LocaleCode {
			currentPageCategory = *category
			break
		}
	}

	var localePath string
	if l10nB != nil {
		localePath = l10nB.GetLocalePath(p.LocaleCode)
	}

	currentPagePublishUrl := p.getPublishUrl(localePath, currentPageCategory.Path)

	{
		// Verify page publish URL does not conflict the other pages' PublishUrl.
		var pagePathInfos []pagePathInfo
		if err := db.Raw(queryLocaleCodeCategoryPathSlugSQL).Scan(&pagePathInfos).Error; err != nil {
			panic(err)
		}

		for _, info := range pagePathInfos {
			if info.ID == p.ID && info.LocaleCode == p.LocaleCode {
				continue
			}
			var localePath string
			if l10nB != nil {
				localePath = l10nB.GetLocalePath(info.LocaleCode)
			}

			if generatePublishUrl(localePath, info.CategoryPath, info.Slug) == currentPagePublishUrl {
				err.FieldError("Page.Slug", conflictSlugMsg)
				return
			}
		}
	}

	if p.Slug != "" {
		var allLocalePaths []string
		if l10nB != nil {
			allLocalePaths = l10nB.GetAllLocalePaths()
		} else {
			allLocalePaths = []string{""}
		}
		for _, category := range categories {
			for _, localePath := range allLocalePaths {
				if generatePublishUrl(localePath, category.Path, "") == currentPagePublishUrl {
					err.FieldError("Page.Slug", conflictSlugMsg)
					return
				}
			}
		}
	}

	return
}

func categoryValidator(category *Category, db *gorm.DB, l10nB *l10n.Builder) (err web.ValidationErrors) {
	categoryPath := path.Clean(category.Path)
	categoryPath = path.Join("/", categoryPath)
	if categoryPath == "/" || !directoryRe.MatchString(categoryPath) {
		err.FieldError("Category.Category", invalidPathMsg)
		return
	}

	var localePath string
	if l10nB != nil {
		localePath = l10nB.GetLocalePath(category.LocaleCode)
	}

	var currentCategoryPathPublishUrl = generatePublishUrl(localePath, categoryPath, "")

	{
		// Verify category does not conflict the pages' PublishUrl.
		var pagePathInfos []pagePathInfo
		if err := db.Raw(queryLocaleCodeCategoryPathSlugSQL).Scan(&pagePathInfos).Error; err != nil {
			panic(err)
		}

		for _, info := range pagePathInfos {
			var pageLocalePath string
			if l10nB != nil {
				pageLocalePath = l10nB.GetLocalePath(info.LocaleCode)
			}
			if generatePublishUrl(pageLocalePath, info.CategoryPath, info.Slug) == currentCategoryPathPublishUrl {
				err.FieldError("Category.Category", conflictPathMsg)
				return
			}
		}
	}

	{
		// Verify category not duplicate.
		categories := []*Category{}
		if err := db.Model(&Category{}).Find(&categories).Error; err != nil {
			panic(err)
		}

		for _, c := range categories {
			if c.ID == category.ID && c.LocaleCode == category.LocaleCode {
				continue
			}
			var localePath string
			if l10nB != nil {
				localePath = l10nB.GetLocalePath(c.LocaleCode)
			}

			if generatePublishUrl(localePath, c.Path, "") == currentCategoryPathPublishUrl {
				err.FieldError("Category.Category", existingPathMsg)
				return
			}
		}
	}

	return
}
