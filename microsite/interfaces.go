package microsite

import "github.com/qor/oss"

type MicroSiteInterface interface {
	GetId() uint
	//GetVersionName() string
	//SetVersionPriority(string)
	GetStatus() string

	GetPackagePath(fileName string) string
	GetPreviewPrePath() string
	GetPreviewUrl(domain, fileName string) string
	GetPublishedPath(fileName string) string
	GetPublishedUrl(domain, fileName string) string
	GetFileList() (arr []string)
	SetFilesList(filesList []string)
	GetPackage() FileSystem
	SetPackage(fileName, url string)
	GetFilesListAndPublishPreviewFiles(fileName, filePath string, storage oss.StorageInterface) (filesList []string, err error)
	PublishArchiveFiles(fileName, filePath string, storage oss.StorageInterface) (err error)
}
