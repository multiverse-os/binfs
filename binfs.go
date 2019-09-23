package binfs

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	//memfd "github.com/multiverse-os/binfs/memfd"
)

// TODO: Build a BKeyValue store as an alternate package but also one that could
// work with this one

const (
	HeadersMagicSequence = "BHS"
	StorageMagicSequence = "BFS"
	HeaderSize           = uint64(64)
)

type Executable struct {
	Path          string
	Filename      string
	Data          []byte
	HeadersOffset uint64
	StorageOffset uint64
	StorageSize   uint64
	StoredFiles   []*StoredFile
	Storage       map[string]*File
}

type File struct {
	Data []byte
}

type StoredFile struct {
	Filename string
	Size     uint64
	Offset   uint64
	Checksum []byte
}

func MarshalHeader(bytes []byte) *StoredFile {
	return &StoredFile{
		Filename: string(bytes[:16]),
		Size:     binary.LittleEndian.Uint64(bytes[16:24]),
		Offset:   binary.LittleEndian.Uint64(bytes[24:32]),
		Checksum: bytes[32:],
	}
}

func (self *StoredFile) End() uint64 { return (self.Offset + self.Size) }

func (self *StoredFile) UnmarshalHeader() (outBytes []byte) {
	outBytes = PadRight([]byte(self.Filename), 16)
	sizeBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(sizeBytes, self.Size)

	outBytes = append(outBytes, sizeBytes...)
	offsetBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(offsetBytes, self.Offset)

	outBytes = append(outBytes, offsetBytes...)
	outBytes = append(outBytes, self.Checksum[:]...)

	return outBytes
}

func (self *StoredFile) ValidChecksum(checksum []byte) bool {
	return (bytes.Compare(self.Checksum, checksum) == 0)
}

func (self *Executable) UpdateOffsets() {
	index := bytes.Index(self.Data, []byte(HeadersMagicSequence))
	if index == -1 {
		self.Data = append(self.Data, []byte(HeadersMagicSequence)...)
		self.HeadersOffset = uint64(len(self.Data))
		self.Data = append(self.Data, []byte(StorageMagicSequence)...)
		self.StorageOffset = uint64(len(self.Data))
	} else {
		self.HeadersOffset = uint64(index + 3)
		index := bytes.Index(self.Data, []byte(StorageMagicSequence))
		if index != -1 {
			self.StorageOffset = uint64(index + 3)
		}
		self.load()
	}
}

func Load() *Executable {
	path, _ := os.Executable()
	filename := filepath.Base(path)
	binaryData, _ := ioutil.ReadFile(path)

	executable := &Executable{
		Data:     binaryData,
		Filename: filename,
		Path:     path,
		Storage:  make(map[string]*File),
	}

	executable.UpdateOffsets()
	return executable
}

func (self *Executable) LoadFromBinary(storedFile *StoredFile) {
	fileData := self.Data[(storedFile.Offset):(storedFile.Offset + storedFile.Size)]
	checksum := sha256.Sum256(fileData)

	if storedFile.ValidChecksum(checksum[:]) {
		self.Storage[storedFile.Filename] = &File{
			Data: fileData,
		}
	}
}

//func (self *Executable) Headers() uint16 {}

func (self *Executable) Size() uint64 { return uint64(len(self.Data)) }
func (self *Executable) HeadersData() []byte {
	return self.Data[self.HeadersOffset:self.StorageOffset]
}
func (self *Executable) HeaderCount() int {
	return len(self.Data[self.HeadersOffset:self.StorageOffset]) / 64
}
func (self *Executable) StorageData() []byte { return self.Data[self.StorageOffset:] }

func (self *Executable) HeaderData(index int) []byte {
	return self.Data[(self.HeadersOffset + (HeaderSize * uint64(index))):(self.HeadersOffset + HeaderSize + (HeaderSize * uint64(index)))]
}

func (self *Executable) Exists(checksum []byte) bool {
	for _, file := range self.StoredFiles {
		if file.ValidChecksum(checksum) {
			return true
		}
	}
	return false
}

func (self *Executable) FilenameExists(name string) bool {
	for _, file := range self.StoredFiles {
		if file.Filename == name {
			return true
		}
	}
	for filename, _ := range self.Storage {
		if filename == name {
			return true
		}
	}
	return false
}

func (self *Executable) LoadFile(filename string, data []byte) {
	checksum := sha256.Sum256(data)
	if !self.Exists(checksum[:]) {
		self.Storage[filename] = &File{
			Data: data,
		}
	}
}

// TODO: Use a byte correcting algorithm like reed solomon, raptor, fountain,
// lb, ...
func (self *Executable) Save() {
	if len(self.Storage) > 0 {
		self.HeadersOffset = self.HeadersOffset - uint64(3)
		self.StorageOffset = self.StorageOffset - uint64(3)
		self.Data = self.Data[:self.HeadersOffset]
		storedHeaders := []byte(HeadersMagicSequence)
		storedFiles := []byte(StorageMagicSequence)

		for filename, file := range self.Storage {
			self.StorageOffset += HeaderSize
			checksum := sha256.Sum256(file.Data)
			storedFile := &StoredFile{
				Filename: filename,
				Size:     uint64(len(file.Data)),
				Checksum: checksum[:],
				Offset:   (self.StorageOffset + uint64(len(storedFiles))),
			}

			storedHeaders = append(storedHeaders, storedFile.UnmarshalHeader()...)
			storedFiles = append(storedFiles, file.Data...)
		}

		binaryData := self.Data
		binaryData = append(binaryData, storedHeaders...)
		binaryData = append(binaryData, storedFiles...)

		os.Remove(self.Path)
		executableFile, _ := os.OpenFile(self.Path, os.O_CREATE|os.O_WRONLY, os.ModePerm)
		_, err := executableFile.Write(binaryData)
		if err != nil {
			fmt.Println("[error] failed to write executable file:", err)
		}
	}
}

func (self *Executable) load() {
	headersData := self.HeadersData()
	if len(headersData) > 0 {
		for i := 0; i < self.HeaderCount(); i++ {
			headerData := self.HeaderData(i)
			storedFile := MarshalHeader(headerData)
			self.StoredFiles = append(self.StoredFiles, storedFile)
			self.LoadFromBinary(storedFile)
		}
	}
}

func PadRight(s []byte, length int) []byte {
	for i := len(s); i < length; i++ {
		s = append(s, byte(0))
	}
	return s
}
