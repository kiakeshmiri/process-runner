package mocks

import "os"

type WriteFileEntries struct {
	Path string
	Data []byte
	Perm os.FileMode
}

type WriteFileMock struct {
	Entries []*WriteFileEntries
	Error   error
}

func (w *WriteFileMock) WriteFile(path string, data []byte, perm os.FileMode) error {
	w.Entries = append(w.Entries, &WriteFileEntries{
		Path: path,
		Data: data,
		Perm: perm,
	})

	return w.Error
}
