package main

import "github.com/satlayer/hello-world-bvs/reward_uploader/uploader"

func main() {
	up := uploader.NewUploader()
	up.Run()
}
