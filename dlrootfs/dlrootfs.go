package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/rmonjo/dlrootfs"
)

const VERSION string = "1.4.0"

var (
	rootfsDest    *string = flag.String("d", "./rootfs", "destination of the resulting rootfs directory")
	imageFullName *string = flag.String("i", "", "name of the image <repository>/<image>:<tag>")
	credentials   *string = flag.String("u", "", "docker hub credentials: <username>:<password>")
	gitLayering   *bool   = flag.Bool("g", false, "use git layering")
	version       *bool   = flag.Bool("v", false, "display dlrootfs version")
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: dlrootfs -i <image_name>:[<image_tag>] [-d <rootfs_destination>] [-u <username>:<password>]\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  dlrootfs -i ubuntu  #if no tag, use latest\n")
		fmt.Fprintf(os.Stderr, "  dlrootfs -i ubuntu:precise -d ubuntu_rootfs\n")
		fmt.Fprintf(os.Stderr, "  dlrootfs -i dockefile/elasticsearch:latest\n")
		fmt.Fprintf(os.Stderr, "  dlrootfs -i my_repo/my_image:latest -u username:password\n")
		fmt.Fprintf(os.Stderr, "Default:\n")
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	if *version {
		fmt.Println(VERSION)
		return
	}

	if *imageFullName == "" {
		flag.Usage()
		return
	}

	fmt.Printf("Retrieving %v info from the DockerHub ...\n", *imageFullName)
	pullContext, err := dlrootfs.RequestPullContext(*imageFullName, *credentials)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Image ID: %v\n", pullContext.ImageId)

	err = dlrootfs.DownloadImage(pullContext, *rootfsDest, *gitLayering, true)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nRootfs of %v:%v in %v\n", pullContext.ImageName, pullContext.ImageTag, *rootfsDest)
	if *credentials != "" {
		fmt.Printf("WARNING: don't forget to remove your docker hub credentials from your history !!\n")
	}
}
