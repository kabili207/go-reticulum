package main

import (
	"os"
	"testing"
)

func TestIntegration_Filetransfer_Flow(t *testing.T) {
	if os.Getenv("RNS_INTEGRATION") == "" {
		t.Skip("set RNS_INTEGRATION=1 to run integration tests")
	}
	TestListFilesFiltersAndSorts(t)
	TestPackFileListChunksRespectsLimit(t)
	TestFileExistsInDirRejectsTraversal(t)
	TestParseTruncatedHashHexLength(t)
}
