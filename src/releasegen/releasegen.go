package releasegen

import "fmt"

type Release struct {
	ReleaseName string   `json:"releaseName"`
	Namespace   string   `json:"namespace"`
	ReleasePlan string   `json:"releasePlan"`
	Snapshot    string   `json:"snapshot"`
	Changes     []string `json:"changes"`
}

const template string = `---
apiVersion: appstudio.redhat.com/v1alpha1
kind: Release
metadata:
 name: {{ .ReleaseName }}
 namespace: {{ .Namespace }}
spec:
 releasePlan: {{ .ReleasePlan }}
 snapshot: {{ .Snapshot }}
 data:
		description: |
After a successful LVMS upgrade, the latest lvms-operator SHA image should be seen.

This erratum corrects the following changes:

{{ range .Changes }}
- {{ . }}
{{ end }}

Users of LVMS are advised to upgrade to the latest version in OpenShift Container Platform, which fixes these bugs and adds these enhancements.`

// GenerateReleaseYaml will take the release data as input and generate a release object that can be applied to the konflux cluster
func GenerateReleaseYaml(r Release) (string, error) {
	fmt.Println(template)

	return "", nil
}
