package filepusher

type FilePusher interface {
	Push(content []byte, localPath, remotePath string) error
}
