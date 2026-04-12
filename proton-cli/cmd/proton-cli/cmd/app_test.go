package cmd

import "testing"

func TestAppInstallCompletedMessage(t *testing.T) {
	got := appInstallCompletedMessage("release-manifests/0.5.0/kweaver-dip.yaml", "kweaver")
	want := "App install completed successfully: manifest=release-manifests/0.5.0/kweaver-dip.yaml namespace=kweaver"
	if got != want {
		t.Fatalf("appInstallCompletedMessage() = %q, want %q", got, want)
	}
}
