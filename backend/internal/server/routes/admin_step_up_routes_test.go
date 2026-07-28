package routes

import (
	"os"
	"strings"
	"testing"
)

func TestBackupCredentialUpdatesRequireStepUp(t *testing.T) {
	source, err := os.ReadFile("admin.go")
	if err != nil {
		t.Fatal(err)
	}

	text := string(source)
	for _, route := range []string{
		`backup.PUT("/s3-config", gin.HandlerFunc(stepUpAuth), h.Admin.Backup.UpdateS3Config)`,
		`backup.PUT("/image-storage", gin.HandlerFunc(stepUpAuth), h.Admin.Backup.UpdateImageStorageConfig)`,
	} {
		if !strings.Contains(text, route) {
			t.Fatalf("backup credential update route must require step-up authentication: %s", route)
		}
	}
}
