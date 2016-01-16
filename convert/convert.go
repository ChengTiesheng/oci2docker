package convert

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Sirupsen/logrus"
	"github.com/opencontainers/specs"
)

type DockerInfo struct {
	Appdir     string
	Entrypoint string
	Expose     string
}

const (
	buildTemplate = `
FROM scratch
MAINTAINER ChengTiesheng <chengtiesheng@huawei.com>
ENTRYPOINT ["{{.Entrypoint}}"]
ADD {{.Appdir}} .
EXPOSE {{.Expose}}
`
)

func RunOCI2Docker(path string, imgName string) error {
	appdir := "./rootfs"
	entrypoint := getEntrypointFromSpecs(path)

	dockerInfo := DockerInfo{
		Appdir:     appdir,
		Entrypoint: entrypoint,
		Expose:     "80",
	}

	generateDockerfile(dockerInfo)

	dirWork := createWorkDir()

	run(exec.Command("mv", "./Dockerfile", dirWork+"/Dockerfile"))
	run(exec.Command("cp", "-rf", path+"/rootfs", dirWork))

	run(exec.Command("docker", "build", "-t", imgName, dirWork))
	return nil
}

func generateDockerfile(dockerInfo DockerInfo) {
	t := template.Must(template.New("buildTemplate").Parse(buildTemplate))

	f, err := os.Create("Dockerfile")
	if err != nil {
		log.Fatal("Error wrinting Dockerfile %v", err.Error())
		return
	}
	defer f.Close()

	t.Execute(f, dockerInfo)

	fmt.Printf("Dockerfile generated, you can build the image with: \n")
	fmt.Printf("$ docker build -t %s .\n", dockerInfo.Entrypoint)

	return
}

// Create work directory for the conversion output
func createWorkDir() string {
	idir, err := ioutil.TempDir("", "oci2docker")
	if err != nil {
		return ""
	}
	rootfs := filepath.Join(idir, "rootfs")
	os.MkdirAll(rootfs, 0755)

	data := []byte{}
	if err := ioutil.WriteFile(filepath.Join(idir, "Dockerfile"), data, 0644); err != nil {
		return ""
	}
	return idir
}

func getEntrypointFromSpecs(path string) string {
	configPath := path + "/config.json"
	config, err := ioutil.ReadFile(configPath)
	if err != nil {
		logrus.Debugf("Open file config.json failed: %v", err)
		return ""
	}

	var spec specs.LinuxSpec
	err = json.Unmarshal(config, &spec)
	if err != nil {
		logrus.Debugf("Unmarshal config.json failed: %v", err)
		return ""
	}

	prefixDir := ""
	entryPoint := spec.Process.Args
	if entryPoint == nil {
		return "/bin/sh"
	}
	if !filepath.IsAbs(entryPoint[0]) {
		if spec.Process.Cwd == "" {
			prefixDir = "/"
		} else {
			prefixDir = spec.Process.Cwd
		}
	}
	entryPoint[0] = prefixDir + entryPoint[0]

	var res []string
	res = strings.SplitAfter(entryPoint[0], "/")
	if len(res) <= 2 {
		entryPoint[0] = "/bin" + entryPoint[0]
	}

	return entryPoint[0]
}
