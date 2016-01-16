package main

import (
	"os"

	"github.com/ChengTiesheng/oci2docker/convert"
	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

func main() {
	logrus.SetLevel(logrus.InfoLevel)
	app := cli.NewApp()
	app.Name = "oci2docker"
	app.Usage = "A tool for coverting oci bundle to docker image"
	app.Version = "0.1.0"
	app.Commands = []cli.Command{
		{
			Name:  "convert",
			Usage: "convert operation",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "oci-bundle",
					Value: "oci-bundle",
					Usage: "path of oci-bundle to convert",
				},
				cli.StringFlag{
					Name:  "image-name",
					Value: "image-name",
					Usage: "docker image name",
				},
			},
			Action: oci2docker,
		},
	}

	app.Run(os.Args)
}

func oci2docker(c *cli.Context) {
	ociPath := c.String("oci-bundle")
	imgName := c.String("image-name")
	if ociPath == "" {
		cli.ShowCommandHelp(c, "convert")
		return
	}

	if imgName == "" {
		cli.ShowCommandHelp(c, "convert")
		return
	}
	_, err := os.Stat(ociPath)
	if os.IsNotExist(err) {
		cli.ShowCommandHelp(c, "convert")
		return
	}

	convert.RunOCI2Docker(ociPath, imgName)
}
