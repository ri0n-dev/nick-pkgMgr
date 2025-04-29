package main

import (
	"fmt"
	"nick/internal/fetcher"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("nick Developer by Ri0n \n 使い方: nick install <package-name>")
		return
	}

	command := os.Args[1]
	packageName := os.Args[2]

	if command == "install" {
		tarballURL := fetcher.GetTarballURL(packageName)
		fmt.Println("Tarball URL:", tarballURL)

		err := fetcher.DownloadAndExtract(tarballURL, packageName)
		if err != nil {
			fmt.Println("ERROR:", err)
		}
	} else {
		fmt.Println("Unknown Command:", command)
	}
}
