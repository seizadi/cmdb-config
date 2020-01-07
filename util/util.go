package util

import (
	"os"
	"gopkg.in/yaml.v2"
)

func PutText(txt string, DestFilename string) error {
	
	out := []byte(txt)
	
	err := CopyBufferContents(out, DestFilename)
	
	if err != nil {
		return err
	}
	
	return nil
}

func PutMap(addr interface{}, DestFilename string) error {
	
	out, err := yaml.Marshal(addr)
	
	if err != nil {
		return err
	}
	
	err = CopyBufferContents(out, DestFilename)
	
	if err != nil {
		return err
	}
	
	return nil
}

// copyBufferContents copies the contents of the buffer named srcBuff to the file
// named by destFile. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source buffer.
func CopyBufferContents(srcBuff []byte, destFile string) (err error) {
	
	out, err := os.Create(destFile)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = out.Write(srcBuff); err != nil {
		return
	}
	err = out.Sync()
	return
}