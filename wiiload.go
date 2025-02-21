package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
)

const (
	WiiloadVersionMajor = 0
	WiiloadVersionMinor = 5
	FileChunkSize       = 1024 * 128
)

const usage = "usage: wiiload ip_address file_path"

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(1)
	}

	ipAddress := os.Args[1]
	filePath := os.Args[2]

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var buf bytes.Buffer
	writer := zlib.NewWriter(&buf)

	_, err = writer.Write(fileData)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = writer.Close()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var compressedFileData = buf.Bytes()

	fileStat, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fileSize := fileStat.Size()

	conn, err := net.Dial("tcp", ipAddress+":4299")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()

	conn.Write([]byte("HAXX"))
	conn.Write([]byte{WiiloadVersionMajor})
	conn.Write([]byte{WiiloadVersionMinor})
	binary.Write(conn, binary.BigEndian, uint16(len(filepath.Base(filePath))))
	binary.Write(conn, binary.BigEndian, uint32(len(compressedFileData)))
	binary.Write(conn, binary.BigEndian, uint32(fileSize))

	for i := 0; i < len(compressedFileData); i += FileChunkSize {
		end := i + FileChunkSize
		if end > len(compressedFileData) {
			end = len(compressedFileData)
		}
		conn.Write(compressedFileData[i:end])
	}

	conn.Write([]byte(filepath.Base(filePath)))
}
