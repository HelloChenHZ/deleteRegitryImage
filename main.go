package main

import (
	"os"
	"fmt"
	"io/ioutil"
	"github.com/kubernetes/staging/src/k8s.io/apimachinery/pkg/util/json"
	"path/filepath"
		"os/exec"
	"strings"
	"flag"
)

type Blob struct {
	Config	Config	`json:"config"`
	Layers [] Layer `json:"layers"`
}

type Config struct {
	MediaType       string `json:"mediaType"`
	Size			int `json:"size"`
	Digest			string `json:"digest"`
}

type Layer struct{
	MediaType       string `json:"mediaType"`
	Size            int `json:"size"`
	Digest          string `json:"digest"`
}

var registryDir = flag.String("registryDir", "/gitlab-registry/docker/registry/v2/", "registry directory")
var keepNumber = flag.Int("keepNumber", 10, "the digest numbers keep in each tag")
var projectName = flag.String("projectName","infra/docker/codis/","project name")
var digestsMap, removeDigestsMap map[string]bool
var blobsFiles [10000] string
var bCount,pCount int

func pathExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("Dir "+path+" does not exist")
		return false
	}
	return true
}

func judgeArgs() bool {
	if !pathExist(*registryDir) {
		return false
	}

	if !pathExist(*registryDir+"repositories/"+*projectName) {
		return false
	}

	if *keepNumber > 50 || *keepNumber < 1 {
		fmt.Println("Invalid keepNumber value. The value must be 1-50")
	}
	return true
}

func findDigest(dirName string) bool{
	_, ok := digestsMap[dirName]
	return ok
}

func getLayerDigests(blobFile string) {
	jsonFile, err := os.Open(blobFile)
	if err != nil {
		fmt.Println(err)
	}

	//fmt.Println("Successfully open blobsfile")
	defer jsonFile.Close()

	var blob Blob
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &blob)

	_, ok := digestsMap[blob.Config.Digest[7:]]
	if ok {
		fmt.Println("Digests "+blob.Config.Digest[7:]+"is exists and ignore it!")
	}else{
		digestsMap[blob.Config.Digest[7:]] = true
		fmt.Println("Add Digests "+blob.Config.Digest[7:])
	}

	for i:=0; i<len(blob.Layers); i++{
		_, ok := digestsMap[blob.Layers[i].Digest[7:]]
		if ok {
			fmt.Println("Digests "+blob.Layers[i].Digest[7:]+"is exists and ignore it!")
		}else{
			digestsMap[blob.Layers[i].Digest[7:]] = true
			fmt.Println("Add Digests "+blob.Layers[i].Digest[7:])
		}
	}
}

func main(){
	flag.Parse()
	if !judgeArgs() {
		return
	}

	pCount=1
	for i:=0;i<pCount;i++{
		digestsMap = map[string]bool{}
		removeDigestsMap = map[string]bool{}
		tags, err :=ioutil.ReadDir(*registryDir+"repositories/"+*projectName+"_manifests/tags/")
		if err!=nil {
			fmt.Println("Get tags from "+*registryDir+"repositories/"+*projectName+"_manifests/tags/ failed!")
			continue
		}

		for _, tag:=range tags{
			cmd := exec.Command("ls", "-th", *registryDir+"repositories/"+*projectName+"_manifests/tags/"+tag.Name()+"/index/sha256/")
			out, err := cmd.CombinedOutput()
			if err!= nil{
				fmt.Println("Get digest form "+*registryDir+"repositories/"+*projectName+"_manifests/tags/"+tag.Name()+"/index/sha256/ failed!")
				continue
			}
			tagDigests := strings.Fields(string(out))

			//fmt.Printf("current tag is %v and tagDigests is %v\n", tag.Name(), string(out))
			var digestCount = 0
			for _, tagDigest := range tagDigests{
				digestPath := *registryDir+"blobs/sha256/" + tagDigest[:2]+"/"+tagDigest+"/data"
				if !pathExist(digestPath){
					os.RemoveAll( *registryDir+"repositories/"+*projectName+"_manifests/tags/"+tag.Name()+"/index/sha256/"+tagDigest)
					continue
				}

				//fmt.Println("File "+digestPath+" exists")
				digestCount++
				if digestCount <= *keepNumber{
					blobsFiles[bCount] = digestPath
					bCount++
				}else{
					os.RemoveAll( *registryDir+"repositories/"+*projectName+"_manifests/tags/"+tag.Name()+"/index/sha256/"+tagDigest)
				}
			}
		}
	}

	for i:=0;i<bCount;i++ {
		fmt.Println("blobsFiles is "+ blobsFiles[i])
	}

	for i:=0;i<bCount;i++{
		getLayerDigests(blobsFiles[i])
	}

	dir := *registryDir+"repositories/"+*projectName+"_layers/sha256/"
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error{
		if err != nil {
			return err
		}

		if info.IsDir() {
			fmt.Printf("dir is %v\n", info.Name())
			if info.Name() != "sha256" {
				if !findDigest(info.Name()) {
					_, ok := removeDigestsMap[info.Name()]
					if ok {
						fmt.Println("Exist digest " + info.Name() + " in removeDigestsMap")
					} else {
						fmt.Println("Push digest " + info.Name() + " to removeDigestMap")
						removeDigestsMap[info.Name()] = true
					}
				}
				return filepath.SkipDir
			}
		}

		fmt.Printf("visited file: %q\n", path)
		return nil
	})

	for digest := range removeDigestsMap{
		fmt.Printf("os.Remove %v %v \n", dir, digest)
		//os.RemoveAll(dir+digest)
	}
}

