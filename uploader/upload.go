package uploader

type Driver interface {
	Upload() error
}
